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
	"regexp"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	errutil "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"

	"strconv"
	"time"
)

type deletableResp struct {
	Deletable bool   `json:"deletable"`
	Message   string `json:"message"`
}

// ProjectAPI handles request to /api/projects/{} /api/projects/{}/logs
type ProjectAPI struct {
	BaseController
	project *models.Project
}

const projectNameMaxLen int = 30
const projectNameMinLen int = 2
const restrictedNameChars = `[a-z0-9]+(?:[._-][a-z0-9]+)*`

// Prepare validates the URL and the user
func (p *ProjectAPI) Prepare() {
	p.BaseController.Prepare()
	if len(p.GetStringFromPath(":id")) != 0 {
		id, err := p.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			text := "invalid project ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", id)
			}
			p.HandleBadRequest(text)
			return
		}

		project, err := p.ProjectMgr.Get(id)
		if err != nil {
			p.ParseAndHandleError(fmt.Sprintf("failed to get project %d", id), err)
			return
		}

		if project == nil {
			p.HandleNotFound(fmt.Sprintf("project %d not found", id))
			return
		}

		p.project = project
	}
}

// Post ...
func (p *ProjectAPI) Post() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}
	var onlyAdmin bool
	var err error
	if config.WithAdmiral() {
		onlyAdmin = true
	} else {
		onlyAdmin, err = config.OnlyAdminCreateProject()
		if err != nil {
			log.Errorf("failed to determine whether only admin can create projects: %v", err)
			p.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	if onlyAdmin && !p.SecurityCtx.IsSysAdmin() {
		log.Errorf("Only sys admin can create project")
		p.RenderError(http.StatusForbidden, "Only system admin can create project")
		return
	}
	var pro *models.ProjectRequest
	p.DecodeJSONReq(&pro)
	err = validateProjectReq(pro)
	if err != nil {
		log.Errorf("Invalid project request, error: %v", err)
		p.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	exist, err := p.ProjectMgr.Exists(pro.Name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			pro.Name), err)
		return
	}
	if exist {
		p.RenderError(http.StatusConflict, "")
		return
	}

	if pro.Metadata == nil {
		pro.Metadata = map[string]string{}
	}
	// accept the "public" property to make replication work well with old versions(<=1.2.0)
	if pro.Public != nil && len(pro.Metadata[models.ProMetaPublic]) == 0 {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(*pro.Public == 1)
	}

	// populate public metadata as false if it isn't set
	if _, ok := pro.Metadata[models.ProMetaPublic]; !ok {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(false)
	}

	projectID, err := p.ProjectMgr.Create(&models.Project{
		Name:      pro.Name,
		OwnerName: p.SecurityCtx.GetUsername(),
		Metadata:  pro.Metadata,
	})
	if err != nil {
		if err == errutil.ErrDupProject {
			log.Debugf("conflict %s", pro.Name)
			p.RenderError(http.StatusConflict, "")
		} else {
			p.ParseAndHandleError("failed to add project", err)
		}
		return
	}

	go func() {
		if err = dao.AddAccessLog(
			models.AccessLog{
				Username:  p.SecurityCtx.GetUsername(),
				ProjectID: projectID,
				RepoName:  pro.Name + "/",
				RepoTag:   "N/A",
				Operation: "create",
				OpTime:    time.Now(),
			}); err != nil {
			log.Errorf("failed to add access log: %v", err)
		}
	}()

	p.Redirect(http.StatusCreated, strconv.FormatInt(projectID, 10))
}

// Head ...
func (p *ProjectAPI) Head() {
	name := p.GetString("project_name")
	if len(name) == 0 {
		p.HandleBadRequest("project_name is needed")
		return
	}

	project, err := p.ProjectMgr.Get(name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to get project %s", name), err)
		return
	}

	if project == nil {
		p.HandleNotFound(fmt.Sprintf("project %s not found", name))
		return
	}
}

