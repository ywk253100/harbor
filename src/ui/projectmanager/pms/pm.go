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

package pms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
	er "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
)

var transport = &http.Transport{}

// ProjectManager implements projectmanager.ProjecdtManager interface
// base on project management service
type ProjectManager struct {
	endpoint string
	token    string
	client   *http.Client
}

type user struct {
	Email string `json:"email"`
}

type project struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Public           bool              `json:"isPublic"`
	OwnerID          string            `json:"documentOwner"`
	CustomProperties map[string]string `json:"customProperties"`
	Administrators   []*user           `json:"administrators"`
	Developers       []*user           `json:"members"`
	Guests           []*user           `json:"guests"` // TODO the json name needs to be modified according to the API
}

// NewProjectManager returns an instance of ProjectManager
func NewProjectManager(endpoint, token string) *ProjectManager {
	return &ProjectManager{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    token,
		client: &http.Client{
			Transport: transport,
		},
	}
}

// Get ...
func (p *ProjectManager) Get(projectIDOrName interface{}) (*models.Project, error) {
	project, err := p.get(projectIDOrName)
	if err != nil {
		return nil, err
	}
	return convert(project)
}

func (p *ProjectManager) get(projectIDOrName interface{}) (*project, error) {
	m := map[string]string{}
	if id, ok := projectIDOrName.(int64); ok {
		m["customProperties.__harborId"] = strconv.FormatInt(id, 10)
	} else if name, ok := projectIDOrName.(string); ok {
		m["name"] = name
	} else {
		return nil, fmt.Errorf("unsupported type: %v", projectIDOrName)
	}

	projects, err := p.filter(m)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, nil
	}

	if len(projects) != 1 {
		return nil, fmt.Errorf("unexpected size of project list: %d != 1", len(projects))
	}

	return projects[0], nil
}

func (p *ProjectManager) filter(m map[string]string) ([]*project, error) {
	query := ""
	for k, v := range m {
		if len(query) == 0 {
			query += "?"
		} else {
			query += "&"
		}
		query += fmt.Sprintf("$filter=%s eq '%s'", k, v)
	}

	path := "/projects" + query
	data, err := p.send(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	return parse(data)
}

// parse the response of GET /projects?xxx to project list
func parse(b []byte) ([]*project, error) {
	documents := &struct {
		//TotalCount    int64               `json:"totalCount"`
		//DocumentCount int64               `json:"documentCount"`
		Projects map[string]*project `json:"documents"`
	}{}
	if err := json.Unmarshal(b, documents); err != nil {
		return nil, err
	}

	projects := []*project{}
	for link, project := range documents.Projects {
		project.ID = strings.TrimLeft(link, "/projects/")
		projects = append(projects, project)
	}

	return projects, nil
}

func convert(p *project) (*models.Project, error) {
	if p == nil {
		return nil, nil
	}

	project := &models.Project{
		Name: p.Name,
	}
	if p.Public {
		project.Public = 1
	}

	value := p.CustomProperties["__harborId"]
	if len(value) == 0 {
		return nil, fmt.Errorf("property __harborId is null")
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse __harborId %s to int64: %v", value, err)
	}
	project.ProjectID = id

	value = p.CustomProperties["__enableContentTrust"]
	if len(value) != 0 {
		enable, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse __enableContentTrust %s to bool: %v", value, err)
		}
		project.EnableContentTrust = enable
	}

	value = p.CustomProperties["__preventVulnerableImagesFromRunning"]
	if len(value) != 0 {
		prevent, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse __preventVulnerableImagesFromRunning %s to bool: %v", value, err)
		}
		project.PreventVulnerableImagesFromRunning = prevent
	}

	value = p.CustomProperties["__preventVulnerableImagesFromRunningSeverity"]
	if len(value) != 0 {
		project.PreventVulnerableImagesFromRunningSeverity = value
	}

	value = p.CustomProperties["__automaticallyScanImagesOnPush"]
	if len(value) != 0 {
		scan, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse __automaticallyScanImagesOnPush %s to bool: %v", value, err)
		}
		project.AutomaticallyScanImagesOnPush = scan
	}

	return project, nil
}

// IsPublic ...
func (p *ProjectManager) IsPublic(projectIDOrName interface{}) (bool, error) {
	project, err := p.get(projectIDOrName)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, nil
	}

	return project.Public, nil
}

// Exist ...
func (p *ProjectManager) Exist(projectIDOrName interface{}) (bool, error) {
	project, err := p.get(projectIDOrName)
	if err != nil {
		return false, err
	}

	return project != nil, nil
}

