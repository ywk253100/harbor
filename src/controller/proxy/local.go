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
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/adapter/harbor/base"
	"github.com/goharbor/harbor/src/replication/model"
	"io"
	"sync"
	"time"
)

var (
	mu       sync.Mutex
	inflight map[string]interface{} = make(map[string]interface{})
)

// LocalInterface defines operations related to local repo under proxy mode
type LocalInterface interface {
	// BlobExist check if the blob exist in local repo
	BlobExist(ctx context.Context, dig string) (bool, error)
	// PushBlob push blob to local repo
	PushBlob(ctx context.Context, p *models.Project, localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error
	// PushManifest push manifest to local repo
	PushManifest(ctx context.Context, p *models.Project, repo string, tag string, mfst distribution.Manifest) error
	// PushManifestList push manifest list to local repo
	PushManifestList(ctx context.Context, p *models.Project, repo string, tag string, art lib.ArtifactInfo, man distribution.Manifest) error
	// CheckDependencies check if the manifest's dependency is ready
	CheckDependencies(ctx context.Context, man distribution.Manifest, dig string, mediaType string) []distribution.Descriptor
	// CleanupTag cleanup delete tag from local cache
	CleanupTag(ctx context.Context, repo, tag string)
}

// local defines operations related to local repo under proxy mode
type local struct {
	adapter *base.Adapter
}

// CreateLocalInterface create the LocalInterface
func CreateLocalInterface() LocalInterface {
	return &local{}
}

// TODO: replace it with head request to local repo
func (l *local) BlobExist(ctx context.Context, dig string) (bool, error) {
	return blob.Ctl.Exist(ctx, dig)
}

func (l *local) init() error {
	if l.adapter != nil {
		return nil
	}
	registryURL := config.GetCoreURL()
	reg := &model.Registry{
		URL: registryURL,
		Credential: &model.Credential{
			Type:         model.CredentialTypeSecret,
			AccessSecret: config.ProxyServiceSecret,
		},
	}
	adapter, err := base.New(reg)
	l.adapter = adapter
	return err
}

func (l *local) PushBlob(ctx context.Context, p *models.Project, localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error {
	log.Debugf("Put blob to local registry, localRepo:%v, digest: %v", localRepo, desc.Digest)
	if err := l.init(); err != nil {
		return err
	}
	err := l.adapter.PushBlob(localRepo, string(desc.Digest), desc.Size, bReader)
	return err
}

func (l *local) PushManifest(ctx context.Context, p *models.Project, repo string, tag string, mfst distribution.Manifest) error {
	// Make sure there is only one go routing to push current artifact to local repo
	artifact := repo + ":" + tag
	mu.Lock()
	_, ok := inflight[artifact]
	if ok {
		mu.Unlock()
		// Skip to copy artifact if there is existing job running
		return nil
	}
	inflight[artifact] = 1
	mu.Unlock()
	defer releaseLock(artifact)

	if err := l.init(); err != nil {
		return err
	}
	mediaType, payload, err := mfst.Payload()
	if err != nil {
		return err
	}
	_, err = l.adapter.PushManifest(repo, tag, mediaType, payload)
	return err
}

// CleanupTag cleanup delete tag from local cache
func (l *local) CleanupTag(ctx context.Context, repo, tag string) {
	log.Debug("Remove tag from repo if it is exist")
	// TODO: remove cached tag if it exist in cache
}

func releaseLock(artifact string) {
	mu.Lock()
	delete(inflight, artifact)
	mu.Unlock()
}

// updateManifestList -- Trim the manifest list, make sure all depend manifests are ready
func (l *local) updateManifestList(ctx context.Context, manifest distribution.Manifest) (distribution.Manifest, error) {
	switch v := manifest.(type) {
	case *manifestlist.DeserializedManifestList:
		existMans := make([]manifestlist.ManifestDescriptor, 0)
		for _, m := range v.Manifests {
			exist, err := l.BlobExist(ctx, string(m.Digest))
			if err != nil {
				continue
			}
			if exist {
				existMans = append(existMans, m)
			}
		}
		if len(existMans) > 0 {
			// Avoid empty manifest in the manifest list
			return manifestlist.FromDescriptors(existMans)
		}
	}
	return manifest, nil
}

func (l *local) PushManifestList(ctx context.Context, p *models.Project, repo string, tag string, art lib.ArtifactInfo, man distribution.Manifest) error {
	// Make sure all depend manifests are pushed to local repo
	time.Sleep(manifestListWaitSec * time.Second)
	newMan, err := l.updateManifestList(ctx, man)
	if err != nil {
		log.Error(err)
	}
	return l.PushManifest(ctx, p, repo, tag, newMan)
}

func (l *local) CheckDependencies(ctx context.Context, man distribution.Manifest, dig string, mediaType string) []distribution.Descriptor {
	descriptors := man.References()
	waitDesc := make([]distribution.Descriptor, 0)
	for _, desc := range descriptors {
		log.Debugf("checking the blob dependency: %v", desc.Digest)
		exist, err := l.BlobExist(ctx, string(desc.Digest))
		if err != nil || !exist {
			log.Debugf("Check dependency failed!")
			waitDesc = append(waitDesc, desc)
		}
	}
	log.Debugf("Check dependency result %v", waitDesc)
	return waitDesc
}
