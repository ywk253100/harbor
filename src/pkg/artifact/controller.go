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
	"github.com/goharbor/harbor/src/pkg/artifact/manager"
	"github.com/goharbor/harbor/src/pkg/artifact/model"
	"time"
)

var (
	// Ctl is a global Controller instance
	Ctl = NewController()
)

// Controller is the only interface of artifact module to provide the
// management functions for artifacts
type Controller interface {
	// List artifacts according  to the query
	List(context.Context, ...*model.Query) (total int64, artifacts []*model.Artifact, err error)
	// Get the artifact specified by the ID
	Get(context.Context, int64) (*model.Artifact, error)
	// Create the artifact specified by the digest under the repository,
	// returns the artifact ID and error. If tags are provided, attach them
	// to the artifact.
	Create(ctx context.Context, repository *models.RepoRecord, digest string, tags ...string) (id int64, err error)
	// Delete just deletes the artifact record. The underlying data of registry will be
	// removed during garbage collection
	Delete(ctx context.Context, id int64) error
	// GetAdditionalResource returns the additional resource of the artifact
	// The additional resource is different according to the artifact type:
	// build history for image; values.yaml, readme and dependencies for chart, etc
	GetAdditionalResource(ctx context.Context, artifactID int64, resource string) ([]byte, error)
	// Attach the tags to the artifact. If the tag doesn't exist, will create it.
	// If the tag is already attached to another artifact, will detach it first
	Attach(ctx context.Context, artifact *model.Artifact, tags ...string) error
	// UpdateTag updates the tag
	UpdateTag(ctx context.Context, tag *model.Tag) error
	// DeleteTag deletes the specified tag
	DeleteTag(ctx context.Context, tagID int64) error
	// Prune removes the useless artifact records. The underlying registry data will
	// be removed during garbage collection
	Prune(context.Context, *model.Option) error
}

// NewController returns an instance of the default controller
func NewController() Controller {
	return &controller{
		mgr: manager.NewManager(),
	}
}

var _ Controller = &controller{}

type controller struct {
	mgr manager.Manager
}

func (c *controller) List(ctx context.Context, query ...*model.Query) (total int64, artifacts []*model.Artifact, err error) {
	return c.mgr.List(ctx, query...)
}
func (c *controller) Get(ctx context.Context, id int64) (*model.Artifact, error) {
	return c.mgr.Get(ctx, id)
}
func (c *controller) Create(ctx context.Context, repository *models.RepoRecord, digest string, tags ...string) (id int64, err error) {
	// TODO implement
	// get the manifest specified by digest from cache or registry
	// abstract the information
	// save it
	return 0, nil
}
func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.mgr.Delete(ctx, id)
}
func (c *controller) GetAdditionalResource(ctx context.Context, artifactID int64, resource string) ([]byte, error) {
	// TODO implement
	return nil, nil
}
func (c *controller) Attach(ctx context.Context, artifact *model.Artifact, tags ...string) error {
	now := time.Now()
	for _, name := range tags {
		tag := &model.Tag{
			Name:       name,
			UploadTime: now,
		}
		if err := c.mgr.Attach(ctx, artifact, tag); err != nil {
			return err
		}
	}
	return nil
}
func (c *controller) UpdateTag(ctx context.Context, tag *model.Tag) error {
	return c.mgr.UpdateTag(ctx, tag)
}
func (c *controller) DeleteTag(ctx context.Context, tagID int64) error {
	return c.mgr.DeleteTag(ctx, tagID)
}
func (c *controller) Prune(ctx context.Context, option *model.Option) error {
	return c.mgr.Prune(ctx, option)
}
