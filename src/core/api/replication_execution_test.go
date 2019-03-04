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

package api

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/replication/ng"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

type fakedOperationController struct{}

func (f *fakedOperationController) StartReplication(policy *model.Policy) (int64, error) {
	return 1, nil
}
func (f *fakedOperationController) StopReplication(int64) error {
	return nil
}
func (f *fakedOperationController) ListExecutions(...*model.ExecutionQuery) (int64, []*model.Execution, error) {
	return 1, []*model.Execution{
		{
			ID:       1,
			PolicyID: 1,
		},
	}, nil
}
func (f *fakedOperationController) GetExecution(id int64) (*model.Execution, error) {
	if id == 1 {
		return &model.Execution{
			ID:       1,
			PolicyID: 1,
		}, nil
	}
	return nil, nil
}
func (f *fakedOperationController) ListTasks(...*model.TaskQuery) (int64, []*model.Task, error) {
	return 1, []*model.Task{
		{
			ID:          1,
			ExecutionID: 1,
		},
	}, nil
}
func (f *fakedOperationController) GetTask(id int64) (*model.Task, error) {
	if id == 1 {
		return &model.Task{
			ID:          1,
			ExecutionID: 1,
		}, nil
	}
	return nil, nil
}
func (f *fakedOperationController) GetTaskLog(int64) ([]byte, error) {
	return []byte("success"), nil
}

type fakedPolicyManager struct{}

func (f *fakedPolicyManager) Create(*model.Policy) (int64, error) {
	return 0, nil
}
func (f *fakedPolicyManager) List(...*model.PolicyQuery) (int64, []*model.Policy, error) {
	return 0, nil, nil
}
func (f *fakedPolicyManager) Get(id int64) (*model.Policy, error) {
	if id == 1 {
		return &model.Policy{
			ID:             1,
			SrcRegistryID:  1,
			SrcNamespaces:  []string{"library"},
			DestRegistryID: 2,
		}, nil
	}
	return nil, nil
}
func (f *fakedPolicyManager) GetByName(name string) (*model.Policy, error) {
	if name == "duplicate_name" {
		return &model.Policy{
			Name: "duplicate_name",
		}, nil
	}
	return nil, nil
}
func (f *fakedPolicyManager) Update(*model.Policy, ...string) error {
	return nil
}
func (f *fakedPolicyManager) Remove(int64) error {
	return nil
}

func TestListExecutions(t *testing.T) {
	operationCtl := ng.OperationCtl
	defer func() {
		ng.OperationCtl = operationCtl
	}()
	ng.OperationCtl = &fakedOperationController{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/replication/executions",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestCreateExecution(t *testing.T) {
	operationCtl := ng.OperationCtl
	policyMgr := ng.PolicyMgr
	defer func() {
		ng.OperationCtl = operationCtl
		ng.PolicyMgr = policyMgr
	}()
	ng.OperationCtl = &fakedOperationController{}
	ng.PolicyMgr = &fakedPolicyManager{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/replication/executions",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPost,
				url:        "/api/replication/executions",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/replication/executions",
				bodyJSON: &model.Execution{
					PolicyID: 2,
				},
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 201
		{
			request: &testingRequest{
				method: http.MethodPost,
				url:    "/api/replication/executions",
				bodyJSON: &model.Execution{
					PolicyID: 1,
				},
				credential: sysAdmin,
			},
			code: http.StatusCreated,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestStopExecution(t *testing.T) {
	operationCtl := ng.OperationCtl
	defer func() {
		ng.OperationCtl = operationCtl
	}()
	ng.OperationCtl = &fakedOperationController{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodPut,
				url:    "/api/replication/executions/1",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/executions/1",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/executions/2",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodPut,
				url:        "/api/replication/executions/1",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestListTasks(t *testing.T) {
	operationCtl := ng.OperationCtl
	defer func() {
		ng.OperationCtl = operationCtl
	}()
	ng.OperationCtl = &fakedOperationController{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/replication/executions/1/tasks",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions/1/tasks",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions/2/tasks",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions/1/tasks",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}

func TestGetTaskLog(t *testing.T) {
	operationCtl := ng.OperationCtl
	defer func() {
		ng.OperationCtl = operationCtl
	}()
	ng.OperationCtl = &fakedOperationController{}

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				method: http.MethodGet,
				url:    "/api/replication/executions/1/tasks/1/log",
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions/1/tasks/1/log",
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 404, execution not found
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions/2/tasks/1/log",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 404, task not found
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions/1/tasks/2/log",
				credential: sysAdmin,
			},
			code: http.StatusNotFound,
		},
		// 200
		{
			request: &testingRequest{
				method:     http.MethodGet,
				url:        "/api/replication/executions/1/tasks/1/log",
				credential: sysAdmin,
			},
			code: http.StatusOK,
		},
	}

	runCodeCheckingCases(t, cases...)
}