// Get ...
func (p *ProjectAPI) Get() {
	if !p.project.IsPublic() {
		if !p.SecurityCtx.IsAuthenticated() {
			p.HandleUnauthorized()
			return
		}

		if !p.SecurityCtx.HasReadPerm(p.project.ProjectID) {
			p.HandleForbidden(p.SecurityCtx.GetUsername())
			return
		}
	}

	p.Data["json"] = p.project
	p.ServeJSON()
}

// Delete ...
func (p *ProjectAPI) Delete() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	if !p.SecurityCtx.HasAllPerm(p.project.ProjectID) {
		p.HandleForbidden(p.SecurityCtx.GetUsername())
		return
	}

	result, err := deletable(p.project.ProjectID)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}
	if !result.Deletable {
		p.CustomAbort(http.StatusPreconditionFailed, result.Message)
	}

	if err = p.ProjectMgr.Delete(p.project.ProjectID); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to delete project %d", p.project.ProjectID), err)
		return
	}

	go func() {
		if err := dao.AddAccessLog(models.AccessLog{
			Username:  p.SecurityCtx.GetUsername(),
			ProjectID: p.project.ProjectID,
			RepoName:  p.project.Name + "/",
			RepoTag:   "N/A",
			Operation: "delete",
			OpTime:    time.Now(),
		}); err != nil {
			log.Errorf("failed to add access log: %v", err)
		}
	}()
}

// Deletable ...
func (p *ProjectAPI) Deletable() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	if !p.SecurityCtx.HasAllPerm(p.project.ProjectID) {
		p.HandleForbidden(p.SecurityCtx.GetUsername())
		return
	}

	result, err := deletable(p.project.ProjectID)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}

	p.Data["json"] = result
	p.ServeJSON()
}

func deletable(projectID int64) (*deletableResp, error) {
	count, err := dao.GetTotalOfRepositoriesByProject([]int64{projectID}, "")
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return &deletableResp{
			Deletable: false,
			Message:   "the project contains repositories, can not be deleled",
		}, nil
	}

	policies, err := dao.GetRepPolicyByProject(projectID)
	if err != nil {
		return nil, err
	}

	if len(policies) > 0 {
		return &deletableResp{
			Deletable: false,
			Message:   "the project contains replication rules, can not be deleled",
		}, nil
	}

	return &deletableResp{
		Deletable: true,
	}, nil
}

// List ...
func (p *ProjectAPI) List() {
	// query strings
	page, size := p.GetPaginationParams()
	query := &models.ProjectQueryParam{
		Name:  p.GetString("name"),
		Owner: p.GetString("owner"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	public := p.GetString("public")
	if len(public) > 0 {
		pub, err := strconv.ParseBool(public)
		if err != nil {
			p.HandleBadRequest(fmt.Sprintf("invalid public: %s", public))
			return
		}
		query.Public = &pub
	}

	// standalone, filter projects according to the privilleges of the user first
	if !config.WithAdmiral() {
		var projects []*models.Project
		if !p.SecurityCtx.IsAuthenticated() {
			// not login, only get public projects
			pros, err := p.ProjectMgr.GetPublic()
			if err != nil {
				p.HandleInternalServerError(fmt.Sprintf("failed to get public projects: %v", err))
				return
			}
			projects = []*models.Project{}
			projects = append(projects, pros...)
		} else {
			if !(p.SecurityCtx.IsSysAdmin() || p.SecurityCtx.IsSolutionUser()) {
				projects = []*models.Project{}
				// login, but not system admin or solution user, get public projects and
				// projects that the user is member of
				pros, err := p.ProjectMgr.GetPublic()
				if err != nil {
					p.HandleInternalServerError(fmt.Sprintf("failed to get public projects: %v", err))
					return
				}
				projects = append(projects, pros...)

				mps, err := p.ProjectMgr.List(&models.ProjectQueryParam{
					Member: &models.MemberQuery{
						Name: p.SecurityCtx.GetUsername(),
					},
				})
				if err != nil {
					p.HandleInternalServerError(fmt.Sprintf("failed to list projects: %v", err))
					return
				}
				projects = append(projects, mps.Projects...)
			}
		}
		if projects != nil {
			projectIDs := []int64{}
			for _, project := range projects {
				projectIDs = append(projectIDs, project.ProjectID)
			}
			query.ProjectIDs = projectIDs
		}
	}

	result, err := p.ProjectMgr.List(query)
	if err != nil {
		p.ParseAndHandleError("failed to list projects", err)
		return
	}

	for _, project := range result.Projects {
		if p.SecurityCtx.IsAuthenticated() {
			roles := p.SecurityCtx.GetProjectRoles(project.ProjectID)
			if len(roles) != 0 {
				project.Role = roles[0]
			}

			if project.Role == common.RoleProjectAdmin ||
				p.SecurityCtx.IsSysAdmin() {
				project.Togglable = true
			}
		}

		repos, err := dao.GetRepositoryByProjectName(project.Name)
		if err != nil {
			log.Errorf("failed to get repositories of project %s: %v", project.Name, err)
			p.CustomAbort(http.StatusInternalServerError, "")
		}

		project.RepoCount = len(repos)
	}

	p.SetPaginationHeader(result.Total, page, size)
	p.Data["json"] = result.Projects
	p.ServeJSON()
}

// Put ...
func (p *ProjectAPI) Put() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	if !p.SecurityCtx.HasAllPerm(p.project.ProjectID) {
		p.HandleForbidden(p.SecurityCtx.GetUsername())
		return
	}

	var req *models.ProjectRequest
	p.DecodeJSONReq(&req)

	if err := p.ProjectMgr.Update(p.project.ProjectID,
		&models.Project{
			Metadata: req.Metadata,
		}); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to update project %d",
			p.project.ProjectID), err)
		return
	}
}

