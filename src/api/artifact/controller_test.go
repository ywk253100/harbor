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

package artifact

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type controllerTestSuite struct {
	suite.Suite
	require *require.Assertions
	assert  *assert.Assertions
	ctl     *controller
	artMgr  *htesting.FakeArtifactManager
	tagMgr  *htesting.FakeTagManager
}

func (c *controllerTestSuite) SetupSuite() {
	c.require = require.New(c.T())
	c.assert = assert.New(c.T())
}

func (c *controllerTestSuite) SetupTest() {
	c.artMgr = &htesting.FakeArtifactManager{}
	c.tagMgr = &htesting.FakeTagManager{}
	c.ctl = &controller{
		artMgr: c.artMgr,
		tagMgr: c.tagMgr,
	}
}

func (c *controllerTestSuite) TestAssembleTag() {
	tg := &tag.Tag{
		ID:           1,
		RepositoryID: 1,
		ArtifactID:   1,
		Name:         "latest",
		PushTime:     time.Now(),
		PullTime:     time.Now(),
	}
	option := &TagOption{
		WithLabel:           true,
		WithImmutableStatus: true,
	}

	tag := c.ctl.assembleTag(nil, tg, option)
	c.require.NotNil(tag)
	c.assert.Equal(tag.ID, tg.ID)
	// TODO check other fields of option
}

func (c *controllerTestSuite) TestAssembleArtifact() {
	art := &artifact.Artifact{
		ID: 1,
	}
	option := &Option{
		WithTag: true,
		TagOption: &TagOption{
			WithLabel:           false,
			WithImmutableStatus: false,
		},
		WithScanResult: true,
		WithSignature:  true,
	}
	tg := &tag.Tag{
		ID:           1,
		RepositoryID: 1,
		ArtifactID:   1,
		Name:         "latest",
		PushTime:     time.Now(),
		PullTime:     time.Now(),
	}
	c.tagMgr.On("List").Return(1, []*tag.Tag{tg}, nil)
	artifact := c.ctl.assembleArtifact(nil, art, option)
	c.require.NotNil(artifact)
	c.tagMgr.AssertExpectations(c.T())
	c.assert.Equal(art.ID, artifact.ID)
	c.assert.Contains(artifact.Tags, &Tag{Tag: *tg})
	// TODO check other fields of option
}

func (c *controllerTestSuite) TestAbstract() {
	// TODO add test case
}

func (c *controllerTestSuite) TestEnsureArtifact() {
	repository := &models.RepoRecord{
		ProjectID:    1,
		RepositoryID: 1,
	}
	digest := "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180"

	// the artifact already exists
	c.artMgr.On("List").Return(1, []*artifact.Artifact{
		{
			ID: 1,
		},
	}, nil)
	art, err := c.ctl.ensureArtifact(nil, repository, digest)
	c.require.Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.require.NotNil(art)
	c.Equal(int64(1), art.ID)

	// reset the mock
	c.SetupTest()

	// the artifact doesn't exist
	c.artMgr.On("List").Return(1, []*artifact.Artifact{}, nil)
	c.artMgr.On("Create").Return(1, nil)
	art, err = c.ctl.ensureArtifact(nil, repository, digest)
	c.require.Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.require.NotNil(art)
	c.Equal(int64(1), art.ID)
}

