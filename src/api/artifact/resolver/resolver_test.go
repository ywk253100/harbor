// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resolver

import (
	"context"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/artifact/blob"
	"github.com/goharbor/harbor/src/pkg/artifact"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"
	"testing"
)

type fakeResolver struct{}

func (f *fakeResolver) ArtifactType(context.Context) string {
	return "FAKE_ARTIFACT"
}
func (f *fakeResolver) Resolve(ctx context.Context, manifest []byte, repoFullName string, artifact *artifact.Artifact, reader blob.Reader) error {
	return nil
}

type resolverTestSuite struct {
	suite.Suite
}

func (r *resolverTestSuite) SetupTest() {
	registry = map[string]Resolver{}
}

func (r *resolverTestSuite) TestRegister() {
	resolver := &fakeResolver{}
	mediaTypes := []string{v1.MediaTypeImageConfig, schema2.MediaTypeImageConfig}
	err := Register(resolver, mediaTypes...)
	r.Assert().Nil(err)

	// try to register a resolver for the existing media type
	err = Register(resolver, v1.MediaTypeImageConfig)
	r.Assert().NotNil(err)
}

func (r *resolverTestSuite) TestGet() {
	mediaType := v1.MediaTypeImageConfig
	// try to get the non-existing resolver
	resolver, err := Get(mediaType)
	r.Require().NotNil(err)

	// get the existing resolver
	err = Register(&fakeResolver{}, mediaType)
	r.Require().Nil(err)
	resolver, err = Get(mediaType)
	r.Require().Nil(err)
	r.Assert().NotNil(resolver)
}

func TestResolverTestSuite(t *testing.T) {
	suite.Run(t, &resolverTestSuite{})
}
