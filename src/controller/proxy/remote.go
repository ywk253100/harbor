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
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/registry"
	"github.com/opencontainers/go-digest"
	"io"
)

type Remote struct {
	adapter *native.Adapter
	regID   int64
}

func (r *Remote) init() error {
	if r.adapter != nil {
		return nil
	}
	reg, err := registry.NewDefaultManager().Get(r.regID)
	if err != nil {
		return err
	}
	r.adapter = native.NewAdapter(reg)
	return nil
}

// Blob use remote to handler blob request
func (r *Remote) Blob(w io.Writer, repository string, dig string) (distribution.Descriptor, error) {
	d := distribution.Descriptor{}
	err := r.init()
	if err != nil {
		return d, err
	}
	size, bReader, err := r.adapter.PullBlob(repository, dig)
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

// BlobReader create a reader for remote blob
func (r *Remote) BlobReader(orgRepo, dig string) (int64, io.ReadCloser, error) {
	if err := r.init(); err != nil {
		return 0, nil, err
	}
	return r.adapter.PullBlob(orgRepo, dig)
}

// ManifestByDigest get manifest by digest
func (r *Remote) ManifestByDigest(repository string, dig string) (distribution.Manifest, error) {
	if err := r.init(); err != nil {
		return nil, err
	}
	man, dig, err := r.adapter.PullManifest(repository, dig)
	return man, err
}

// ManifestByTag get manifest by tag
func (r *Remote) ManifestByTag(repository string, tag string) (distribution.Manifest, error) {
	if err := r.init(); err != nil {
		return nil, err
	}
	man, _, err := r.adapter.PullManifest(repository, tag)
	return man, err
}
