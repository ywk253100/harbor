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

package flow

import (
	"testing"

	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type stageTestSuite struct {
	suite.Suite
}

func (s *stageTestSuite) SetupTest() {
}

func (s *stageTestSuite) TestInitialize() {
	factory := &mockFactory{}
	factory.On("AdapterPattern").Return(nil)
	adapter.RegisterFactory(model.RegistryTypeHarbor, factory)

	policy := &model.Policy{
		SrcRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
		DestRegistry: &model.Registry{
			Type: model.RegistryTypeHarbor,
		},
	}
	factory.On("Create", mock.Anything).Return(&mockAdapter{}, nil)
	_, _, err := initialize(policy)
	s.Nil(err)
	factory.AssertExpectations(s.T())
}

func (s *stageTestSuite) TestFetchResources() {
	adapter := &mockAdapter{}
	adapter.On("Info").Return(&model.RegistryInfo{
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeArtifact,
		},
	}, nil)
	adapter.On("FetchArtifacts", mock.Anything).Return([]*model.Resource{
		{},
		{},
	}, nil)
	policy := &model.Policy{}
	resources, err := fetchResources(adapter, policy)
	s.Require().Nil(err)
	s.Len(resources, 2)
	adapter.AssertExpectations(s.T())
}

func (s *stageTestSuite) TestFilterResources() {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Artifacts: []*model.Artifact{
					{
						Tags: []string{"latest"},
					},
				},
			},
			Deleted:  true,
			Override: true,
		},
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/harbor",
				},
				Artifacts: []*model.Artifact{
					{
						Tags: []string{"0.2.0"},
					},
					{
						Tags: []string{"0.3.0"},
					},
				},
			},
			Deleted:  true,
			Override: true,
		},
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/mysql",
				},
				Artifacts: []*model.Artifact{
					{
						Tags: []string{"1.0"},
					},
				},
			},
			Deleted:  true,
			Override: true,
		},
	}
	filters := []*model.Filter{
		{
			Type:  model.FilterTypeResource,
			Value: model.ResourceTypeChart,
		},
		{
			Type:  model.FilterTypeName,
			Value: "library/*",
		},
		{
			Type:  model.FilterTypeName,
			Value: "library/harbor",
		},
		{
			Type:  model.FilterTypeTag,
			Value: "0.2.?",
		},
	}
	res, err := filterResources(resources, filters)
	s.Require().Nil(err)
	s.Len(res, 1)
	s.Equal("library/harbor", res[0].Metadata.Repository.Name)
	s.Equal(1, len(res[0].Metadata.Artifacts))
	s.Equal("0.2.0", res[0].Metadata.Artifacts[0].Tags[0])
}

func (s *stageTestSuite) TestAssembleSourceResources() {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	policy := &model.Policy{
		SrcRegistry: &model.Registry{
			ID: 1,
		},
	}
	res := assembleSourceResources(resources, policy)
	s.Len(res, 1)
	s.Equal(int64(1), res[0].Registry.ID)
}

func (s *stageTestSuite) TestAssembleDestinationResources() {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	policy := &model.Policy{
		DestRegistry:  &model.Registry{},
		DestNamespace: "test",
		Override:      true,
	}
	res := assembleDestinationResources(resources, policy)
	s.Len(res, 1)
	s.Equal(model.ResourceTypeChart, res[0].Type)
	s.Equal("test/hello-world", res[0].Metadata.Repository.Name)
	s.Equal(1, len(res[0].Metadata.Vtags))
	s.Equal("latest", res[0].Metadata.Vtags[0])
}

func (s *stageTestSuite) TestReplaceNamespace() {
	// empty namespace
	repository := "c"
	namespace := ""
	result := replaceNamespace(repository, namespace)
	s.Equal("c", result)
	// repository contains no "/"
	repository = "c"
	namespace = "n"
	result = replaceNamespace(repository, namespace)
	s.Equal("n/c", result)
	// repository contains only one "/"
	repository = "b/c"
	namespace = "n"
	result = replaceNamespace(repository, namespace)
	s.Equal("n/c", result)
	// repository contains more than one "/"
	repository = "a/b/c"
	namespace = "n"
	result = replaceNamespace(repository, namespace)
	s.Equal("n/c", result)
}

func TestStage(t *testing.T) {
	suite.Run(t, &stageTestSuite{})
}
