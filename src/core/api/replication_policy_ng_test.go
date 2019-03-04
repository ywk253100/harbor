// Copyright 2018 Project Harbor Authors
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

package api

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/replication/ng/model"

	"github.com/goharbor/harbor/src/replication/ng"
)

// TODO rename the file to "replication.go"

type fakedRegistryManager struct{}

func (f *fakedRegistryManager) Add(*model.Registry) (int64, error) {
	return 0, nil
}
func (f *fakedRegistryManager) List(...*model.RegistryQuery) (int64, []*model.Registry, error) {
	return 0, nil, nil
}
func (f *fakedRegistryManager) Get(id int64) (*model.Registry, error) {
	if id == 1 {
		return &model.Registry{
			Type: "faked_registry",
		}, nil
	}
	return nil, nil
}
func (f *fakedRegistryManager) Update(*model.Registry, ...string) error {
	return nil
}
func (f *fakedRegistryManager) Remove(int64) error {
	return nil
}

func TestReplicationPolicyAPIList(t *testing.T) {
	policyMgr := ng.PolicyMgr
	defer func() {
		ng.PolicyMgr = policyMgr
	}()
	ng.PolicyMgr = &fakedPolicyManager{}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/replication/policies",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/policies",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/policies",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestReplicationPolicyAPICreate(t *testing.T) {
	policyMgr := ng.PolicyMgr
	registryMgr := ng.RegistryMgr
	defer func() {
		ng.PolicyMgr = policyMgr
		ng.RegistryMgr = registryMgr
	}()
	ng.PolicyMgr = &fakedPolicyManager{}
	ng.RegistryMgr = &fakedRegistryManager{}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/replication/policies",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/policies",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 400 empty policy name
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/policies",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					SrcRegistryID: 1,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusBadRequest,
		},
		// 400 empty registry
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/policies",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "policy01",
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusBadRequest,
		},
		// 400 empty source namespaces
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/policies",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "policy01",
					SrcRegistryID: 1,
				},
			},
			code: http.StatusBadRequest,
		},
		// 409, duplicate policy name
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/policies",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "duplicate_name",
					SrcRegistryID: 1,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusConflict,
		},
		// 404, registry not found
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/policies",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "policy01",
					SrcRegistryID: 2,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusNotFound,
		},
		// 201
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/policies",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "policy01",
					SrcRegistryID: 1,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusCreated,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestReplicationPolicyAPIGet(t *testing.T) {
	policyMgr := ng.PolicyMgr
	defer func() {
		ng.PolicyMgr = policyMgr
	}()
	ng.PolicyMgr = &fakedPolicyManager{}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/replication/policies/1",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/policies/1",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404, policy not found
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/policies/2",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/policies/1",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestReplicationPolicyAPIUpdate(t *testing.T) {
	policyMgr := ng.PolicyMgr
	registryMgr := ng.RegistryMgr
	defer func() {
		ng.PolicyMgr = policyMgr
		ng.RegistryMgr = registryMgr
	}()
	ng.PolicyMgr = &fakedPolicyManager{}
	ng.RegistryMgr = &fakedRegistryManager{}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    "/api/replication/policies/1",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/policies/1",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404 policy not found
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/policies/2",
				credential: sysAdmin,
				bodyJSON:   &model.Policy{},
			},
			code: http.StatusNotFound,
		},
		// 400 empty policy name
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/policies/1",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					SrcRegistryID: 1,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusBadRequest,
		},
		// 409, duplicate policy name
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/policies/1",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "duplicate_name",
					SrcRegistryID: 1,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusConflict,
		},
		// 404, registry not found
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/policies/1",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "policy01",
					SrcRegistryID: 2,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/policies/1",
				credential: sysAdmin,
				bodyJSON: &model.Policy{
					Name:          "policy01",
					SrcRegistryID: 1,
					SrcNamespaces: []string{"library"},
				},
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestReplicationPolicyAPIDelete(t *testing.T) {
	policyMgr := ng.PolicyMgr
	defer func() {
		ng.PolicyMgr = policyMgr
	}()
	ng.PolicyMgr = &fakedPolicyManager{}
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodDelete,
				url:    "/api/replication/policies/1",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/replication/policies/1",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404, policy not found
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/replication/policies/2",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodDelete,
				url:        "/api/replication/policies/1",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}
