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
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/internal/orm"
)

// DAO is the data access object interface for repository
type DAO interface {
	// if the repository with the specific name exists, read it, or create it
	ReadOrCreate(ctx context.Context, repository *models.RepoRecord) (created bool, id int64, err error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) ReadOrCreate(ctx context.Context, repository *models.RepoRecord) (bool, int64, error) {
	return orm.GetOrmer(ctx).ReadOrCreate(repository, "Name")
}