func (c *controllerTestSuite) TestEnsureTag() {
	art := &artifact.Artifact{
		ID:           1,
		RepositoryID: 1,
	}
	tg := "latest"

	// the tag already exists under the repository and is attached to the artifact
	c.tagMgr.On("List").Return(1, []*tag.Tag{
		{
			ID:           1,
			RepositoryID: 1,
			ArtifactID:   1,
			Name:         "latest",
		},
	}, nil)
	err := c.ctl.ensureTag(nil, art, tg)
	c.require.Nil(err)
	c.tagMgr.AssertExpectations(c.T())

	// reset the mock
	c.SetupTest()

	// the tag exists under the repository, but it is attached to other artifact
	c.tagMgr.On("List").Return(1, []*tag.Tag{
		{
			ID:           1,
			RepositoryID: 1,
			ArtifactID:   2,
			Name:         "latest",
		},
	}, nil)
	c.tagMgr.On("Update").Return(nil)
	err = c.ctl.ensureTag(nil, art, tg)
	c.require.Nil(err)
	c.tagMgr.AssertExpectations(c.T())

	// reset the mock
	c.SetupTest()

	// the tag doesn't exist under the repository, create it
	c.tagMgr.On("List").Return(1, []*tag.Tag{}, nil)
	c.tagMgr.On("Create").Return(1, nil)
	err = c.ctl.ensureTag(nil, art, tg)
	c.require.Nil(err)
	c.tagMgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestEnsure() {
	repository := &models.RepoRecord{
		ProjectID:    1,
		RepositoryID: 1,
	}
	digest := "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180"
	tg := "latest"

	// both the artifact and the tag don't exist
	c.artMgr.On("List").Return(1, []*artifact.Artifact{}, nil)
	c.artMgr.On("Create").Return(1, nil)
	c.tagMgr.On("List").Return(1, []*tag.Tag{}, nil)
	c.tagMgr.On("Create").Return(1, nil)
	err := c.ctl.Ensure(nil, repository, digest, tg)
	c.require.Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.tagMgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestList() {
	query := &q.Query{}
	option := &Option{
		WithTag:        true,
		WithScanResult: true,
		WithSignature:  true,
	}
	c.artMgr.On("List").Return(1, []*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	c.tagMgr.On("List").Return(1, []*tag.Tag{
		{
			ID:           1,
			RepositoryID: 1,
			ArtifactID:   1,
			Name:         "latest",
		},
	}, nil)
	total, artifacts, err := c.ctl.List(nil, query, option)
	c.require.Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.tagMgr.AssertExpectations(c.T())
	c.assert.Equal(int64(1), total)
	c.require.Len(artifacts, 1)
	c.assert.Equal(int64(1), artifacts[0].ID)
	c.require.Len(artifacts[0].Tags, 1)
	c.assert.Equal(int64(1), artifacts[0].Tags[0].ID)
}

func (c *controllerTestSuite) TestGet() {
	c.artMgr.On("Get").Return(&artifact.Artifact{
		ID:           1,
		RepositoryID: 1,
	}, nil)
	art, err := c.ctl.Get(nil, 1, nil)
	c.require.Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.require.NotNil(art)
	c.assert.Equal(int64(1), art.ID)
}

func (c *controllerTestSuite) TestDelete() {
	c.artMgr.On("Delete").Return(nil)
	c.tagMgr.On("List").Return(0, []*tag.Tag{
		{
			ID: 1,
		},
	}, nil)
	c.tagMgr.On("Delete").Return(nil)
	err := c.ctl.Delete(nil, 1)
	c.require.Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.tagMgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestListTags() {
	c.tagMgr.On("List").Return(1, []*tag.Tag{
		{
			ID:           1,
			RepositoryID: 1,
			ArtifactID:   1,
			Name:         "latest",
		},
	}, nil)
	total, tags, err := c.ctl.ListTags(nil, nil, nil)
	c.require.Nil(err)
	c.assert.Equal(int64(1), total)
	c.assert.Len(tags, 1)
	c.tagMgr.AssertExpectations(c.T())
	// TODO check other properties: label, etc
}

func (c *controllerTestSuite) TestDeleteTag() {
	c.tagMgr.On("Delete").Return(nil)
	err := c.ctl.DeleteTag(nil, 1)
	c.require.Nil(err)
	c.tagMgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestUpdatePullTime() {
	c.artMgr.On("UpdatePullTime").Return(nil)
	c.tagMgr.On("Update").Return(nil)
	err := c.ctl.UpdatePullTime(nil, 1, 1, time.Now())
	c.require.Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.tagMgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestGetSubResource() {
	// TODO
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
