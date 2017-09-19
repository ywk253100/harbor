// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package pmsdriver

import (
	"github.com/vmware/harbor/src/common/models"
)

// PMSDriver defines the operations that a project manage service driver
// should implement
type PMSDriver interface {
	Get(projectIDOrName interface{}) (*models.Project, error)
	Create(*models.Project) (int64, error)
	Delete(projectIDOrName interface{}) error
	Update(projectIDOrName interface{}, metadata map[string]string) error
	// TODO remove base
	List(query *models.ProjectQueryParam,
		base ...*models.BaseProjectCollection) (*models.ProjectQueryResult, error)
	EnableMetaMgr() bool
}
