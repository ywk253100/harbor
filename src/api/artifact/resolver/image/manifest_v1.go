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
	"github.com/docker/distribution/manifest/schema1"
	"github.com/goharbor/harbor/src/api/artifact/blob"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

var (
	// the media types that the manifestV1Resolver supports
	ManifestV1ResolverMediaTypes = []string{schema1.MediaTypeSignedManifest}
)

// NewManifestV1Resolver returns a resolver to resolve image with docker v1 manifest
func NewManifestV1Resolver() *manifestV1Resolver {
	return &manifestV1Resolver{}
}

// resolve artifact with docker v1 manifest
type manifestV1Resolver struct {
}

func (m *manifestV1Resolver) ArtifactType(ctx context.Context) string {
	return ArtifactTypeImage
}
func (m *manifestV1Resolver) Resolve(ctx context.Context, manifest []byte, repoFullName string,
	artifact *artifact.Artifact, reader blob.Reader) error {
	// TODO implement
	return nil
}
