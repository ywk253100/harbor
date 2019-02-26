package api

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/replication/ng"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

var (
	testRegistry = &model.Registry{
		Name: "test1",
		URL:  "https://test.harbor.io",
		Type: "harbor",
		Credential: &model.Credential{
			Type:         model.CredentialTypeBasic,
			AccessKey:    "admin",
			AccessSecret: "Harbor12345",
		},
	}
	testRegistry2 = &model.Registry{
		Name: "test2",
		URL:  "https://test2.harbor.io",
		Type: "harbor",
		Credential: &model.Credential{
			Type:         model.CredentialTypeBasic,
			AccessKey:    "admin",
			AccessSecret: "Harbor12345",
		},
	}
)

type RegistrySuite struct {
	suite.Suite
	testAPI         *testapi
	defaultRegistry *model.Registry
}

func (suite *RegistrySuite) SetupSuite() {
	//assert := assert.New(suite.T())
	require := require.New(suite.T())
	require.Nil(ng.Init())

	suite.testAPI = newHarborAPI()
	r, err := suite.testAPI.RegistryCreate(*admin, testRegistry)
	require.Nil(err)
	suite.defaultRegistry = r

	_, registries, err := dao.ListRegistries()
	if err != nil {
		log.Errorf("==========%v \n", err)
		return
	}
	log.Infof("+++++++++++SuitSetup%v \n", registries)

	CommonAddUser()
}

func (suite *RegistrySuite) TearDownSuite() {
	log.Info("==================================TearDownSuite")
	assert := assert.New(suite.T())
	err := suite.testAPI.RegistryDelete(*admin, suite.defaultRegistry.ID)
	assert.Nil(err)
	log.Info("+++++++++++the registry is deleted \n")

	CommonDelUser()
}

func (suite *RegistrySuite) TestGet() {
	log.Info("**************entering the TestGet case")
	assert := assert.New(suite.T())

	// Get a non-existed registry
	_, code, _ := suite.testAPI.RegistryGet(*admin, 0)
	assert.Equal(http.StatusNotFound, code)

	_, registries, err := dao.ListRegistries()
	if err != nil {
		log.Errorf("==========%v \n", err)
		return
	}
	log.Infof("+++++++++++TestGet %v \n", registries)

	// Get as admin
	retrieved, code, err := suite.testAPI.RegistryGet(*admin, suite.defaultRegistry.ID)
	require := require.New(suite.T())
	require.Nil(err)
	require.Equal(http.StatusOK, code)
	assert.Equal("test1", retrieved.Name)

	// Get as user
	retrieved, code, err = suite.testAPI.RegistryGet(*testUser, suite.defaultRegistry.ID)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal("test1", retrieved.Name)
}

func (suite *RegistrySuite) TestList() {
	assert := assert.New(suite.T())

	// List as admin, should succeed
	registries, err := suite.testAPI.RegistryList(*admin)
	assert.Nil(err)
	assert.Equal(1, len(registries))

	// List as user, should succeed
	registries, err = suite.testAPI.RegistryList(*testUser)
	assert.Nil(err)
	assert.Equal(1, len(registries))
}

func (suite *RegistrySuite) TestPost() {
	log.Info("==================================TestPost")
	assert := assert.New(suite.T())

	// Should conflict when create exited registry
	_, err := suite.testAPI.RegistryCreate(*admin, testRegistry)
	assert.NotNil(err)

	// Create as user, should fail
	_, err = suite.testAPI.RegistryCreate(*testUser, testRegistry2)
	assert.NotNil(err)

	// Create new as admin, should succeed
	r, err := suite.testAPI.RegistryCreate(*admin, testRegistry2)
	assert.Nil(err)
	defer func(id int64) {
		err := suite.testAPI.RegistryDelete(*admin, id)
		assert.Nil(err)
	}(r.ID)
	assert.Equal("test2", r.Name)
}

func (suite *RegistrySuite) TestRegistryPut() {
	assert := assert.New(suite.T())

	// Update as admin, should succeed
	suite.defaultRegistry.Credential.AccessSecret = "NewSecret"
	err := suite.testAPI.RegistryUpdate(*admin, suite.defaultRegistry.ID, suite.defaultRegistry)
	assert.Nil(err)
	updated, code, err := suite.testAPI.RegistryGet(*admin, suite.defaultRegistry.ID)
	assert.Nil(err)
	assert.Equal(http.StatusOK, code)
	assert.Equal("NewSecret", updated.Credential.AccessSecret)

	// Update as user, should fail
	err = suite.testAPI.RegistryUpdate(*testUser, suite.defaultRegistry.ID, suite.defaultRegistry)
	assert.NotNil(err)
}

func (suite *RegistrySuite) TestDelete() {
	log.Info("==================================TestDelete")
	assert := assert.New(suite.T())

	// Delete as user, should fail
	err := suite.testAPI.RegistryDelete(*testUser, suite.defaultRegistry.ID)
	assert.NotNil(err)

	// Delete as admin
	err = suite.testAPI.RegistryDelete(*admin, suite.defaultRegistry.ID)
	assert.Nil(err)
}

func TestRegistrySuite(t *testing.T) {
	log.Info("======================================")
	suite.Run(t, new(RegistrySuite))
}
