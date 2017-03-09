/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"net/http"
	"sort"
	"strings"

	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

// SearchAPI handles requesst to /api/search
type SearchAPI struct {
	api.BaseAPI
}

type searchResult struct {
	Project    []models.Project         `json:"project"`
	Repository []map[string]interface{} `json:"repository"`
}

// Get ...
func (s *SearchAPI) Get() {
	userID, _, ok := s.GetUserIDForRequest()
	if !ok {
		userID = dao.NonExistUserID
	}

	keyword := s.GetString("q")

	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("failed to check whether the user %d is system admin: %v", userID, err)
		s.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	var projects []models.Project

	if isSysAdmin {
		projects, err = dao.GetProjects("")
		if err != nil {
			log.Errorf("failed to get all projects: %v", err)
			s.CustomAbort(http.StatusInternalServerError, "internal error")
		}
	} else {
		projects, err = dao.SearchProjects(userID)
		if err != nil {
			log.Errorf("failed to get user %d 's relevant projects: %v", userID, err)
			s.CustomAbort(http.StatusInternalServerError, "internal error")
		}
	}

	projectSorter := &models.ProjectSorter{Projects: projects}
	sort.Sort(projectSorter)
	projectResult := []models.Project{}
	for _, p := range projects {
		if len(keyword) > 0 && !strings.Contains(p.Name, keyword) {
			continue
		}

		if userID != dao.NonExistUserID {
			if isSysAdmin {
				p.Role = models.PROJECTADMIN
			} else {
				roles, err := dao.GetUserProjectRoles(userID, p.ProjectID)
				if err != nil {
					log.Errorf("failed to get user's project role: %v", err)
					s.CustomAbort(http.StatusInternalServerError, "")
				}
				p.Role = roles[0].RoleID
			}

			if p.Role == models.PROJECTADMIN {
				p.Togglable = true
			}
		}

		repos, err := dao.GetRepositoryByProjectName(p.Name)
		if err != nil {
			log.Errorf("failed to get repositories of project %s: %v", p.Name, err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}

		p.RepoCount = len(repos)

		projectResult = append(projectResult, p)
	}

	repositoryResult, err := filterRepositories(projects, keyword)
	if err != nil {
		log.Errorf("failed to filter repositories: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "")
	}

	result := &searchResult{Project: projectResult, Repository: repositoryResult}
	s.Data["json"] = result
	s.ServeJSON()
}

func filterRepositories(projects []models.Project, keyword string) (
	[]map[string]interface{}, error) {

	repositories, err := dao.GetAllRepositories()
	if err != nil {
		return nil, err
	}

	i, j := 0, 0
	result := []map[string]interface{}{}
	for i < len(repositories) && j < len(projects) {
		r := repositories[i]
		p, _ := utils.ParseRepository(r.Name)
		d := strings.Compare(p, projects[j].Name)
		if d < 0 {
			i++
			continue
		} else if d == 0 {
			i++
			if len(keyword) != 0 && !strings.Contains(r.Name, keyword) {
				continue
			}
			entry := make(map[string]interface{})
			entry["repository_name"] = r.Name
			entry["project_name"] = projects[j].Name
			entry["project_id"] = projects[j].ProjectID
			entry["project_public"] = projects[j].Public
			entry["pull_count"] = r.PullCount

			tags, err := getTags(r.Name)
			if err != nil {
				return nil, err
			}
			entry["tags_count"] = len(tags)

			result = append(result, entry)
		} else {
			j++
		}
	}
	return result, nil
}

func getTags(repository string) ([]string, error) {
	url, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	client, err := NewRepositoryClient(url, true,
		"admin", repository, "repository", repository, "pull")
	if err != nil {
		return nil, err
	}

	tags, err := listTag(client)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
