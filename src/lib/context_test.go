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

package lib

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetAPIVersion(t *testing.T) {
	ctx := WithAPIVersion(context.Background(), "1.0")
	assert.NotNil(t, ctx)
}

func TestGetAPIVersion(t *testing.T) {
	// nil context
	version := GetAPIVersion(nil)
	assert.Empty(t, version)

	// no version set in context
	version = GetAPIVersion(context.Background())
	assert.Empty(t, version)

	// version set in context
	ctx := WithAPIVersion(context.Background(), "1.0")
	version = GetAPIVersion(ctx)
	assert.Equal(t, "1.0", version)
}

func TestRemoteRepo(t *testing.T) {

	cases := []struct {
		name string
		in   ArtifactInfo
		want string
	}{
		{
			name: `normal test`,
			in:   ArtifactInfo{ProjectName: "dockerhub_proxy", Repository: "dockerhub_proxy/firstfloor/hello-world"},
			want: "firstfloor/hello-world",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.ProxyCacheRemoteRepo()
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}

		})
	}

}
