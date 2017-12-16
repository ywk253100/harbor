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

package api

import (
	"fmt"

	"net/http"
	"strconv"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/replication/core"
	rep_models "github.com/vmware/harbor/src/replication/models"
	api_models "github.com/vmware/harbor/src/ui/api/models"
	"github.com/vmware/harbor/src/ui/promgr"
)

// RepPolicyAPI handles /api/replicationPolicies /api/replicationPolicies/:id/enablement
type RepPolicyAPI struct {
	BaseController
}

// Prepare validates whether the user has system admin role
func (pa *RepPolicyAPI) Prepare() {
	pa.BaseController.Prepare()
	if !pa.SecurityCtx.IsAuthenticated() {
		pa.HandleUnauthorized()
		return
	}

	if pa.Ctx.Request.Method != http.MethodGet && !pa.SecurityCtx.IsSysAdmin() {
		pa.HandleForbidden(pa.SecurityCtx.GetUsername())
		return
	}
}

// Get ...
func (pa *RepPolicyAPI) Get() {
	id := pa.GetIDFromURL()
	policy, err := core.GlobalController.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if policy.ID == 0 {
		pa.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	if !pa.SecurityCtx.HasAllPerm(policy.ProjectIDs[0]) {
		pa.HandleForbidden(pa.SecurityCtx.GetUsername())
		return
	}

	ply, err := convertFromRepPolicy(pa.ProjectMgr, policy)
	if err != nil {
		pa.ParseAndHandleError(fmt.Sprintf("failed to convert from replication policy"), err)
		return
	}

	pa.Data["json"] = ply
	pa.ServeJSON()
}

// List ...
func (pa *RepPolicyAPI) List() {
	queryParam := rep_models.QueryParameter{
		Name: pa.GetString("name"),
	}
	projectIDStr := pa.GetString("project_id")
	if len(projectIDStr) > 0 {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			pa.CustomAbort(http.StatusBadRequest, "invalid project ID")
		}
		queryParam.ProjectID = projectID
	}

	result := []*api_models.ReplicationPolicy{}

	policies, err := core.GlobalController.GetPolicies(queryParam)
	if err != nil {
		log.Errorf("failed to get policies: %v, query parameters: %v", err, queryParam)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	for _, policy := range policies {
		if !pa.SecurityCtx.HasAllPerm(policy.ProjectIDs[0]) {
			continue
		}
		ply, err := convertFromRepPolicy(pa.ProjectMgr, policy)
		if err != nil {
			pa.ParseAndHandleError(fmt.Sprintf("failed to convert from replication policy"), err)
			return
		}
		result = append(result, ply)
	}

	pa.Data["json"] = result
	pa.ServeJSON()
}

// Post creates a replicartion policy
func (pa *RepPolicyAPI) Post() {
	policy := &api_models.ReplicationPolicy{}
	pa.DecodeJSONReqAndValidate(policy)

	// check the existence of projects
	for _, project := range policy.Projects {
		pro, err := pa.ProjectMgr.Get(project.ProjectID)
		if err != nil {
			pa.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %d", project.ProjectID), err)
			return
		}
		if pro == nil {
			pa.HandleNotFound(fmt.Sprintf("project %d not found", project.ProjectID))
			return
		}
		project.Name = pro.Name
	}

	// check the existence of targets
	for _, target := range policy.Targets {
		t, err := dao.GetRepTarget(target.ID)
		if err != nil {
			pa.HandleInternalServerError(fmt.Sprintf("failed to get target %d: %v", target.ID, err))
			return
		}

		if t == nil {
			pa.HandleNotFound(fmt.Sprintf("target %d not found", target.ID))
			return
		}
	}

	id, err := core.GlobalController.CreatePolicy(convertToRepPolicy(policy))
	if err != nil {
		pa.HandleInternalServerError(fmt.Sprintf("failed to create policy: %v", err))
		return
	}

	if policy.ReplicateExistingImageNow {
		go func() {
			if err = startReplication(id); err != nil {
				log.Errorf("failed to send replication signal for policy %d: %v", id, err)
				return
			}
			log.Infof("replication signal for policy %d sent", id)
		}()
	}

	pa.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put updates the replication policy
func (pa *RepPolicyAPI) Put() {
	id := pa.GetIDFromURL()

	originalPolicy, err := core.GlobalController.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if originalPolicy.ID == 0 {
		pa.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	policy := &api_models.ReplicationPolicy{}
	pa.DecodeJSONReqAndValidate(policy)

	policy.ID = id

	// check the existence of projects
	for _, project := range policy.Projects {
		pro, err := pa.ProjectMgr.Get(project.ProjectID)
		if err != nil {
			pa.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %d", project.ProjectID), err)
			return
		}
		if pro == nil {
			pa.HandleNotFound(fmt.Sprintf("project %d not found", project.ProjectID))
			return
		}
		project.Name = pro.Name
	}

	// check the existence of targets
	for _, target := range policy.Targets {
		t, err := dao.GetRepTarget(target.ID)
		if err != nil {
			pa.HandleInternalServerError(fmt.Sprintf("failed to get target %d: %v", target.ID, err))
			return
		}

		if t == nil {
			pa.HandleNotFound(fmt.Sprintf("target %d not found", target.ID))
			return
		}
	}

	if err = core.GlobalController.UpdatePolicy(convertToRepPolicy(policy)); err != nil {
		pa.HandleInternalServerError(fmt.Sprintf("failed to update policy %d: %v", id, err))
		return
	}

	if policy.ReplicateExistingImageNow {
		go func() {
			if err = startReplication(id); err != nil {
				log.Errorf("failed to send replication signal for policy %d: %v", id, err)
				return
			}
			log.Infof("replication signal for policy %d sent", id)
		}()
	}
}

// Delete the replication policy
func (pa *RepPolicyAPI) Delete() {
	id := pa.GetIDFromURL()

	policy, err := core.GlobalController.GetPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if policy.ID == 0 {
		pa.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	// TODO
	jobs, err := dao.GetRepJobByPolicy(id)
	if err != nil {
		log.Errorf("failed to get jobs of policy %d: %v", id, err)
		pa.CustomAbort(http.StatusInternalServerError, "")
	}

	for _, job := range jobs {
		if job.Status == models.JobRunning ||
			job.Status == models.JobRetrying ||
			job.Status == models.JobPending {
			pa.CustomAbort(http.StatusPreconditionFailed, "policy has running/retrying/pending jobs, can not be deleted")
		}
	}

	if err = core.GlobalController.RemovePolicy(id); err != nil {
		log.Errorf("failed to delete policy %d: %v", id, err)
		pa.CustomAbort(http.StatusInternalServerError, "")
	}
}

func convertFromRepPolicy(projectMgr promgr.ProjectManager, policy rep_models.ReplicationPolicy) (*api_models.ReplicationPolicy, error) {
	if policy.ID == 0 {
		return nil, nil
	}

	// populate simple properties
	ply := &api_models.ReplicationPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		Filters:           policy.Filters,
		ReplicateDeletion: policy.ReplicateDeletion,
		Trigger:           policy.Trigger,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
	}

	// populate projects
	for _, projectID := range policy.ProjectIDs {
		project, err := projectMgr.Get(projectID)
		if err != nil {
			return nil, err
		}

		ply.Projects = append(ply.Projects, project)
	}

	// populate targets
	for _, targetID := range policy.TargetIDs {
		target, err := dao.GetRepTarget(targetID)
		if err != nil {
			return nil, err
		}
		target.Password = ""
		ply.Targets = append(ply.Targets, target)
	}

	// TODO call the method from replication controller
	_, errJobCount, err := dao.FilterRepJobs(policy.ID, "", "error", nil, nil, 0, 0)
	if err != nil {
		return nil, err
	}
	ply.ErrorJobCount = errJobCount

	return ply, nil
}

func convertToRepPolicy(policy *api_models.ReplicationPolicy) rep_models.ReplicationPolicy {
	if policy == nil {
		return rep_models.ReplicationPolicy{}
	}

	ply := rep_models.ReplicationPolicy{
		ID:                policy.ID,
		Name:              policy.Name,
		Description:       policy.Description,
		Filters:           policy.Filters,
		ReplicateDeletion: policy.ReplicateDeletion,
		Trigger:           policy.Trigger,
		CreationTime:      policy.CreationTime,
		UpdateTime:        policy.UpdateTime,
	}

	for _, project := range policy.Projects {
		ply.ProjectIDs = append(ply.ProjectIDs, project.ProjectID)
		ply.Namespaces = append(ply.Namespaces, project.Name)
	}

	for _, target := range policy.Targets {
		ply.TargetIDs = append(ply.TargetIDs, target.ID)
	}

	return ply
}
