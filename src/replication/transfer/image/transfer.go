// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"errors"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	common_http "github.com/goharbor/harbor/src/common/http"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"net/http"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	trans "github.com/goharbor/harbor/src/replication/transfer"
)

func init() {
	if err := trans.RegisterFactory(model.ResourceTypeImage, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
	if err := trans.RegisterFactory(model.ResourceTypeArtifact, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
}

type repository struct {
	repository string
	tags       []string
}

func factory(logger trans.Logger, stopFunc trans.StopFunc) (trans.Transfer, error) {
	return &transfer{
		logger:    logger,
		isStopped: stopFunc,
	}, nil
}

type transfer struct {
	logger    trans.Logger
	isStopped trans.StopFunc
	src       adapter.ArtifactRegistry
	dst       adapter.ArtifactRegistry
}

func (t *transfer) Transfer(src *model.Resource, dst *model.Resource) error {
	// initialize
	if err := t.initialize(src, dst); err != nil {
		return err
	}

	// delete the artifacts/tags on the destination registry
	if dst.Deleted {
		// delete tag
		if dst.IsDeleteTag {
			return t.deleteTag(dst)
		}
		// delete artifact
		return t.delete(t.convert(dst))
	}

	// copy the repository from source registry to the destination
	return t.copy(t.convert(src), t.convert(dst), dst.Override)
}

func (t *transfer) convert(resource *model.Resource) *repository {
	repository := &repository{
		repository: resource.Metadata.Repository.Name,
	}
	for _, artifact := range resource.Metadata.Artifacts {
		if len(artifact.Tags) > 0 {
			repository.tags = append(repository.tags, artifact.Tags...)
			continue
		}
		// no tags
		if len(artifact.Digest) > 0 {
			repository.tags = append(repository.tags, artifact.Digest)
		}
	}
	if len(repository.tags) > 0 {
		return repository
	}
	// fallback to vtags
	repository.tags = resource.Metadata.Vtags
	return repository
}

func (t *transfer) initialize(src *model.Resource, dst *model.Resource) error {
	if t.shouldStop() {
		return nil
	}
	// create client for source registry
	srcReg, err := createRegistry(src.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for source registry: %v", err)
		return err
	}
	t.src = srcReg
	t.logger.Infof("client for source registry [type: %s, URL: %s, insecure: %v] created",
		src.Registry.Type, src.Registry.URL, src.Registry.Insecure)

	// create client for destination registry
	dstReg, err := createRegistry(dst.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for destination registry: %v", err)
		return err
	}
	t.dst = dstReg
	t.logger.Infof("client for destination registry [type: %s, URL: %s, insecure: %v] created",
		dst.Registry.Type, dst.Registry.URL, dst.Registry.Insecure)

	return nil
}

func createRegistry(reg *model.Registry) (adapter.ArtifactRegistry, error) {
	factory, err := adapter.GetFactory(reg.Type)
	if err != nil {
		return nil, err
	}
	ad, err := factory.Create(reg)
	if err != nil {
		return nil, err
	}
	registry, ok := ad.(adapter.ArtifactRegistry)
	if !ok {
		return nil, errors.New("the adapter doesn't implement the \"ArtifactRegistry\" interface")
	}
	return registry, nil
}

func (t *transfer) shouldStop() bool {
	isStopped := t.isStopped()
	if isStopped {
		t.logger.Info("the job is stopped")
	}
	return isStopped
}

func (t *transfer) copy(src *repository, dst *repository, override bool) error {
	srcRepo := src.repository
	dstRepo := dst.repository
	t.logger.Infof("copying %s:[%s](source registry) to %s:[%s](destination registry)...",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))
	var err error
	for i := range src.tags {
		if e := t.copyArtifact(srcRepo, src.tags[i], dstRepo, dst.tags[i], override); e != nil {
			t.logger.Errorf(e.Error())
			err = e
		}
	}
	if err != nil {
		err = errors.New("got error during the whole transfer period, mark the job failure")
		t.logger.Error(err)
		return err
	}

	t.logger.Infof("copy %s:[%s](source registry) to %s:[%s](destination registry) completed",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))
	return nil
}

func (t *transfer) copyArtifact(srcRepo, srcRef, dstRepo, dstRef string, override bool) error {
	t.logger.Infof("copying %s:%s(source registry) to %s:%s(destination registry)...",
		srcRepo, srcRef, dstRepo, dstRef)
	// pull the manifest from the source registry
	manifest, digest, err := t.pullManifest(srcRepo, srcRef)
	if err != nil {
		return err
	}

	// check the existence of the artifact on the destination registry
	exist, digest2, err := t.exist(dstRepo, dstRef)
	if err != nil {
		return err
	}
	if exist {
		// the same artifact already exists
		if digest == digest2 {
			t.logger.Infof("the artifact %s:%s already exists on the destination registry, skip",
				dstRepo, dstRef)
			return nil
		}
		// the same name artifact exists, but not allowed to override
		if !override {
			t.logger.Warningf("the same name artifact %s:%s exists on the destination registry, but the \"override\" is set to false, skip",
				dstRepo, dstRef)
			return nil
		}
		// the same name artifact exists, but allowed to override
		t.logger.Warningf("the same name artifact %s:%s exists on the destination registry and the \"override\" is set to true, continue...",
			dstRepo, dstRef)
	}

	// copy contents between the source and destination registries
	for _, content := range manifest.References() {
		if err = t.copyContent(content, srcRepo, dstRepo); err != nil {
			return err
		}
	}

	// push the manifest to the destination registry
	if err := t.pushManifest(manifest, dstRepo, dstRef); err != nil {
		return err
	}

	t.logger.Infof("copy %s:%s(source registry) to %s:%s(destination registry) completed",
		srcRepo, srcRef, dstRepo, dstRef)
	return nil
}

