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

package repository

import (
	"context"
	"github.com/goharbor/harbor/src/chartserver"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository/dao"
)

// Mgr is the global repository manager instance
// TODO after the refactor, the params should be removed
var Mgr = New(nil, nil)

// TODO refactor the repository manager

// Manager is used for repository management
// currently, the interface only defines the methods needed for tag retention
// will expand it when doing refactor
type Manager interface {
	// if the repository with the specific name exists, read it, or create it
	GetOrCreate(ctx context.Context, repository *models.RepoRecord) (created bool, id int64, err error)
	// List image repositories under the project specified by the ID
	ListImageRepositories(projectID int64) ([]*models.RepoRecord, error)
	// List chart repositories under the project specified by the ID
	ListChartRepositories(projectID int64) ([]*chartserver.ChartInfo, error)
}

// New returns a default implementation of Manager
func New(projectMgr project.Manager, chartCtl *chartserver.Controller) Manager {
	return &manager{
		dao:        dao.New(),
		projectMgr: projectMgr,
		chartCtl:   chartCtl,
	}
}

type manager struct {
	dao        dao.DAO
	projectMgr project.Manager
	chartCtl   *chartserver.Controller
}

func (m *manager) GetOrCreate(ctx context.Context, repository *models.RepoRecord) (bool, int64, error) {
	return m.dao.ReadOrCreate(ctx, repository)
}

// List image repositories under the project specified by the ID
func (m *manager) ListImageRepositories(projectID int64) ([]*models.RepoRecord, error) {
	return common_dao.GetRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{projectID},
	})
}

// List chart repositories under the project specified by the ID
func (m *manager) ListChartRepositories(projectID int64) ([]*chartserver.ChartInfo, error) {
	project, err := m.projectMgr.Get(projectID)
	if err != nil {
		return nil, err
	}
	return m.chartCtl.ListCharts(project.Name)
}
