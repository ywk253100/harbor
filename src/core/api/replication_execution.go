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
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/replication/ng"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// ReplicationOperationAPI handles the replication operation requests
type ReplicationOperationAPI struct {
	BaseController
}

// Prepare ...
func (r *ReplicationOperationAPI) Prepare() {
	r.BaseController.Prepare()
	// TODO if we delegate the jobservice to trigger the scheduled replication,
	// add the logic to check whether the user is a solution user
	if !r.SecurityCtx.IsSysAdmin() {
		if !r.SecurityCtx.IsAuthenticated() {
			r.HandleUnauthorized()
			return
		}
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}
}

// The API is open only for system admin currently, we can use
// the code commentted below to make the API available to the
// users who have permission for all projects that the policy
// refers
/*
func (r *ReplicationOperationAPI) authorized(policy *model.Policy, resource rbac.Resource, action rbac.Action) bool {

	projects := []string{}
	// pull mode
	if policy.SrcRegistryID != 0 {
		projects = append(projects, policy.DestNamespace)
	} else {
		// push mode
		projects = append(projects, policy.SrcNamespaces...)
	}

	for _, project := range projects {
		resource := rbac.NewProjectNamespace(project).Resource(resource)
		if !r.SecurityCtx.Can(action, resource) {
			r.HandleForbidden(r.SecurityCtx.GetUsername())
			return false
		}
	}

	return true
}
*/

// ListExecutions ...
func (r *ReplicationOperationAPI) ListExecutions() {
	query := &model.ExecutionQuery{
		Status:  r.GetString("status"),
		Trigger: r.GetString("trigger"),
	}
	if len(r.GetString("policy_id")) > 0 {
		policyID, err := r.GetInt64("policy_id")
		if err != nil || policyID <= 0 {
			r.HandleBadRequest(fmt.Sprintf("invalid policy_id %s", r.GetString("policy_id")))
			return
		}
		query.PolicyID = policyID
	}
	query.Page, query.Size = r.GetPaginationParams()
	total, executions, err := ng.OperationCtl.ListExecutions(query)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to list executions: %v", err))
		return
	}
	r.SetPaginationHeader(total, query.Page, query.Size)
	r.WriteJSONData(executions)
}

// CreateExecution starts a replication
func (r *ReplicationOperationAPI) CreateExecution() {
	execution := &model.Execution{}
	r.DecodeJSONReq(execution)
	policy, err := ng.PolicyMgr.Get(execution.PolicyID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get policy %d: %v", execution.PolicyID, err))
		return
	}

	if policy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", execution.PolicyID))
		return
	}

	executionID, err := ng.OperationCtl.StartReplication(policy)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to start replication for policy %d: %v", execution.PolicyID, err))
		return
	}
	r.Redirect(http.StatusCreated, strconv.FormatInt(executionID, 10))
}

// StopExecution stops one execution of the replication
func (r *ReplicationOperationAPI) StopExecution() {
	executionID, err := r.GetInt64FromPath(":id")
	if err != nil || executionID <= 0 {
		r.HandleBadRequest("invalid execution ID")
		return
	}
	execution, err := ng.OperationCtl.GetExecution(executionID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get execution %d: %v", executionID, err))
		return
	}

	if execution == nil {
		r.HandleNotFound(fmt.Sprintf("execution %d not found", executionID))
		return
	}

	if err := ng.OperationCtl.StopReplication(executionID); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to stop execution %d: %v", executionID, err))
		return
	}
}

// ListTasks ...
func (r *ReplicationOperationAPI) ListTasks() {
	executionID, err := r.GetInt64FromPath(":id")
	if err != nil || executionID <= 0 {
		r.HandleBadRequest("invalid execution ID")
		return
	}

	execution, err := ng.OperationCtl.GetExecution(executionID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get execution %d: %v", executionID, err))
		return
	}
	if execution == nil {
		r.HandleNotFound(fmt.Sprintf("execution %d not found", executionID))
		return
	}

	query := &model.TaskQuery{
		ExecutionID:  executionID,
		ResourceType: (model.ResourceType)(r.GetString("resource_type")),
		Status:       r.GetString("status"),
	}
	query.Page, query.Size = r.GetPaginationParams()
	total, tasks, err := ng.OperationCtl.ListTasks(query)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to list tasks: %v", err))
		return
	}
	r.SetPaginationHeader(total, query.Page, query.Size)
	r.WriteJSONData(tasks)
}

// GetTaskLog ...
func (r *ReplicationOperationAPI) GetTaskLog() {
	executionID, err := r.GetInt64FromPath(":id")
	if err != nil || executionID <= 0 {
		r.HandleBadRequest("invalid execution ID")
		return
	}

	execution, err := ng.OperationCtl.GetExecution(executionID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get execution %d: %v", executionID, err))
		return
	}
	if execution == nil {
		r.HandleNotFound(fmt.Sprintf("execution %d not found", executionID))
		return
	}

	taskID, err := r.GetInt64FromPath(":tid")
	if err != nil || taskID <= 0 {
		r.HandleBadRequest("invalid task ID")
		return
	}
	task, err := ng.OperationCtl.GetTask(taskID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get task %d: %v", taskID, err))
		return
	}
	if task == nil {
		r.HandleNotFound(fmt.Sprintf("task %d not found", taskID))
		return
	}

	logBytes, err := ng.OperationCtl.GetTaskLog(taskID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get log of task %d: %v", taskID, err))
		return
	}
	r.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	r.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = r.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to write log of task %d: %v", taskID, err))
		return
	}
}
