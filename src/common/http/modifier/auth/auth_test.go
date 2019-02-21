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

package auth

import (
	"net/http"
	"testing"

	commonsecret "github.com/goharbor/harbor/src/common/secret"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizeOfSecretAuthorizer(t *testing.T) {
	secret := "secret"
	authorizer := NewSecretAuthorizer(secret)

	// nil request
	require.NotNil(t, authorizer.Modify(nil))

	// valid request
	req, err := http.NewRequest("", "", nil)
	require.Nil(t, err)
	require.Nil(t, authorizer.Modify(req))
	assert.Equal(t, secret, commonsecret.FromRequest(req))
}

func TestAuthorizeOfBasicAuthorizer(t *testing.T) {
	username := "username"
	password := "password"
	authorizer := NewBasicAuthorizer(username, password)

	// nil request
	require.NotNil(t, authorizer.Modify(nil))

	// valid request
	req, err := http.NewRequest("", "", nil)
	require.Nil(t, err)
	require.Nil(t, authorizer.Modify(req))
	usr, pwd, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, username, usr)
	assert.Equal(t, password, pwd)
}
