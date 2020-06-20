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
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/registry"
	"github.com/opencontainers/go-digest"
	"io"
)

// RemoteInterface defines operations related to remote repository under proxy
type RemoteInterface interface {
	// Blob use remote to handler blob request
	Blob(w io.Writer, repository string, dig string) (distribution.Descriptor, error)
	// BlobReader create a reader for remote blob
	BlobReader(orgRepo, dig string) (int64, io.ReadCloser, error)
	// ManifestByDigest get manifest by digest
	ManifestByDigest(repository string, dig string) (distribution.Manifest, error)
	// ManifestByTag get manifest by tag
	ManifestByTag(repository string, tag string) (distribution.Manifest, error)
}

// remote defines operations related to remote repository under proxy
type remote struct {
	regID    int64
	registry adapter.ArtifactRegistry
}

// CreateRemoteInterface create a remote interface
func CreateRemoteInterface(regID int64) RemoteInterface {
	return &remote{regID: regID}
}

func (r *remote) init() error {

	if r.registry != nil {
		return nil
	}
	reg, err := registry.NewDefaultManager().Get(r.regID)
	factory, err := adapter.GetFactory(reg.Type)
	if err != nil {
		return err
	}
	adp, err := factory.Create(reg)
	if err != nil {
		return err
	}
	r.registry = adp.(adapter.ArtifactRegistry)
	return nil
}

func (r *remote) Blob(w io.Writer, repository string, dig string) (distribution.Descriptor, error) {
	d := distribution.Descriptor{}
	err := r.init()
	if err != nil {
		return d, err
	}
	size, bReader, err := r.registry.PullBlob(repository, dig)
	defer bReader.Close()
	if err != nil {
		log.Error(err)
	}
	written, err := io.CopyN(w, bReader, size)
	if err != nil {
		log.Error(err)
	}
	if written != size {
		log.Errorf("The size mismatch, actual:%d, expected: %d", written, size)
	}
	d.Size = size
	d.Digest = digest.Digest(dig)
	return d, err
}

func (r *remote) BlobReader(orgRepo, dig string) (int64, io.ReadCloser, error) {
	if err := r.init(); err != nil {
		return 0, nil, err
	}
	return r.registry.PullBlob(orgRepo, dig)
}

func (r *remote) ManifestByDigest(repository string, dig string) (distribution.Manifest, error) {
	if err := r.init(); err != nil {
		return nil, err
	}
	man, dig, err := r.registry.PullManifest(repository, dig)
	return man, err
}

func (r *remote) ManifestByTag(repository string, tag string) (distribution.Manifest, error) {
	if err := r.init(); err != nil {
		return nil, err
	}
	man, _, err := r.registry.PullManifest(repository, tag)
	return man, err
}
