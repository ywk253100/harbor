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

	"github.com/vmware/harbor/src/common"
)

// ProjectMemberAPI handles request to /api/projects/{}/members/{}
type ProjectMemberAPI struct {
	BaseController
	pid int64
	mid int64
}

type memberReq struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	Roles    []int  `json:"roles"`
}

// Prepare validates the URL and parms
func (pma *ProjectMemberAPI) Prepare() {
	pma.BaseController.Prepare()
	pid, err := pma.GetInt64FromPath(":pid")
	if err != nil || pid <= 0 {
		text := "invalid project ID: "
		if err != nil {
			text += err.Error()
		} else {
			text += fmt.Sprintf("%d", pid)
		}
		pma.HandleBadRequest(text)
		return
	}

	if !pma.ProjectMgr.Exist(pid) {
		pma.HandleNotFound(fmt.Sprintf("project %d not found", pid))
		return
	}
	pma.pid = pid

	if len(pma.GetStringFromPath(":mid")) != 0 {
		mid, err := pma.GetInt64FromPath(":mid")
		if err != nil || mid <= 0 {
			text := "invalid member ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", mid)
			}
			pma.HandleBadRequest(text)
			return
		}

		if pma.ProjectMgr.GetMember(pid, mid) == nil {
			pma.HandleNotFound(fmt.Sprintf("member %d not found", mid))
			return
		}
	}

	if !pma.SecurityCtx.IsAuthenticated() {
		pma.HandleUnauthorized()
		return
	}
}

// Get ...
func (pma *ProjectMemberAPI) Get() {
	if !pma.SecurityCtx.HasReadPerm(pma.pid) {
		pma.HandleForbidden(pma.SecurityCtx.GetUsername())
		return
	}

	// list members
	if pma.mid == 0 {
		username := pma.GetString("username")
		members := pma.ProjectMgr.GetMembers(pma.pid, username)
		pma.Data["json"] = members
		pma.ServeJSON()
		return
	}

	// get the member specified by mid
	pma.Data["json"] = pma.ProjectMgr.GetMember(pma.pid, pma.mid)

	pma.ServeJSON()
}

// Post ...
func (pma *ProjectMemberAPI) Post() {
	if !pma.SecurityCtx.HasAllPerm(pma.pid) {
		pma.HandleForbidden(pma.SecurityCtx.GetUsername())
		return
	}

	var req memberReq
	pma.DecodeJSONReq(&req)

	// TODO add validate function to memberReq
	if req.UserID == 0 && len(req.Username) == 0 {
		pma.HandleBadRequest("user_id and username at lease one should not be nil")
		return
	}

	if len(req.Roles) != 1 {
		pma.HandleBadRequest("only one role is supported")
		return
	}

	rid := req.Roles[0]
	if !(rid == common.RoleProjectAdmin ||
		rid == common.RoleDeveloper ||
		rid == common.RoleGuest) {
		pma.CustomAbort(http.StatusBadRequest, "invalid role")
	}

	var userIDOrName interface{}
	if req.UserID != 0 {
		userIDOrName = req.UserID
	} else {
		userIDOrName = req.Username
	}

	if pma.ProjectMgr.MemberExist(pma.pid, userIDOrName) {
		pma.HandleConflict(fmt.Sprintf("user %v is already the member of project %d",
			userIDOrName, pma.pid))
		return
	}

	if err := pma.ProjectMgr.AddMember(pma.pid, userIDOrName, rid); err != nil {
		pma.HandleInternalServerError(fmt.Sprintf("failed to add user %v to project %d",
			userIDOrName, pma.pid))
		return
	}
}

// Put ...
func (pma *ProjectMemberAPI) Put() {
	if !pma.SecurityCtx.HasAllPerm(pma.pid) {
		pma.HandleForbidden(pma.SecurityCtx.GetUsername())
		return
	}

	if pma.mid == 0 {
		pma.HandleBadRequest("mid is needed in path")
		return
	}

	if !pma.ProjectMgr.MemberExist(pma.pid, pma.mid) {
		pma.HandleNotFound(fmt.Sprintf("user %d is not the member of project %d", pma.mid, pma.pid))
		return
	}

	var req memberReq
	pma.DecodeJSONReq(&req)

	if len(req.Roles) != 1 {
		pma.HandleBadRequest("only one role is supported")
		return
	}

	role := req.Roles[0]
	if !(role == common.RoleProjectAdmin ||
		role == common.RoleDeveloper ||
		role == common.RoleGuest) {
		pma.CustomAbort(http.StatusBadRequest, "invalid role")
	}

	if err := pma.ProjectMgr.UpdateMember(pma.pid, pma.mid, role); err != nil {
		pma.HandleInternalServerError(
			fmt.Sprintf("failed to update the member %d of project %d with role %d",
				pma.mid, pma.pid, role))
		return
	}
}

// Delete ...
func (pma *ProjectMemberAPI) Delete() {
	if !pma.SecurityCtx.HasAllPerm(pma.pid) {
		pma.HandleForbidden(pma.SecurityCtx.GetUsername())
		return
	}

	if pma.mid == 0 {
		pma.HandleBadRequest("mid is needed in path")
		return
	}

	if !pma.ProjectMgr.MemberExist(pma.pid, pma.mid) {
		pma.HandleNotFound(fmt.Sprintf("user %d is not the member of project %d", pma.mid, pma.pid))
		return
	}

	if err := pma.ProjectMgr.DeleteMember(pma.pid, pma.mid); err != nil {
		pma.HandleInternalServerError(
			fmt.Sprintf("failed to delete member %d from project %d: %v",
				pma.mid, pma.pid, err))
		return
	}
}
