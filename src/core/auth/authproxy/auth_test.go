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

package authproxy

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/auth/authproxy/test"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"os"
	"testing"
)

var mockSvr *httptest.Server
var a *Auth
var pwd = "1234567ab"
var cmt = "By Authproxy"

func TestMain(m *testing.M) {
	mockSvr = test.NewMockServer(map[string]string{"jt": "pp", "Admin@vsphere.local": "Admin!23"})
	defer mockSvr.Close()
	a = &Auth{
		Endpoint:       mockSvr.URL + "/test/login",
		SkipCertVerify: true,
	}
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestAuth_Authenticate(t *testing.T) {
	t.Log("auth endpoint: ", a.Endpoint)
	type output struct {
		user models.User
		err  error
	}
	type tc struct {
		input  models.AuthModel
		expect output
	}
	suite := []tc{
		{
			input: models.AuthModel{
				Principal: "jt", Password: "pp"},
			expect: output{
				user: models.User{
					Username: "jt",
				},
				err: nil,
			},
		},
		{
			input: models.AuthModel{
				Principal: "Admin@vsphere.local",
				Password:  "Admin!23",
			},
			expect: output{
				user: models.User{
					Username: "Admin@vsphere.local",
					// Email:    "Admin@placeholder.com",
					// Password: pwd,
					// Comment:  fmt.Sprintf(cmtTmpl, path.Join(mockSvr.URL, "/test/login")),
				},
				err: nil,
			},
		},
		{
			input: models.AuthModel{
				Principal: "jt",
				Password:  "ppp",
			},
			expect: output{
				err: auth.ErrAuth{},
			},
		},
	}
	assert := assert.New(t)
	for _, c := range suite {
		r, e := a.Authenticate(c.input)
		if c.expect.err == nil {
			assert.Nil(e)
			assert.Equal(c.expect.user, *r)
		} else {
			assert.Nil(r)
			assert.NotNil(e)
			if _, ok := e.(auth.ErrAuth); ok {
				assert.IsType(auth.ErrAuth{}, e)
			}
		}
	}
}

/* TODO: Enable this case after adminserver refactor is merged.
func TestAuth_PostAuthenticate(t *testing.T) {
	type tc struct {
		input  *models.User
		expect models.User
	}
	suite := []tc{
		{
			input: &models.User{
				Username: "jt",
			},
			expect: models.User{
				Username: "jt",
				Email:    "jt@placeholder.com",
				Realname: "jt",
				Password: pwd,
				Comment:  fmt.Sprintf(cmtTmpl, mockSvr.URL+"/test/login"),
			},
		},
		{
			input: &models.User{
				Username: "Admin@vsphere.local",
			},
			expect: models.User{
				Username: "Admin@vsphere.local",
				Email:    "jt@placeholder.com",
				Realname: "Admin@vsphere.local",
				Password: pwd,
				Comment:  fmt.Sprintf(cmtTmpl, mockSvr.URL+"/test/login"),
			},
		},
	}
	for _, c := range suite {
		a.PostAuthenticate(c.input)
		assert.Equal(t, c.expect, *c.input)
	}
}
*/
