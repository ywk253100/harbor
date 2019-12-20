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

package dao

import (
	common_dao "github.com/goharbor/harbor/src/common/dao"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

var (
	repositoryID int64 = 1000
	artifactID   int64 = 1000
	name               = "latest"
)

type daoTestSuite struct {
	suite.Suite
	require *require.Assertions
	assert  *assert.Assertions
	dao     DAO
	tagID   int64
}

func (d *daoTestSuite) SetupSuite() {
	d.dao = New()
	d.require = require.New(d.T())
	d.assert = assert.New(d.T())
	common_dao.PrepareTestForPostgresSQL()
}

func (d *daoTestSuite) SetupTest() {
	tag := &tag.Tag{
		RepositoryID: repositoryID,
		ArtifactID:   artifactID,
		Name:         name,
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	id, err := d.dao.Create(nil, tag)
	d.require.Nil(err)
	d.tagID = id
}

func (d *daoTestSuite) TearDownTest() {
	err := d.dao.Delete(nil, d.tagID)
	d.require.Nil(err)
}

func (d *daoTestSuite) TestGetTotal() {
	// nil query
	total, err := d.dao.GetTotal(nil, nil)
	d.require.Nil(err)
	d.assert.True(total > 0)
	// query by repository ID and name
	total, err = d.dao.GetTotal(nil, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"name":          name,
		},
	})
	d.require.Nil(err)
	d.assert.Equal(int64(1), total)
}

func (d *daoTestSuite) TestList() {
	// nil query
	tags, err := d.dao.List(nil, nil)
	d.require.Nil(err)
	found := false
	for _, tag := range tags {
		if tag.ID == d.tagID {
			found = true
			break
		}
	}
	d.assert.True(found)

	// query by repository ID and name
	tags, err = d.dao.List(nil, &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"name":          name,
		},
	})
	d.require.Nil(err)
	d.require.Equal(1, len(tags))
	d.assert.Equal(d.tagID, tags[0].ID)
}

func (d *daoTestSuite) TestGet() {
	// get the non-exist tag
	_, err := d.dao.Get(nil, 10000)
	d.require.NotNil(err)
	d.assert.True(ierror.IsErr(err, ierror.NotFoundCode))

	// get the exist tag
	tag, err := d.dao.Get(nil, d.tagID)
	d.require.Nil(err)
	d.require.NotNil(tag)
	d.assert.Equal(d.tagID, tag.ID)
}

func (d *daoTestSuite) TestCreate() {
	// the happy pass case is covered in Setup

	// conflict
	tag := &tag.Tag{
		RepositoryID: repositoryID,
		ArtifactID:   artifactID,
		Name:         name,
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	_, err := d.dao.Create(nil, tag)
	d.require.NotNil(err)
	d.assert.True(ierror.IsErr(err, ierror.ConflictCode))
}

// Delete is covered in TearDown

func (d *daoTestSuite) TestUpdate() {
	// pass
	err := d.dao.Update(nil, &tag.Tag{
		ID:         d.tagID,
		ArtifactID: 2,
	}, "ArtifactID")
	d.require.Nil(err)

	tag, err := d.dao.Get(nil, d.tagID)
	d.require.Nil(err)
	d.require.NotNil(tag)
	d.assert.Equal(int64(2), tag.ArtifactID)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &daoTestSuite{})
}
