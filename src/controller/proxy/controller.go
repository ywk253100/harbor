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
	serror "github.com/goharbor/harbor/src/server/error"
	"net/http"
	"sync"
	"time"
)

const (
	maxWait             = 10
	manifestListWaitSec = 1800
	sleepIntervalSec    = 20
)

var (
	// Ctl is a global proxy controller instance
	Ctl = NewController()
)

type Controller interface {
	// UseLocalManifest check if the manifest should use local
	UseLocalManifest(ctx context.Context, p *models.Project, art lib.ArtifactInfo) bool
	// UseLocalBlob check if the blob should use local
	UseLocalBlob(ctx context.Context, p *models.Project, digest string) bool
	// ProxyBlob proxy the blob request to the target server
	ProxyBlob(ctx context.Context, p *models.Project, repo string, dig string, w http.ResponseWriter) error
	// ProxyManifest proxy the manifest to the target server
	ProxyManifest(ctx context.Context, p *models.Project, repo string, art lib.ArtifactInfo, w http.ResponseWriter) error
}

type controller struct {
	blobCtl     blob.Controller
	registryMgr registry.Manager
	mu          sync.Mutex
	inflight    map[string]interface{}
	artifactCtl artifact.Controller
	local       *Local
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

func (c *controller) UseLocalManifest(ctx context.Context, p *models.Project, art lib.ArtifactInfo) bool {
	if !c.isProxyReady(p) {
		return true
	}
	if len(string(art.Digest)) > 0 {
		exist, err := c.local.blobExist(ctx, art.Digest)
		if err == nil && exist {
			return true
		}
	}
	return false
}

func (c *controller) UseLocalBlob(ctx context.Context, p *models.Project, digest string) bool {
	if !c.isProxyReady(p) {
		return true
	}
	exist, err := c.local.blobExist(ctx, digest)
	if err != nil {
		return false
	}
	return exist
}

func (c *controller) ProxyManifest(ctx context.Context, p *models.Project, repo string, art lib.ArtifactInfo, w http.ResponseWriter) error {
	var man distribution.Manifest
	var err error
	r := &Remote{
		regID: p.RegistryID,
	}
	l := &Local{}
	if len(string(art.Digest)) > 0 {
		// pull by digest
		log.Debugf("Getting manifest by digiest %v", art.Digest)
		man, err = r.ManifestByDigest(repo, string(art.Digest))
	} else { // pull by tag
		if len(art.Tag) == 0 {
			art.Tag = "latest"
		}
		log.Debugf("Getting manifest by tag %v", art.Tag)
		man, err = r.ManifestByTag(repo, string(art.Tag))
	}

	if err != nil {
		if errors.IsNotFoundErr(err) && len(art.Tag) > 0 {
			go func() {
				l.cleanupTagInLocal(ctx, repo, string(art.Tag))
			}()
		}
		serror.SendError(w, err)
		return err
	}

	ct, payload, err := man.Payload()
	setHeaders(w, int64(len(payload)), ct, art.Digest)
	w.Write(payload)

	// Push manifest in background
	go func() {
		c.waitAndPushManifest(ctx, p, repo, string(art.Tag), man, art, ct)
	}()

	return nil
}

func (c *controller) ProxyBlob(ctx context.Context, p *models.Project, repo string, dig string, w http.ResponseWriter) error {
	log.Debugf("The blob doesn't exist, proxy the request to the target server, url:%v", repo)
	r := &Remote{
		regID: p.RegistryID,
	}
	desc, err := r.Blob(w, repo, dig)
	if err != nil {
		log.Error(err)
		serror.SendError(w, err)
		return err
	}
	setHeaders(w, desc.Size, desc.MediaType, string(desc.Digest))
	go func() {
		err := c.putBlobToLocal(ctx, p, repo, p.Name+"/"+repo, desc)
		if err != nil {
			log.Errorf("error while putting blob to local, %v", err)
		}
	}()
	return nil
}

// NewController create an instance of the Controller
func NewController() Controller {
	return &controller{
		blobCtl:     blob.Ctl,
		registryMgr: registry.NewDefaultManager(),
		inflight:    make(map[string]interface{}),
		artifactCtl: artifact.Ctl,
		local:       &Local{},
	}
}

func (c *controller) putBlobToLocal(ctx context.Context, p *models.Project, orgRepo string, localRepo string, desc distribution.Descriptor) error {
	log.Debugf("Put blob to local registry!, sourceRepo:%v, localRepo:%v, digest: %v", orgRepo, localRepo, desc.Digest)
	r := &Remote{regID: p.RegistryID}
	_, bReader, err := r.BlobReader(orgRepo, string(desc.Digest))
	defer bReader.Close()
	if err != nil {
		log.Error(err)
		return err
	}
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

func (c *controller) waitAndPushManifest(ctx context.Context, p *models.Project, repo, tag string, man distribution.Manifest, art lib.ArtifactInfo, contType string) {
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
				c.putBlobToLocal(ctx, p, repo, localRepo, desc)
			}
		}
	}
	err := c.local.PushManifest(ctx, p, localRepo, tag, man)
	if err != nil {
		log.Errorf("failed to push manifest, error %v", err)
	}
}
