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
	"io"
)

// RemoteInterface defines operations related to remote repository under proxy
type RemoteInterface interface {
	// BlobReader create a reader for remote blob
	BlobReader(orgRepo, dig string) (int64, io.ReadCloser, error)
	// Manifest get manifest by reference
	Manifest(repository string, dig string) (distribution.Manifest, error)
}

// remote defines operations related to remote repository under proxy
type remote struct {
	regID    int64
	registry adapter.ArtifactRegistry
}

// CreateRemoteInterface create a remote interface
func CreateRemoteInterface(regID int64) RemoteInterface {
	r := &remote{regID: regID}
	if err := r.init(); err != nil {
		log.Errorf("failed to create remote error %v", err)
	}
	return r
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

func (r *remote) BlobReader(orgRepo, dig string) (int64, io.ReadCloser, error) {
	return r.registry.PullBlob(orgRepo, dig)
}

func (r *remote) Manifest(repository string, ref string) (distribution.Manifest, error) {
	man, _, err := r.registry.PullManifest(repository, ref)
	return man, err
}
