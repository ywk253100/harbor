// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"context"
	"fmt"

	sess "github.com/astaxie/beego/session"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
)

// Name of session authenticator
const Name = "session"

func init() {
	auth.Register(Name, &sessionAuthenticatorFactory{})
}

type session struct {
	manager *sess.Manager
}

// Authenticate users with session and add user information into the context
// parameters contain the sessionID
func (s *session) Authenticate(ctx context.Context,
	parameters map[string]interface{}) (context.Context, error) {

	if ctx == nil {
		ctx = context.Background()
	}

	if parameters == nil {
		return ctx, fmt.Errorf("parameters should not be null")
	}

	sessionID := ""
	sID, exist := parameters["sessionID"]
	if !exist {
		return ctx, fmt.Errorf("sessionID should not be null")
	}

	sessionID, ok := sID.(string)
	if !ok {
		return ctx, fmt.Errorf("sessionID should be string type")
	}

	if len(sessionID) == 0 {
		return ctx, fmt.Errorf("sessionID should not be null")
	}

	sessionStore, err := s.manager.GetSessionStore(sessionID)
	if err != nil {
		return ctx, err
	}

	if sessionStore == nil {
		return ctx, fmt.Errorf("invalid session ID: %s", sessionID)
	}

	u := sessionStore.Get("user")
	if u == nil {
		return ctx, fmt.Errorf("can not get user from session %s", sessionID)
	}

	user, ok := u.(models.User)
	if !ok {
		return ctx, fmt.Errorf("user got from session %s is not User type", sessionID)
	}

	ctx = context.WithValue(ctx, common.CtxKeyUser, user)
	log.Infof("user %s authenticated by %s has been added into context",
		user.Username, Name)
	return ctx, nil
}

type sessionAuthenticatorFactory struct{}

// parameters contain session manager parameter named manager
func (s *sessionAuthenticatorFactory) Create(parameters map[string]interface{}) (
	auth.Authenticator, error) {
	if parameters == nil {
		return nil, fmt.Errorf("parameters should not be null")
	}

	mg, exist := parameters["manager"]
	if !exist {
		return nil, fmt.Errorf("manager should not be null")
	}

	manager, ok := mg.(*sess.Manager)
	if !ok {
		return nil, fmt.Errorf("manager should be beego session.Manager type")
	}

	return &session{
		manager: manager,
	}, nil
}
