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
	"context"
	"encoding/json"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/artifact/blob"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// ArtifactTypeImage is the artifact type for image
	ArtifactTypeImage = "IMAGE"
)

var (
	// the media types that the manifestV2Resolver supports
	ManifestV2ResolverMediaTypes = []string{v1.MediaTypeImageConfig, schema2.MediaTypeImageConfig}
)

// NewManifestV2Resolver returns a resolver to resolve image with OCI manifest and docker v2 manifest
func NewManifestV2Resolver() *manifestV2Resolver {
	return &manifestV2Resolver{}
}

// resolve artifact with OCI manifest and docker v2 manifest
type manifestV2Resolver struct {
}

func (m *manifestV2Resolver) ArtifactType(ctx context.Context) string {
	return ArtifactTypeImage
}
func (m *manifestV2Resolver) Resolve(ctx context.Context, content []byte, repoFullName string,
	artifact *artifact.Artifact, reader blob.Reader) error {
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(content, manifest); err != nil {
		return err
	}
	digest := manifest.Config.Digest.String()
	_, blob, err := reader.Read(repoFullName, digest, false)
	if err != nil {
		return err
	}
	image := &v1.Image{}
	if err := json.Unmarshal(blob, image); err != nil {
		return err
	}
	artifact.ExtraAttrs = map[string]interface{}{
		"created":      image.Created,
		"author":       image.Author,
		"architecture": image.Architecture,
		"os":           image.OS,
	}
	return nil
}