// copy the content from source registry to destination according to its media type
func (t *transfer) copyContent(content distribution.Descriptor, srcRepo, dstRepo string) error {
	digest := content.Digest.String()
	switch content.MediaType {
	// when the media type of pulled manifest is index,
	// the contents it contains are a few manifests/indexes
	case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList,
		v1.MediaTypeImageManifest, schema2.MediaTypeManifest,
		schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest:
		// as using digest as the reference, so set the override to true directly
		return t.copyArtifact(srcRepo, digest, dstRepo, digest, true)
	// handle foreign layer
	case schema2.MediaTypeForeignLayer:
		t.logger.Infof("the layer %s is a foreign layer, skip", digest)
		return nil
	// copy layer or artifact config
	// the media type of the layer or config can be "application/octet-stream",
	// schema1.MediaTypeManifestLayer, schema2.MediaTypeLayer, schema2.MediaTypeImageConfig
	default:
		return t.copyBlob(srcRepo, dstRepo, digest)
	}
}

// copy the layer or artifact config from the source registry to destination
func (t *transfer) copyBlob(srcRepo, dstRepo, digest string) error {
	if t.shouldStop() {
		return nil
	}
	t.logger.Infof("copying the blob %s...", digest)
	exist, err := t.dst.BlobExist(dstRepo, digest)
	if err != nil {
		t.logger.Errorf("failed to check the existence of blob %s on the destination registry: %v", digest, err)
		return err
	}
	if exist {
		t.logger.Infof("the blob %s already exists on the destination registry, skip", digest)
		return nil
	}

	t.logger.Infof("the blob %s doesn't exist on the destination registry, continue", digest)

	size, data, err := t.src.PullBlob(srcRepo, digest)
	if err != nil {
		t.logger.Errorf("failed to pulling the blob %s: %v", digest, err)
		return err
	}
	defer data.Close()
	t.logger.Infof("the blob %s pulled, size: %d", digest, size)

	if err = t.dst.PushBlob(dstRepo, digest, size, data); err != nil {
		t.logger.Errorf("failed to pushing the blob %s: %v", digest, err)
		return err
	}
	t.logger.Infof("copy the blob %s completed", digest)
	return nil
}

func (t *transfer) pullManifest(repository, reference string) (
	distribution.Manifest, string, error) {
	if t.shouldStop() {
		return nil, "", nil
	}
	t.logger.Infof("pulling the manifest of artifact %s:%s ...", repository, reference)
	manifest, digest, err := t.src.PullManifest(repository, reference)
	if err != nil {
		t.logger.Errorf("failed to pull the manifest of artifact %s:%s: %v", repository, reference, err)
		return nil, "", err
	}
	t.logger.Infof("the manifest of artifact %s:%s pulled", repository, reference)

	return manifest, digest, nil
}

func (t *transfer) exist(repository, tag string) (bool, string, error) {
	exist, digest, err := t.dst.ManifestExist(repository, tag)
	if err != nil {
		t.logger.Errorf("failed to check the existence of the manifest of artifact %s:%s on the destination registry: %v",
			repository, tag, err)
		return false, "", err
	}
	return exist, digest, nil
}

func (t *transfer) pushManifest(manifest distribution.Manifest, repository, tag string) error {
	if t.shouldStop() {
		return nil
	}
	t.logger.Infof("pushing the manifest of artifact %s:%s ...", repository, tag)
	mediaType, payload, err := manifest.Payload()
	if err != nil {
		t.logger.Errorf("failed to push manifest of artifact %s:%s: %v",
			repository, tag, err)
		return err
	}
	if _, err := t.dst.PushManifest(repository, tag, mediaType, payload); err != nil {
		t.logger.Errorf("failed to push manifest of artifact %s:%s: %v",
			repository, tag, err)
		return err
	}
	t.logger.Infof("the manifest of artifact %s:%s pushed",
		repository, tag)
	return nil
}

func (t *transfer) delete(repo *repository) error {
	if t.shouldStop() {
		return nil
	}

	repository := repo.repository
	for _, tag := range repo.tags {
		exist, _, err := t.dst.ManifestExist(repository, tag)
		if err != nil {
			t.logger.Errorf("failed to check the existence of the manifest of artifact %s:%s on the destination registry: %v",
				repository, tag, err)
			return err
		}
		if !exist {
			t.logger.Infof("the artifact %s:%s doesn't exist on the destination registry, skip",
				repository, tag)
			continue
		}
		if err := t.dst.DeleteManifest(repository, tag); err != nil {
			t.logger.Errorf("failed to delete the manifest of artifact %s:%s on the destination registry: %v",
				repository, tag, err)
			return err
		}
		t.logger.Infof("the manifest of artifact %s:%s is deleted", repository, tag)
	}
	return nil
}

func (t *transfer) deleteTag(res *model.Resource) error {
	if t.shouldStop() {
		return nil
	}
	repository := res.Metadata.Repository.Name
	for _, art := range res.Metadata.Artifacts {
		for _, tag := range art.Tags {
			if err := t.dst.DeleteTag(repository, tag); err != nil {
				if e, ok := err.(*common_http.Error); ok && e.Code == http.StatusNotFound {
					t.logger.Infof("the tag %s:%s doesn't exist on the destination registry, skip", repository, tag)
					continue
				}
				t.logger.Errorf("failed to delete the tag %s:%s on the destination registry: %v",
					repository, tag, err)
				return err
			}
			t.logger.Infof("the tag %s:%s is deleted", repository, tag)
		}
	}
	return nil
}