// GetRoles ...
// TODO empty this method after implementing security context with auth context
func (p *ProjectManager) GetRoles(username string, projectIDOrName interface{}) ([]int, error) {
	if len(username) == 0 || projectIDOrName == nil {
		return nil, nil
	}

	id, err := p.getIDbyHarborIDOrName(projectIDOrName)
	if err != nil {
		return nil, err
	}

	// get expanded project which contains role info by GET /projects/id?expand=true
	path := fmt.Sprintf("/projects/%s?expand=true", id)
	data, err := p.send(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	pro := &project{}
	if err = json.Unmarshal(data, pro); err != nil {
		return nil, err
	}

	roles := []int{}

	for _, user := range pro.Administrators {
		if user.Email == username {
			roles = append(roles, common.RoleProjectAdmin)
			break
		}
	}

	for _, user := range pro.Developers {
		if user.Email == username {
			roles = append(roles, common.RoleDeveloper)
			break
		}
	}

	for _, user := range pro.Guests {
		if user.Email == username {
			roles = append(roles, common.RoleGuest)
			break
		}
	}

	return roles, nil
}

func (p *ProjectManager) getIDbyHarborIDOrName(projectIDOrName interface{}) (string, error) {
	pro, err := p.get(projectIDOrName)
	if err != nil {
		return "", err
	}

	if pro == nil {
		return "", fmt.Errorf("project %v not found", projectIDOrName)
	}

	return pro.ID, nil
}

// GetPublic ...
func (p *ProjectManager) GetPublic() ([]*models.Project, error) {
	m := map[string]string{
		"isPublic": "true",
	}

	projects, err := p.filter(m)
	if err != nil {
		return nil, err
	}

	list := []*models.Project{}
	for _, p := range projects {
		project, err := convert(p)
		if err != nil {
			return nil, err
		}
		list = append(list, project)
	}

	return list, nil
}

// GetByMember ...
func (p *ProjectManager) GetByMember(username string) ([]*models.Project, error) {
	// TODO add implement
	return nil, nil
}

// Create ...
func (p *ProjectManager) Create(pro *models.Project) (int64, error) {
	proj := &project{
		CustomProperties: make(map[string]string),
	}
	proj.Name = pro.Name
	proj.Public = pro.Public == 1
	proj.CustomProperties["__enableContentTrust"] = strconv.FormatBool(pro.EnableContentTrust)
	proj.CustomProperties["__preventVulnerableImagesFromRunning"] = strconv.FormatBool(pro.PreventVulnerableImagesFromRunning)
	proj.CustomProperties["__preventVulnerableImagesFromRunningSeverity"] = pro.PreventVulnerableImagesFromRunningSeverity
	proj.CustomProperties["__automaticallyScanImagesOnPush"] = strconv.FormatBool(pro.AutomaticallyScanImagesOnPush)

	// TODO remove the logic if Admiral generates the harborId
	proj.CustomProperties["__harborId"] = strconv.FormatInt(time.Now().UnixNano(), 10)

	data, err := json.Marshal(proj)
	if err != nil {
		return 0, err
	}

	b, err := p.send(http.MethodPost, "/projects", bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}

	proj = &project{}
	if err = json.Unmarshal(b, proj); err != nil {
		return 0, err
	}

	pp, err := convert(proj)
	if err != nil {
		return 0, err
	}

	return pp.ProjectID, err
}

// Delete ...
func (p *ProjectManager) Delete(projectIDOrName interface{}) error {
	id, err := p.getIDbyHarborIDOrName(projectIDOrName)
	if err != nil {
		return err
	}

	_, err = p.send(http.MethodDelete, fmt.Sprintf("/projects/%s", id), nil)
	return err
}

// Update ...
func (p *ProjectManager) Update(projectIDOrName interface{}, project *models.Project) error {
	return errors.New("project update is unsupported")
}

// GetAll ...
func (p *ProjectManager) GetAll(query *models.ProjectQueryParam, base ...*models.BaseProjectCollection) ([]*models.Project, error) {
	return nil, errors.New("get all projects is unsupported")
}

// GetTotal ...
func (p *ProjectManager) GetTotal(query *models.ProjectQueryParam, base ...*models.BaseProjectCollection) (int64, error) {
	return 0, errors.New("get total of projects is unsupported")
}

// GetHasReadPerm returns all projects that user has read perm to
// TODO maybe can be removed as search isn't implemented in integration mode
func (p *ProjectManager) GetHasReadPerm(username ...string) ([]*models.Project, error) {
	// TODO add implement
	return nil, nil
}

func (p *ProjectManager) send(method, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, p.endpoint+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("x-xenon-auth-token", p.token)

	url := req.URL.String()

	req.URL.RawQuery = req.URL.Query().Encode()
	resp, err := p.client.Do(req)
	if err != nil {
		log.Debugf("\"%s %s\" failed", req.Method, url)
		return nil, err
	}
	defer resp.Body.Close()
	log.Debugf("\"%s %s\" %d", req.Method, url, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &er.Error{
			StatusCode: resp.StatusCode,
			Detail:     string(b),
		}
	}

	return b, nil
}
