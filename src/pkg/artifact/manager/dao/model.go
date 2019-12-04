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
	"encoding/json"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact/model"
)

func init() {
	orm.RegisterModel(&Artifact{})
	orm.RegisterModel(&Tag{})
	orm.RegisterModel(&ArtifactReference{})
}

// Artifact model in database
type Artifact struct {
	ID           int64     `orm:"pk;auto;column(id)"`
	Type         string    `orm:"column(type)"`       // image or chart
	ProjectID    int64     `orm:"column(project_id)"` // needed for quota
	RepositoryID int64     `orm:"column(repository_id)"`
	MediaType    string    `orm:"column(media_type)"` // the specific media type
	Digest       string    `orm:"column(digest)"`
	Size         int64     `orm:"column(size)"`
	UploadTime   time.Time `orm:"column(upload_time)"`
	ExtraAttrs   string    `orm:"column(extra_attrs)"` // json string
	Annotations  string    `orm:"column(annotations)"` // json string
}

// TableName for artifact
func (a *Artifact) TableName() string {
	// TODO use "artifact" after finishing the upgrade/migration work
	return "artifact_2"
}

// To converts the artifact to the business level object
func (a *Artifact) To() *model.Artifact {
	artifact := &model.Artifact{
		ID:   a.ID,
		Type: a.Type,
		Repository: &models.RepoRecord{
			ProjectID:    a.ProjectID,
			RepositoryID: a.RepositoryID,
		},
		MediaType:   a.MediaType,
		Digest:      a.Digest,
		Size:        a.Size,
		UploadTime:  a.UploadTime,
		ExtraAttrs:  map[string]interface{}{},
		Annotations: map[string]string{},
	}
	if len(a.ExtraAttrs) > 0 {
		if err := json.Unmarshal([]byte(a.ExtraAttrs), &artifact.ExtraAttrs); err != nil {
			log.Errorf("failed to unmarshal the extra attrs of artifact %d: %v", a.ID, err)
		}
	}
	if len(a.Annotations) > 0 {
		if err := json.Unmarshal([]byte(a.Annotations), &artifact.Annotations); err != nil {
			log.Errorf("failed to unmarshal the annotations of artifact %d: %v", a.ID, err)
		}
	}
	return artifact
}

// From converts the artifact to the database level object
func (a *Artifact) From(artifact *model.Artifact) {
	a.ID = artifact.ID
	a.Type = artifact.Type
	a.ProjectID = artifact.Repository.ProjectID
	a.RepositoryID = artifact.Repository.RepositoryID
	a.MediaType = artifact.MediaType
	a.Digest = artifact.Digest
	a.Size = artifact.Size
	a.UploadTime = artifact.UploadTime
	if len(artifact.ExtraAttrs) > 0 {
		attrs, err := json.Marshal(artifact.ExtraAttrs)
		if err != nil {
			log.Errorf("failed to marshal the extra attrs of artifact %d: %v", artifact.ID, err)
		}
		a.ExtraAttrs = string(attrs)
	}
	if len(artifact.Annotations) > 0 {
		annotations, err := json.Marshal(artifact.Annotations)
		if err != nil {
			log.Errorf("failed to marshal the annotations of artifact %d: %v", artifact.ID, err)
		}
		a.Annotations = string(annotations)
	}
}

// Tag model in database
type Tag struct {
	ID                 int64     `orm:"pk;auto;column(id)"`
	RepositoryID       int64     `orm:"column(repository_id)"` // tags are the resources of repository, one repository only contains one same name tag
	ArtifactID         int64     `orm:"column(artifact_id)"`   // the artifact ID that the tag attaches to, it changes when pushing a same name but different digest artifact
	Name               string    `orm:"column(name)"`
	UploadTime         time.Time `orm:"column(upload_time)"`
	LatestDownloadTime time.Time `orm:"column(latest_download_time)"`
}

// TableName for tag
func (t *Tag) TableName() string {
	return "tag"
}

// To converts the tag to the business level model
func (t *Tag) To() *model.Tag {
	return &model.Tag{
		ID:                 t.ID,
		Name:               t.Name,
		UploadTime:         t.UploadTime,
		LatestDownloadTime: t.LatestDownloadTime,
	}
}

// From converts the tag to the database level object
func (t *Tag) From(tag *model.Tag) {
	t.ID = tag.ID
	t.Name = tag.Name
	t.UploadTime = tag.UploadTime
	t.LatestDownloadTime = tag.LatestDownloadTime
}

// ArtifactReference records the child artifact referenced by parent artifact
type ArtifactReference struct {
	ID          int64  `orm:"pk;auto;column(id)"`
	ArtifactID  int64  `orm:"column(artifact_id)"`
	ReferenceID int64  `orm:"column(reference_id)"`
	Platform    string `orm:"column(platform)"` // json string
}

// TableName for artifact reference
func (a *ArtifactReference) TableName() string {
	return "artifact_reference"
}
