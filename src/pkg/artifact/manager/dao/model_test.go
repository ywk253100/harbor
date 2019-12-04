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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/artifact/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type modelTestSuite struct {
	suite.Suite
}

func (m *modelTestSuite) TestArtifactTo() {
	t := m.T()
	dbArt := &Artifact{
		ID:           1,
		Type:         "IMAGE",
		ProjectID:    1,
		RepositoryID: 1,
		MediaType:    "application/vnd.oci.image.manifest.v1+json",
		Digest:       "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:         1024,
		UploadTime:   time.Now(),
		ExtraAttrs:   `{"attr1":"value1"}`,
		Annotations:  `{"anno1":"value1"}`,
	}
	art := dbArt.To()
	assert.Equal(t, dbArt.ID, art.ID)
	assert.Equal(t, dbArt.Type, art.Type)
	assert.Equal(t, dbArt.ProjectID, art.Repository.ProjectID)
	assert.Equal(t, dbArt.RepositoryID, art.Repository.RepositoryID)
	assert.Equal(t, dbArt.MediaType, art.MediaType)
	assert.Equal(t, dbArt.Digest, art.Digest)
	assert.Equal(t, dbArt.Size, art.Size)
	assert.Equal(t, dbArt.UploadTime, art.UploadTime)
	assert.Equal(t, "value1", art.ExtraAttrs["attr1"].(string))
	assert.Equal(t, "value1", art.Annotations["anno1"])
}

func (m *modelTestSuite) TestArtifactFrom() {
	t := m.T()
	dbArt := &Artifact{}
	art := &model.Artifact{
		ID:   1,
		Type: "IMAGE",
		Repository: &models.RepoRecord{
			ProjectID:    1,
			RepositoryID: 1,
		},
		MediaType:  "application/vnd.oci.image.manifest.v1+json",
		Digest:     "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:       1024,
		UploadTime: time.Now(),
		ExtraAttrs: map[string]interface{}{
			"attr1": "value1",
		},
		Annotations: map[string]string{
			"anno1": "value1",
		},
	}
	dbArt.From(art)
	assert.Equal(t, art.ID, dbArt.ID)
	assert.Equal(t, art.Type, dbArt.Type)
	assert.Equal(t, art.Repository.ProjectID, dbArt.ProjectID)
	assert.Equal(t, art.Repository.RepositoryID, dbArt.RepositoryID)
	assert.Equal(t, art.MediaType, dbArt.MediaType)
	assert.Equal(t, art.Digest, dbArt.Digest)
	assert.Equal(t, art.Size, dbArt.Size)
	assert.Equal(t, art.UploadTime, dbArt.UploadTime)
	assert.Equal(t, art.ExtraAttrs["attr1"].(string), "value1")
	assert.Equal(t, art.Annotations["anno1"], "value1")
}

func (m *modelTestSuite) TestTagTo() {
	t := m.T()
	dbTag := &Tag{
		ID:                 1,
		Name:               "1.0",
		UploadTime:         time.Now(),
		LatestDownloadTime: time.Now(),
	}
	tag := dbTag.To()
	assert.Equal(t, dbTag.ID, tag.ID)
	assert.Equal(t, dbTag.Name, tag.Name)
	assert.Equal(t, dbTag.UploadTime, tag.UploadTime)
	assert.Equal(t, dbTag.LatestDownloadTime, tag.LatestDownloadTime)
}

func (m *modelTestSuite) TestTagFrom() {
	t := m.T()
	dbTag := &Tag{}
	tag := &model.Tag{
		ID:                 1,
		Name:               "1.0",
		UploadTime:         time.Now(),
		LatestDownloadTime: time.Now(),
	}
	dbTag.From(tag)
	assert.Equal(t, tag.ID, dbTag.ID)
	assert.Equal(t, tag.Name, dbTag.Name)
	assert.Equal(t, tag.UploadTime, dbTag.UploadTime)
	assert.Equal(t, tag.LatestDownloadTime, dbTag.LatestDownloadTime)
}

func TestModel(t *testing.T) {
	suite.Run(t, &modelTestSuite{})
}
