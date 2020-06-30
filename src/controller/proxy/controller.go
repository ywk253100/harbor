//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"context"
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/registry"
	"github.com/opencontainers/go-digest"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	maxWait             = 10
	manifestListWaitSec = 900
	sleepIntervalSec    = 20
)

var (
	// Ctl is a global proxy controller instance
	ctl     Controller
	ctlLock sync.Mutex
)

// Controller defines the operations related with pull through proxy
type Controller interface {
	// UseLocal check if the manifest should use local
	UseLocal(ctx context.Context, p *models.Project, art lib.ArtifactInfo) bool
	// ProxyBlob proxy the blob request to the target server
	ProxyBlob(ctx context.Context, p *models.Project, repo string, dig string, w http.ResponseWriter, r RemoteInterface) error
	// ProxyManifest proxy the manifest to the target server
	ProxyManifest(ctx context.Context, p *models.Project, repo string, art lib.ArtifactInfo, w http.ResponseWriter, r RemoteInterface) error
}
type controller struct {
	blobCtl     blob.Controller
	registryMgr registry.Manager
	artifactCtl artifact.Controller
	local       LocalInterface
}

func ControllerInstance() Controller {
	ctlLock.Lock()
	defer ctlLock.Unlock()
	if ctl == nil {
		ctl = &controller{
			blobCtl:     blob.Ctl,
			registryMgr: registry.NewDefaultManager(),
			artifactCtl: artifact.Ctl,
			local:       CreateLocalInterface(),
		}
	}
	return ctl
}

func (c *controller) isProxyReady(p *models.Project) bool {
	if p.RegistryID < 1 {
		return false
	}
	reg, err := c.registryMgr.Get(p.RegistryID)
	if err != nil {
		log.Errorf("failed to get registry, error:%v", err)
		return false
	}
	return reg.Status == model.Healthy
}

func (c *controller) UseLocal(ctx context.Context, p *models.Project, art lib.ArtifactInfo) bool {
	if !c.isProxyReady(p) {
		return true
	}
	if len(string(art.Digest)) > 0 {
		exist, err := c.local.BlobExist(ctx, art.Digest)
		if err == nil && exist {
			return true
		}
	}
	return false
}

func (c *controller) ProxyManifest(ctx context.Context, p *models.Project, repo string, art lib.ArtifactInfo, w http.ResponseWriter, r RemoteInterface) error {
	var man distribution.Manifest
	var err error
	ref := string(art.Digest)
	if len(ref) == 0 {
		ref = art.Tag
	}
	if len(ref) == 0 {
		ref = "latest"
	}
	man, err = r.Manifest(repo, ref)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			go func() {
				c.local.DeleteManifest(ctx, repo, string(art.Tag))
			}()
		}
		return err
	}

	ct, payload, err := man.Payload()
	setHeaders(w, int64(len(payload)), ct, art.Digest)
	w.Write(payload)

	// Push manifest in background
	go func() {
		c.waitAndPushManifest(ctx, p, repo, string(art.Tag), man, art, ct, r)
	}()

	return nil
}

func (c *controller) ProxyBlob(ctx context.Context, p *models.Project, repo string, dig string, w http.ResponseWriter, r RemoteInterface) error {
	log.Debugf("The blob doesn't exist, proxy the request to the target server, url:%v", repo)
	desc := distribution.Descriptor{}
	size, bReader, err := r.BlobReader(repo, dig)
	if err != nil {
		log.Errorf("failed to pull blob, error %v", err)
		return err
	}
	defer bReader.Close()
	written, err := io.CopyN(w, bReader, size)
	if written != size {
		e := errors.Errorf("The size mismatch, actual:%d, expected: %d", written, size)
		return e
	}
	desc.Size = size
	desc.Digest = digest.Digest(dig)

	setHeaders(w, size, desc.MediaType, dig)
	go func() {
		err := c.putBlobToLocal(ctx, p, repo, p.Name+"/"+repo, desc, r)
		if err != nil {
			log.Errorf("error while putting blob to local, %v", err)
		}
	}()
	return nil
}

func (c *controller) putBlobToLocal(ctx context.Context, p *models.Project, orgRepo string, localRepo string, desc distribution.Descriptor, r RemoteInterface) error {
	log.Debugf("Put blob to local registry!, sourceRepo:%v, localRepo:%v, digest: %v", orgRepo, localRepo, desc.Digest)
	_, bReader, err := r.BlobReader(orgRepo, string(desc.Digest))
	if err != nil {
		log.Error(err)
		return err
	}
	defer bReader.Close()
	err = c.local.PushBlob(ctx, p, localRepo, desc, bReader)
	return err
}

func setHeaders(w http.ResponseWriter, size int64, mediaType string, dig string) {
	w.Header().Set("Content-Length", fmt.Sprintf("%v", size))
	if len(mediaType) > 0 {
		w.Header().Set("Content-Type", mediaType)
	}
	w.Header().Set("Docker-Content-Digest", dig)
	w.Header().Set("Etag", dig)
}

func (c *controller) waitAndPushManifest(ctx context.Context, p *models.Project, repo, tag string, man distribution.Manifest, art lib.ArtifactInfo, contType string, r RemoteInterface) {
	localRepo := art.ProjectName + "/" + repo
	if contType == manifestlist.MediaTypeManifestList {
		err := c.local.PushManifestList(ctx, p, localRepo, tag, art, man)
		if err != nil {
			log.Errorf("error when push manifest list to local:%v", err)
		}
		return
	}

	for n := 0; n < maxWait; n = n + 1 {
		time.Sleep(sleepIntervalSec * time.Second)
		waitBlobs := c.local.CheckDependencies(ctx, man, string(art.Digest), contType)
		if len(waitBlobs) == 0 {
			break
		}
		log.Debugf("Current n=%v", n)
		if n+1 == maxWait && len(waitBlobs) > 0 {
			log.Debug("Waiting blobs not empty, push it to local repo manually")
			for _, desc := range waitBlobs {
				c.putBlobToLocal(ctx, p, repo, localRepo, desc, r)
			}
		}
	}
	err := c.local.PushManifest(ctx, p, localRepo, art.Tag, man)
	if err != nil {
		log.Errorf("failed to push manifest, error %v", err)
	}
}
