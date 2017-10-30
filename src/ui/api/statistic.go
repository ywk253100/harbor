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

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	// PriPC : count of private projects
	PriPC = "private_project_count"
	// PriRC : count of private repositories
	PriRC = "private_repo_count"
	// PubPC : count of public projects
	PubPC = "public_project_count"
	// PubRC : count of public repositories
	PubRC = "public_repo_count"
	// TPC : total count of projects
	TPC = "total_project_count"
	// TRC : total count of repositories
	TRC = "total_repo_count"
)

// StatisticAPI handles request to /api/statistics/
type StatisticAPI struct {
	BaseController
	username string
}

//Prepare validates the URL and the user
func (s *StatisticAPI) Prepare() {
	s.BaseController.Prepare()
	if !s.SecurityCtx.IsAuthenticated() {
		s.HandleUnauthorized()
		return
	}
	s.username = s.SecurityCtx.GetUsername()
}

// Get total projects and repos of the user
func (s *StatisticAPI) Get() {
	statistic := map[string]int64{}
	pubProjs, err := s.ProjectMgr.GetPublic()
	if err != nil {
		s.ParseAndHandleError("failed to get public projects", err)
		return
	}

	statistic[PubPC] = (int64)(len(pubProjs))

	ids := []int64{}
	for _, p := range pubProjs {
		ids = append(ids, p.ProjectID)
	}
	n, err := dao.GetTotalOfRepositoriesByProject(ids, "")
	if err != nil {
		log.Errorf("failed to get total of public repositories: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "")
	}
	statistic[PubRC] = n

	if s.SecurityCtx.IsSysAdmin() {
		result, err := s.ProjectMgr.List(nil)
		if err != nil {
			log.Errorf("failed to get total of projects: %v", err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[TPC] = result.Total
		statistic[PriPC] = result.Total - statistic[PubPC]

		n, err := dao.GetTotalOfRepositories("")
		if err != nil {
			log.Errorf("failed to get total of repositories: %v", err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[TRC] = n
		statistic[PriRC] = n - statistic[PubRC]
	} else {
		value := false
		result, err := s.ProjectMgr.List(&models.ProjectQueryParam{
			Public: &value,
			Member: &models.MemberQuery{
				Name: s.username,
			},
		})
		if err != nil {
			s.ParseAndHandleError(fmt.Sprintf(
				"failed to get projects of user %s", s.username), err)
			return
		}

		statistic[PriPC] = result.Total

		ids := []int64{}
		for _, p := range result.Projects {
			ids = append(ids, p.ProjectID)
		}

		n, err = dao.GetTotalOfRepositoriesByProject(ids, "")
		if err != nil {
			s.HandleInternalServerError(fmt.Sprintf(
				"failed to get total of repositories for user %s: %v",
				s.username, err))
			return
		}
		statistic[PriRC] = n
	}

	s.Data["json"] = statistic
	s.ServeJSON()
}
