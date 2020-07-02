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
	"github.com/goharbor/harbor/src/replication/registry"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
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
	// UseLocal check if the manifest should use localHelper
	UseLocal(ctx context.Context, art lib.ArtifactInfo) bool
	// ProxyBlob proxy the blob request to the target server
	ProxyBlob(ctx context.Context, p *models.Project, art lib.ArtifactInfo, w http.ResponseWriter) error
	// ProxyManifest proxy the manifest to the target server
	ProxyManifest(ctx context.Context, p *models.Project, art lib.ArtifactInfo, w http.ResponseWriter) error
}
type controller struct {
	blobCtl     blob.Controller
	registryMgr registry.Manager
	artifactCtl artifact.Controller
	local       localInterface
}

// ControllerInstance -- Get the proxy controller instance
func ControllerInstance() (Controller, error) {
	// Lazy load the controller
	// Because LocalHelper is not ready unless core startup completely
	ctlLock.Lock()
	defer ctlLock.Unlock()
	if ctl == nil {
		helper, err := newLocalHelper()
		if err != nil {
			return nil, err
		}
		ctl = &controller{
			blobCtl:     blob.Ctl,
			registryMgr: registry.NewDefaultManager(),
			artifactCtl: artifact.Ctl,
			local:       helper,
		}
	}
	return ctl, nil
}

func (c *controller) UseLocal(ctx context.Context, art lib.ArtifactInfo) bool {
	if len(art.Digest) > 0 {
		exist, err := c.local.BlobExist(ctx, art.Repository, art.Digest)
		if err == nil && exist {
			return true
		}
	}
	return false
}

func (c *controller) ProxyManifest(ctx context.Context, p *models.Project, art lib.ArtifactInfo, w http.ResponseWriter) error {
	var man distribution.Manifest
	var err error
	remoteRepo := art.ProxyCacheRemoteRepo()
	r := newRemoteHelper(p.RegistryID)
	ref := art.Digest
	if len(ref) == 0 {
		ref = art.Tag
	}
	man, err = r.Manifest(remoteRepo, ref)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			go func() {
				c.local.DeleteManifest(ctx, remoteRepo, art.Tag)
			}()
		}
		return err
	}

	ct, payload, err := man.Payload()
	if err != nil {
		return err
	}
	setHeaders(w, int64(len(payload)), ct, art.Digest)
	_, err = w.Write(payload)
	if err != nil {
		return err
	}

	// Push manifest in background
	go func() {
		c.waitAndPushManifest(ctx, remoteRepo, man, art, ct, r)
	}()

	return nil
}

func (c *controller) ProxyBlob(ctx context.Context, p *models.Project, art lib.ArtifactInfo, w http.ResponseWriter) error {
	remoteRepo := art.ProxyCacheRemoteRepo()
	log.Debugf("The blob doesn't exist, proxy the request to the target server, url:%v", remoteRepo)
	r := newRemoteHelper(p.RegistryID)
	desc := distribution.Descriptor{}
	size, bReader, err := r.BlobReader(remoteRepo, art.Digest)
	if err != nil {
		log.Errorf("failed to pull blob, error %v", err)
		return err
	}
	defer bReader.Close()
	// Use io.CopyN to avoid avoid out of memory when pulling big blob
	written, err := io.CopyN(w, bReader, size)
	if err != nil {
		log.Errorf("failed to proxy the digest: %v, error %v", art.Digest, err)
		return err
	}
	if written != size {
		return errors.Errorf("The size mismatch, actual:%d, expected: %d", written, size)
	}
	desc.Size = size
	desc.Digest = digest.Digest(art.Digest)

	setHeaders(w, size, desc.MediaType, art.Digest)
	// put blob to localHelper will start another connection to the remoteHelper,
	// to reduce the impact of it, cache the blob after it send to the client
	go func() {
		err := c.putBlobToLocal(ctx, remoteRepo, art.Repository, desc, r)
		if err != nil {
			log.Errorf("error while putting blob to localHelper, %v", err)
		}
	}()
	return nil
}

func (c *controller) putBlobToLocal(ctx context.Context, remoteRepo string, localRepo string, desc distribution.Descriptor, r remoteInterface) error {
	log.Debugf("Put blob to localHelper registry!, sourceRepo:%v, localRepo:%v, digest: %v", remoteRepo, localRepo, desc.Digest)
	_, bReader, err := r.BlobReader(remoteRepo, string(desc.Digest))
	if err != nil {
		log.Error(err)
		return err
	}
	defer bReader.Close()
	err = c.local.PushBlob(ctx, localRepo, desc, bReader)
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

func (c *controller) waitAndPushManifest(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, contType string, r remoteInterface) {

	if contType == manifestlist.MediaTypeManifestList || contType == v1.MediaTypeImageIndex {
		err := c.local.PushManifestList(ctx, art.Repository, art.Tag, man)
		if err != nil {
			log.Errorf("error when push manifest list to localHelper:%v", err)
		}
		return
	}

	for n := 0; n < maxWait; n = n + 1 {
		time.Sleep(sleepIntervalSec * time.Second)
		waitBlobs := c.local.CheckDependencies(ctx, art.Repository, man)
		if len(waitBlobs) == 0 {
			break
		}
		log.Debugf("Current n=%v artifact: %v:%v", n, art.Repository, art.Tag)
		if n+1 == maxWait && len(waitBlobs) > 0 {
			// docker client will skip to pull layers exist in localHelper
			// these blobs is not exist in the proxy server
			// it will cause the manifest dependency check always fail
			// need to push these blobs before push manifest to avoid failure
			log.Debug("Waiting blobs not empty, push it to localHelper remoteRepo directly")
			for _, desc := range waitBlobs {
				err := c.putBlobToLocal(ctx, remoteRepo, art.Repository, desc, r)
				if err != nil {
					log.Errorf("Failed to push blob to cache error: %v", err)
					return
				}
			}
		}
	}
	err := c.local.PushManifest(ctx, art.Repository, art.Tag, man)
	if err != nil {
		log.Errorf("failed to push manifest, error %v", err)
	}
}
