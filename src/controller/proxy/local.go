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
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/harbor/base"
	"github.com/goharbor/harbor/src/replication/adapter/harbor/v2"
	"github.com/goharbor/harbor/src/replication/model"
	"io"
	"time"
)

// localInterface defines operations related to localHelper repo under proxy mode
type localInterface interface {
	// BlobExist check if the blob exist in localHelper repo
	BlobExist(ctx context.Context, dig string) (bool, error)
	// PushBlob push blob to localHelper repo
	PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error
	// PushManifest push manifest to localHelper repo
	PushManifest(repo string, tag string, manifest distribution.Manifest) error
	// PushManifestList push manifest list to localHelper repo
	PushManifestList(ctx context.Context, repo string, tag string, man distribution.Manifest) error
	// CheckDependencies check if the manifest's dependency is ready
	CheckDependencies(ctx context.Context, man distribution.Manifest) []distribution.Descriptor
	// DeleteManifest cleanup delete tag from localHelper cache
	DeleteManifest(repo, ref string)
}

// localHelper defines operations related to localHelper repo under proxy mode
type localHelper struct {
	registry adapter.ArtifactRegistry
}

// newLocalHelper create the localInterface
func newLocalHelper() (localInterface, error) {
	l := &localHelper{}
	if err := l.init(); err != nil {
		log.Errorf("Failed to init localHelper, error %v", err)
		return nil, fmt.Errorf("failed to initialize the local helper")
	}
	return l, nil
}

func (l *localHelper) BlobExist(ctx context.Context, dig string) (bool, error) {
	// not using l.registry.BlobExist(repo, dig)
	// sometimes the blob exist in storage,
	// but it does not exist in the Harbor database.
	// if push manifest to Harbor, it still get failed
	return blob.Ctl.Exist(ctx, dig)
}

func (l *localHelper) init() error {
	if l.registry != nil {
		return nil
	}
	log.Debugf("core url:%s, localHelper core url: %v", config.GetCoreURL(), config.LocalCoreURL())
	// the traffic is internal only
	registryURL := config.LocalCoreURL()
	// TODO: need to verify it works
	// registryURL := config.GetCoreURL()
	reg := &model.Registry{
		URL: registryURL,
		Credential: &model.Credential{
			Type:         model.CredentialTypeSecret,
			AccessSecret: config.ProxyServiceSecret,
		},
		Insecure: true,
	}
	baseAdapter, err := base.New(reg)
	if err != nil {
		return err
	}
	adp := v2.New(baseAdapter)
	l.registry = adp.(adapter.ArtifactRegistry)
	return err
}

func (l *localHelper) PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error {
	log.Debugf("Put blob to localHelper registry, localRepo:%v, digest: %v", localRepo, desc.Digest)
	err := l.registry.PushBlob(localRepo, string(desc.Digest), desc.Size, bReader)
	return err
}

func (l *localHelper) PushManifest(repo string, tag string, manifest distribution.Manifest) error {
	// Make sure there is only one go routing to push current artifact to localHelper repo
	if len(tag) == 0 {
		// when push a manifest list, the tag is empty, for example: busybox
		// if tag is empty, set to latest
		tag = "latest"
	}
	artifact := repo + ":" + tag
	if !inflightChecker.addRequest(artifact) {
		return nil
	}
	defer inflightChecker.removeRequest(artifact)

	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return err
	}
	_, err = l.registry.PushManifest(repo, tag, mediaType, payload)
	return err
}

// DeleteManifest cleanup delete tag from localHelper cache
func (l *localHelper) DeleteManifest(repo, ref string) {
	log.Debug("Remove tag from repo if it is exist")
	if err := l.registry.DeleteManifest(repo, ref); err != nil {
		//sometimes user pull a non-exist image
		log.Debugf("failed to remove artifact, error %v", err)
	}
}

// updateManifestList -- Trim the manifest list, make sure all depend manifests are ready
func (l *localHelper) updateManifestList(ctx context.Context, manifest distribution.Manifest) (distribution.Manifest, error) {
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

func (l *localHelper) PushManifestList(ctx context.Context, repo string, tag string, man distribution.Manifest) error {
	// For manifest list, it might include some different platforms, such as amd64, arm
	// the client only pull one platform, such as amd64, the arm platform is not pulled.
	// if pushing the original directly, it will fail to check the dependencies
	// to avoid this error, need to update the manifest list, keep the existing platform
	// the proxy wait 30 minutes, and push the updated manifest list in cache
	time.Sleep(manifestListWaitSec * time.Second)
	newMan, err := l.updateManifestList(ctx, man)
	if err != nil {
		log.Error(err)
	}
	return l.PushManifest(repo, tag, newMan)
}

func (l *localHelper) CheckDependencies(ctx context.Context, man distribution.Manifest) []distribution.Descriptor {
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
