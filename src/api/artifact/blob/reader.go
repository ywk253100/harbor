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

package blob

import (
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	coreutils "github.com/goharbor/harbor/src/core/utils"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io/ioutil"
)

var (
	accept = []string{
		schema1.MediaTypeSignedManifest,
		schema2.MediaTypeManifest,
		v1.MediaTypeImageManifest,
		manifestlist.MediaTypeManifestList,
		v1.MediaTypeImageIndex,
	}
)

type Reader interface {
	Read(name, digest string, manifest bool) (mediaType string, content []byte, err error)
}

func NewReader() Reader {
	return &reader{}
}

type reader struct{}

// TODO re-implement it based on OCI registry driver
func (r *reader) Read(name, digest string, manifest bool) (string, []byte, error) {
	// TODO read from cache first
	client, err := coreutils.NewRepositoryClientForLocal("admin", name)
	if err != nil {
		return "", nil, err
	}
	if manifest {
		_, mediaType, payload, err := client.PullManifest(digest, accept)
		return mediaType, payload, err
	}
	_, reader, err := client.PullBlob(digest)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	return "", data, err
}
