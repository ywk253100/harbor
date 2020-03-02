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

package filter

import (
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

// TODO upgrade replication policy: filter type, resource

func BuildRepositoryFilters(filters []*model.Filter) (RepositoryFilters, error) {
	var fs RepositoryFilters
	for _, filter := range filters {
		var f RepositoryFilter
		switch filter.Type {
		case model.FilterTypeName:
			f = &repositoryNameFilter{
				pattern: filter.Value.(string),
			}
		case model.FilterTypeResource:
			f = &repositoryTypeFilter{
				types: []string{
					// TODO upgrade
					filter.Value.(string),
				},
			}
		}
		if f != nil {
			fs = append(fs, f)
		}
	}
	return fs, nil
}

type RepositoryFilter interface {
	Filter([]*model.Repository) ([]*model.Repository, error)
}

type RepositoryFilters []RepositoryFilter

func (r RepositoryFilters) Filter(repositories []*model.Repository) ([]*model.Repository, error) {
	var err error
	for _, filter := range r {
		repositories, err = filter.Filter(repositories)
		if err != nil {
			return nil, err
		}
	}
	return repositories, nil
}

type repositoryNameFilter struct {
	pattern string
}

func (r *repositoryNameFilter) Filter(repositories []*model.Repository) ([]*model.Repository, error) {
	if len(r.pattern) == 0 {
		return repositories, nil
	}
	var result []*model.Repository
	for _, repository := range repositories {
		match, err := util.Match(r.pattern, repository.Name)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, repository)
			continue
		}
	}
	return result, nil
}

type repositoryTypeFilter struct {
	types []string
}

func (r *repositoryTypeFilter) Filter(repositories []*model.Repository) ([]*model.Repository, error) {
	if len(r.types) == 0 {
		return repositories, nil
	}
	var result []*model.Repository
	for _, repository := range repositories {
		for _, t := range r.types {
			if t == repository.Type {
				result = append(result, repository)
				continue
			}
		}
	}
	return result, nil
}
