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

package local

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	errutil "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
)

func TestMain(m *testing.M) {
	dbHost := os.Getenv("MYSQL_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable MYSQL_HOST is not set")
	}
	dbPortStr := os.Getenv("MYSQL_PORT")
	if len(dbPortStr) == 0 {
		log.Fatalf("environment variable MYSQL_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("invalid MYSQL_PORT: %v", err)
	}
	dbUser := os.Getenv("MYSQL_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable MYSQL_USR is not set")
	}

	dbPassword := os.Getenv("MYSQL_PWD")
	dbDatabase := os.Getenv("MYSQL_DATABASE")
	if len(dbDatabase) == 0 {
		log.Fatalf("environment variable MYSQL_DATABASE is not set")
	}

	database := &models.Database{
		Type: "mysql",
		MySQL: &models.MySQL{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Database: dbDatabase,
		},
	}

	log.Infof("MYSQL_HOST: %s, MYSQL_USR: %s, MYSQL_PORT: %d, MYSQL_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	os.Exit(m.Run())
}

func TestGet(t *testing.T) {
	pm := &driver{}

	// project name
	project, err := pm.Get("library")
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, "library", project.Name)

	// project ID
	project, err = pm.Get(int64(1))
	assert.Nil(t, err)
	assert.NotNil(t, project)
	assert.Equal(t, int64(1), project.ProjectID)

	// non-exist project
	project, err = pm.Get("non-exist-project")
	assert.Nil(t, err)
	assert.Nil(t, project)

	// invalid type
	project, err = pm.Get(true)
	assert.NotNil(t, err)
}

func TestExist(t *testing.T) {
	pm := &driver{}

	// exist project
	exist, err := pm.Exist("library")
	assert.Nil(t, err)
	assert.True(t, exist)

	// non-exist project
	exist, err = pm.Exist("non-exist-project")
	assert.Nil(t, err)
	assert.False(t, exist)
}

func TestIsPublic(t *testing.T) {
	pms := &driver{}
	// public project
	public, err := pms.IsPublic("library")
	assert.Nil(t, err)
	assert.True(t, public)
	// non exist project
	public, err = pms.IsPublic("non_exist_project")
	assert.Nil(t, err)
	assert.False(t, public)
}

func TestGetPublic(t *testing.T) {
	pm := &driver{}
	projects, err := pm.GetPublic()
	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(projects))

	for _, project := range projects {
		assert.Equal(t, 1, project.Public)
	}
}

func TestCreateAndDelete(t *testing.T) {
	pm := &driver{}

	// nil project
	_, err := pm.Create(nil)
	assert.NotNil(t, err)

	// nil project name
	_, err = pm.Create(&models.Project{
		OwnerID: 1,
	})
	assert.NotNil(t, err)

	// nil owner id and nil owner name
	_, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "non_exist_user",
	})
	assert.NotNil(t, err)

	// valid project, owner id
	id, err := pm.Create(&models.Project{
		Name:    "test",
		OwnerID: 1,
	})
	assert.Nil(t, err)
	assert.Nil(t, pm.Delete(id))

	// valid project, owner name
	id, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "admin",
	})
	assert.Nil(t, err)
	assert.Nil(t, pm.Delete(id))

	// duplicate project name
	id, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "admin",
	})
	assert.Nil(t, err)
	defer pm.Delete(id)
	_, err = pm.Create(&models.Project{
		Name:      "test",
		OwnerName: "admin",
	})
	assert.Equal(t, errutil.ErrDupProject, err)
}

func TestUpdate(t *testing.T) {
	pm := &driver{}

	id, err := pm.Create(&models.Project{
		Name:    "test",
		OwnerID: 1,
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	project, err := pm.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, 0, project.Public)

	project.Public = 1
	assert.Nil(t, pm.Update(id, project))

	project, err = pm.Get(id)
	assert.Nil(t, err)
	assert.Equal(t, 1, project.Public)
}

func TestGetTotal(t *testing.T) {
	pm := &driver{}

	id, err := pm.Create(&models.Project{
		Name:    "get_total_test",
		OwnerID: 1,
		Public:  1,
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	// get by name
	total, err := pm.GetTotal(&models.ProjectQueryParam{
		Name: "get_total_test",
	})
	assert.Nil(t, err)
	assert.Equal(t, int64(1), total)

	// get by owner
	total, err = pm.GetTotal(&models.ProjectQueryParam{
		Owner: "admin",
	})
	assert.Nil(t, err)
	assert.NotEqual(t, 0, total)

	// get by public
	value := true
	total, err = pm.GetTotal(&models.ProjectQueryParam{
		Public: &value,
	})
	assert.Nil(t, err)
	assert.NotEqual(t, 0, total)
}

func TestGetAll(t *testing.T) {
	pm := &driver{}

	id, err := pm.Create(&models.Project{
		Name:    "get_all_test",
		OwnerID: 1,
		Public:  1,
	})
	assert.Nil(t, err)
	defer pm.Delete(id)

	// get by name
	projects, err := pm.GetAll(&models.ProjectQueryParam{
		Name: "get_all_test",
	})
	assert.Nil(t, err)
	assert.Equal(t, id, projects[0].ProjectID)

	// get by owner
	projects, err = pm.GetAll(&models.ProjectQueryParam{
		Owner: "admin",
	})
	assert.Nil(t, err)
	exist := false
	for _, project := range projects {
		if project.ProjectID == id {
			exist = true
			break
		}
	}
	assert.True(t, exist)

	// get by public
	value := true
	projects, err = pm.GetAll(&models.ProjectQueryParam{
		Public: &value,
	})
	assert.Nil(t, err)
	exist = false
	for _, project := range projects {
		if project.ProjectID == id {
			exist = true
			break
		}
	}
	assert.True(t, exist)
}
