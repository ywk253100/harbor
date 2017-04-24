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

package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeAuthenticator struct{}

func (f *fakeAuthenticator) Authenticate(context context.Context,
	parameters map[string]interface{}) (context.Context, error) {
	return nil, nil
}

type fakeAuthenticatorFactory struct{}

func (f *fakeAuthenticatorFactory) Create(parameters map[string]interface{}) (
	Authenticator, error) {
	return &fakeAuthenticator{}, nil
}

func TestCreate(t *testing.T) {
	_, err := Create("fake", nil)
	assert.NotNil(t, err)

	factory := &fakeAuthenticatorFactory{}
	Register("fake", factory)

	_, err = Create("fake", nil)
	assert.Nil(t, err)
}
