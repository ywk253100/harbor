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
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/replication/ng"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// TODO rename the file to "replication.go"

// ReplicationPolicyAPI handles the replication policy requests
type ReplicationPolicyAPI struct {
	BaseController
}

// Prepare ...
func (r *ReplicationPolicyAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsSysAdmin() {
		if !r.SecurityCtx.IsAuthenticated() {
			r.HandleUnauthorized()
			return
		}
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}
}

// List the replication policies
func (r *ReplicationPolicyAPI) List() {
	// TODO: support more query
	query := &model.PolicyQuery{
		Name: r.GetString("name"),
	}
	query.Page, query.Size = r.GetPaginationParams()

	total, policies, err := ng.PolicyMgr.List(query)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to list policies: %v", err))
		return
	}
	r.SetPaginationHeader(total, query.Page, query.Size)
	r.WriteJSONData(policies)
}

// Create the replication policy
func (r *ReplicationPolicyAPI) Create() {
	policy := &model.Policy{}
	r.DecodeJSONReqAndValidate(policy)

	if !r.validateName(policy) {
		return
	}
	if !r.validateRegistry(policy) {
		return
	}

	id, err := ng.PolicyMgr.Create(policy)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to create the policy: %v", err))
		return
	}

	// TODO handle replication_now?

	r.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// make sure the policy name doesn't exist
func (r *ReplicationPolicyAPI) validateName(policy *model.Policy) bool {
	p, err := ng.PolicyMgr.GetByName(policy.Name)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get policy %s: %v", policy.Name, err))
		return false
	}
	if p != nil {
		r.HandleConflict(fmt.Sprintf("policy %s already exists", policy.Name))
		return false
	}
	return true
}

// make the registry referenced exists
func (r *ReplicationPolicyAPI) validateRegistry(policy *model.Policy) bool {
	registryID := policy.SrcRegistryID
	if registryID == 0 {
		registryID = policy.DestRegistryID
	}
	registry, err := ng.RegistryMgr.Get(registryID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get registry %d: %v", registryID, err))
		return false
	}
	if registry == nil {
		r.HandleNotFound(fmt.Sprintf("registry %d not found", registryID))
		return false
	}
	return true
}

// TODO validate trigger in create and update

// Get the specified replication policy
func (r *ReplicationPolicyAPI) Get() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.HandleBadRequest("invalid policy ID")
		return
	}

	policy, err := ng.PolicyMgr.Get(id)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the policy %d: %v", id, err))
		return
	}
	if policy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	r.WriteJSONData(policy)
}

// Update the replication policy
func (r *ReplicationPolicyAPI) Update() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.HandleBadRequest("invalid policy ID")
		return
	}

	originalPolicy, err := ng.PolicyMgr.Get(id)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the policy %d: %v", id, err))
		return
	}
	if originalPolicy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	policy := &model.Policy{}
	r.DecodeJSONReqAndValidate(policy)
	if policy.Name != originalPolicy.Name &&
		!r.validateName(policy) {
		return
	}

	if !r.validateRegistry(policy) {
		return
	}

	// TODO passing the properties need to be updated?
	if err := ng.PolicyMgr.Update(policy); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to update the policy %d: %v", id, err))
		return
	}
}

// Delete the replication policy
func (r *ReplicationPolicyAPI) Delete() {
	id, err := r.GetInt64FromPath(":id")
	if id <= 0 || err != nil {
		r.HandleBadRequest("invalid policy ID")
		return
	}

	policy, err := ng.PolicyMgr.Get(id)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the policy %d: %v", id, err))
		return
	}
	if policy == nil {
		r.HandleNotFound(fmt.Sprintf("policy %d not found", id))
		return
	}

	_, executions, err := ng.OperationCtl.ListExecutions(&model.ExecutionQuery{
		PolicyID: id,
	})
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get the executions of policy %d: %v", id, err))
		return
	}

	for _, execution := range executions {
		if execution.Status == model.ExecutionStatusInProgress {
			r.HandleStatusPreconditionFailed(fmt.Sprintf("the policy %d has running executions, can not be deleted", id))
			return
		}
	}

	if err := ng.PolicyMgr.Remove(id); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to delete the policy %d: %v", id, err))
		return
	}
}
