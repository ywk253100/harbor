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
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/tag"
	tag_model "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"time"
)

var (
	// Ctl is a global artifact controller instance
	Ctl = NewController(artifact.Mgr, tag.Mgr)
)

// Controller defines the operations related with artifacts and tags
type Controller interface {
	// Ensure the artifact specified by the digest exists under the repository,
	// creates it if it doesn't exist. If tags are provided, ensure they exist
	// and are attached to the artifact. If the tags don't exist, create them first
	Ensure(ctx context.Context, repository *models.RepoRecord, digest string, tags ...string) (err error)
	// List artifacts according to the query, specify the properties returned with option
	List(ctx context.Context, query *q.Query, option *Option) (total int64, artifacts []*Artifact, err error)
	// Get the artifact specified by ID
	Get(ctx context.Context, id int64, option *Option) (artifact *Artifact, err error)
	// Delete the artifact specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// ListTags according to the query, specify the properties returned with option
	ListTags(ctx context.Context, query *q.Query, option *TagOption) (total int64, tags []*Tag, err error)
	// DeleteTag deletes the tag specified by tagID
	DeleteTag(ctx context.Context, tagID int64) (err error)
	// UpdatePullTime updates the pull time for the artifact. If the tagID is provides, update the pull
	// time of the tag as well
	UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) (err error)
	// GetSubResource returns the sub resource of the artifact
	// The sub resource is different according to the artifact type:
	// build history for image; values.yaml, readme and dependencies for chart, etc
	GetSubResource(ctx context.Context, artifactID int64, resource string) (*Resource, error)
	// TODO move this to GC controller?
	// Prune removes the useless artifact records. The underlying registry data will
	// be removed during garbage collection
	// Prune(ctx context.Context, option *Option) error
}

// NewController creates an instance of the default artifact controller
func NewController(artMgr artifact.Manager, tagMgr tag.Manager) Controller {
	return &controller{
		artMgr: artMgr,
		tagMgr: tagMgr,
	}
}

// TODO handle concurrency, the redis lock doesn't cover all cases
// TODO As a redis lock is applied during the artifact pushing, we do not to handle the concurrent issues
// for artifacts and tags？？

type controller struct {
	artMgr artifact.Manager
	tagMgr tag.Manager
}

func (c *controller) Ensure(ctx context.Context, repository *models.RepoRecord, digest string, tags ...string) error {
	artifact, err := c.ensureArtifact(ctx, repository, digest)
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if err = c.ensureTag(ctx, artifact, tag); err != nil {
			return err
		}
	}
	return nil
}

// ensure the artifact exists under the repository, create it if doesn't exist.
func (c *controller) ensureArtifact(ctx context.Context, repository *models.RepoRecord, digest string) (*artifact.Artifact, error) {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repository.RepositoryID,
			"digest":        digest,
		},
	}
	_, artifacts, err := c.artMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	// the artifact already exists under the repository, return directly
	if len(artifacts) > 0 {
		return artifacts[0], nil
	}

	// the artifact doesn't exist under the repository, create it first
	artifact := &artifact.Artifact{
		ProjectID:    repository.ProjectID,
		RepositoryID: repository.RepositoryID,
		Digest:       digest,
		PushTime:     time.Now(),
	}
	// abstract the specific information for the artifact
	c.abstract(ctx, repository, digest, artifact)
	// create it
	id, err := c.artMgr.Create(ctx, artifact)
	if err != nil {
		return nil, err
	}
	artifact.ID = id
	return artifact, nil
}

