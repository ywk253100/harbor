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

package session

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sess "github.com/astaxie/beego/session"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
)

func TestCreate(t *testing.T) {
	factory := &sessionAuthenticatorFactory{}

	var params map[string]interface{}

	params = nil
	_, err := factory.Create(params)
	assert.NotNil(t, err)

	params = map[string]interface{}{
		"wrong_key": nil,
	}
	_, err = factory.Create(params)
	assert.NotNil(t, err)

	params = map[string]interface{}{
		"manager": "wrong_type",
	}
	_, err = factory.Create(params)
	assert.NotNil(t, err)

	params = map[string]interface{}{
		"manager": &sess.Manager{},
	}
	_, err = factory.Create(params)
	assert.Nil(t, err)
}

func TestAuthenticate(t *testing.T) {
	manager, err := sess.NewManager("memory", "{}")
	if err != nil {
		t.Fatalf("failed to create the session manager: %v", err)
	}

	w := &httptest.ResponseRecorder{}
	r, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create the request: %v", err)
	}
	sessionStore, err := manager.SessionStart(w, r)
	if err != nil {
		t.Fatalf("failed to start a session: %v", err)
	}

	sessionID := sessionStore.SessionID()
	sessionStore.Set("user", models.User{
		Username: "test",
	})

	authenticator := &session{
		manager: manager,
	}
	var params map[string]interface{}

	// parameters are nil
	params = nil
	_, err = authenticator.Authenticate(nil, params)
	assert.NotNil(t, err)

	// parameters are not nil, but contain no parameter
	params = map[string]interface{}{}
	_, err = authenticator.Authenticate(nil, params)
	assert.NotNil(t, err)

	// parameters contain wrong key
	params = map[string]interface{}{
		"wrong_key": nil,
	}
	_, err = authenticator.Authenticate(nil, params)
	assert.NotNil(t, err)

	// parameters contain wrong type
	params = map[string]interface{}{
		"sessionID": nil,
	}
	_, err = authenticator.Authenticate(nil, params)
	assert.NotNil(t, err)

	// sessionID is null
	params = map[string]interface{}{
		"sessionID": "",
	}
	_, err = authenticator.Authenticate(nil, params)
	assert.NotNil(t, err)

	// invalid sessionID
	params = map[string]interface{}{
		"sessionID": "invalid_session_id",
	}
	_, err = authenticator.Authenticate(nil, params)
	assert.NotNil(t, err)

	// valid sessionID
	params = map[string]interface{}{
		"sessionID": sessionID,
	}
	ctx, err := authenticator.Authenticate(nil, params)
	assert.Nil(t, err)

	u := ctx.Value(common.CtxKeyUser)
	user, ok := u.(models.User)
	if !ok {
		t.Fatalf("user got from session is not User type")
	}

	assert.Equal(t, "test", user.Username)
}