// Logs ...
func (p *ProjectAPI) Logs() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.HandleUnauthorized()
		return
	}

	if !p.SecurityCtx.HasReadPerm(p.project.ProjectID) {
		p.HandleForbidden(p.SecurityCtx.GetUsername())
		return
	}

	page, size := p.GetPaginationParams()
	query := &models.LogQueryParam{
		ProjectIDs: []int64{p.project.ProjectID},
		Username:   p.GetString("username"),
		Repository: p.GetString("repository"),
		Tag:        p.GetString("tag"),
		Operations: p.GetStrings("operation"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	timestamp := p.GetString("begin_timestamp")
	if len(timestamp) > 0 {
		t, err := utils.ParseTimeStamp(timestamp)
		if err != nil {
			p.HandleBadRequest(fmt.Sprintf("invalid begin_timestamp: %s", timestamp))
			return
		}
		query.BeginTime = t
	}

	timestamp = p.GetString("end_timestamp")
	if len(timestamp) > 0 {
		t, err := utils.ParseTimeStamp(timestamp)
		if err != nil {
			p.HandleBadRequest(fmt.Sprintf("invalid end_timestamp: %s", timestamp))
			return
		}
		query.EndTime = t
	}

	total, err := dao.GetTotalOfAccessLogs(query)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf(
			"failed to get total of access log: %v", err))
		return
	}

	logs, err := dao.GetAccessLogs(query)
	if err != nil {
		p.HandleInternalServerError(fmt.Sprintf(
			"failed to get access log: %v", err))
		return
	}

	p.SetPaginationHeader(total, page, size)
	p.Data["json"] = logs
	p.ServeJSON()
}

// TODO move this to package models
func validateProjectReq(req *models.ProjectRequest) error {
	pn := req.Name
	if isIllegalLength(req.Name, projectNameMinLen, projectNameMaxLen) {
		return fmt.Errorf("Project name is illegal in length. (greater than 2 or less than 30)")
	}
	validProjectName := regexp.MustCompile(`^` + restrictedNameChars + `$`)
	legal := validProjectName.MatchString(pn)
	if !legal {
		return fmt.Errorf("project name is not in lower case or contains illegal characters")
	}

	metas, err := validateProjectMetadata(req.Metadata)
	if err != nil {
		return err
	}

	req.Metadata = metas
	return nil
}
