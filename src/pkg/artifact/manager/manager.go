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

package manager

import (
	"context"

	"github.com/goharbor/harbor/src/pkg/artifact/model"
)

// Manager provides the basic management functions for artifacts
type Manager interface {
	// List artifacts according to the query
	List(context.Context, ...*model.Query) (total int64, artifacts []*model.Artifact, err error)
	// Get the specified artifact
	Get(context.Context, int64) (*model.Artifact, error)
	// Create artifact, also the references and tags if any
	Create(context.Context, *model.Artifact) (int64, error)
	// Update the artifact
	Update(context.Context, *model.Artifact) error
	// Delete the artifact specified by the ID, will deletes the references and tags if any as well
	Delete(context.Context, int64) error
	// Attach the tag to the artifact
	Attach(context.Context, *model.Artifact, *model.Tag) error
	// UpdateTag updates the tag
	UpdateTag(ctx context.Context, tag *model.Tag) error
	// DeleteTag deletes the specified tag
	DeleteTag(ctx context.Context, tagID int64) error
	// Prune removes the useless artifact records. The underlying registry data will
	// be removed during garbage collection
	Prune(context.Context, *model.Option) error
}

// NewManager returns an instance of the default artifact manager
func NewManager() Manager {
	// TODO implement
	return nil
}