func (c *controller) ensureTag(ctx context.Context, artifact *artifact.Artifact, name string) error {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": artifact.RepositoryID,
			"name":          name,
		},
	}
	_, tags, err := c.tagMgr.List(ctx, query)
	if err != nil {
		return err
	}
	// the tag already exists under the repository
	if len(tags) > 0 {
		tag := tags[0]
		// the tag already exists under the repository and is attached to the artifact, return directly
		if tag.ArtifactID == artifact.ID {
			return nil
		}
		// the tag exists under the repository, but it is attached to other artifact
		// update it to point to the provided artifact
		tag.ArtifactID = artifact.ID
		tag.PushTime = time.Now()
		return c.tagMgr.Update(ctx, tag, "ArtifactID", "PushTime")
	}
	// the tag doesn't exist under the repository, create it
	_, err = c.tagMgr.Create(ctx, &tag_model.Tag{
		RepositoryID: artifact.RepositoryID,
		ArtifactID:   artifact.ID,
		Name:         name,
		PushTime:     time.Now(),
	})
	return err
}

func (c *controller) List(ctx context.Context, query *q.Query, option *Option) (int64, []*Artifact, error) {
	total, arts, err := c.artMgr.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	var artifacts []*Artifact
	for _, art := range arts {
		artifacts = append(artifacts, c.assembleArtifact(ctx, art, option))
	}
	return total, artifacts, nil
}
func (c *controller) Get(ctx context.Context, id int64, option *Option) (*Artifact, error) {
	art, err := c.artMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.assembleArtifact(ctx, art, option), nil
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	// delete artifact first in case the artifact is referenced by other artifact
	if err := c.artMgr.Delete(ctx, id); err != nil {
		return err
	}

	// delete all tags that attached to the artifact
	_, tags, err := c.tagMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"artifact_id": id,
		},
	})
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if err = c.DeleteTag(ctx, tag.ID); err != nil {
			return err
		}
	}
	// TODO fire delete artifact event
	return nil
}
func (c *controller) ListTags(ctx context.Context, query *q.Query, option *TagOption) (int64, []*Tag, error) {
	total, tgs, err := c.tagMgr.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	var tags []*Tag
	for _, tg := range tgs {
		tags = append(tags, c.assembleTag(ctx, tg, option))
	}
	return total, tags, nil
}

func (c *controller) DeleteTag(ctx context.Context, tagID int64) error {
	// immutable checking is covered in middleware
	// TODO check signature
	// TODO delete label
	// TODO fire delete tag event
	return c.tagMgr.Delete(ctx, tagID)
}

func (c *controller) UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) error {
	if err := c.artMgr.UpdatePullTime(ctx, artifactID, time); err != nil {
		return err
	}
	return c.tagMgr.Update(ctx, &tag_model.Tag{
		ID: tagID,
	}, "PullTime")
}
func (c *controller) GetSubResource(ctx context.Context, artifactID int64, resource string) (*Resource, error) {
	// TODO implement
	return nil, nil
}

// assemble several part into a single artifact
func (c *controller) assembleArtifact(ctx context.Context, art *artifact.Artifact, option *Option) *Artifact {
	artifact := &Artifact{
		Artifact: *art,
	}
	if option == nil {
		return artifact
	}
	// populate tags
	if option.WithTag {
		_, tgs, err := c.tagMgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"artifact_id": artifact.ID,
			},
		})
		if err == nil {
			// assemble tags
			for _, tg := range tgs {
				artifact.Tags = append(artifact.Tags, c.assembleTag(ctx, tg, option.TagOption))
			}
		} else {
			log.Errorf("failed to list tag of artifact %d: %v", artifact.ID, err)
		}
	}
	// TODO populate other properties: scan, signature etc.
	return artifact
}

// assemble several part into a single tag
func (c *controller) assembleTag(ctx context.Context, tag *tag_model.Tag, option *TagOption) *Tag {
	t := &Tag{
		Tag: *tag,
	}
	if option == nil {
		return t
	}
	// TODO populate label, signature, immutable status for tag
	return t
}

func (c *controller) abstract(ctx context.Context, repository *models.RepoRecord, digest string, artifact *artifact.Artifact) {
	// TODO abstract the specific info for the artifact
	// handler references
}
