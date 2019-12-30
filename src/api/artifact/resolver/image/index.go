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
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/api/artifact/blob"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/opencontainers/image-spec/specs-go/v1"
)

var (
	// the media types that the indexResolver supports
	IndexResolverMediaTypes = []string{v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList}
)

// NewIndexResolver returns a resolver to resolve image with OCI index and docker manifest list
func NewIndexResolver() *indexResolver {
	return &indexResolver{}
}

// resolve artifact with OCI index and docker manifest list
type indexResolver struct {
}

func (m *indexResolver) ArtifactType(ctx context.Context) string {
	return ArtifactTypeImage
}
func (m *indexResolver) Resolve(ctx context.Context, manifest []byte, repoFullName string,
	artifact *artifact.Artifact, reader blob.Reader) error {
	// TODO implement
	// how to make sure the artifact referenced by the index has already been saved in database
	return nil
}
