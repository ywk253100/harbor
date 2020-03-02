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

/*
import (
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func BuildTagFilters(filters []*model.Filter) (TagFilters, error) {
	var fs TagFilters
	for _, filter := range filters {
		var f TagFilter
		switch filter.Type {
		case model.FilterTypeLabel:
			f = &tagNameFilter{
				pattern: filter.Value.(string),
			}
		}
		if f != nil {
			fs = append(fs, f)
		}
	}
	return fs, nil
}

type TagFilter interface {
	Filter([]*adapter.VTag) ([]*adapter.VTag, error)
}

type TagFilters []TagFilter

func (t TagFilters) Filter(tags []*adapter.VTag) ([]*adapter.VTag, error) {
	var err error
	for _, filter := range t {
		tags, err = filter.Filter(tags)
		if err != nil {
			return nil, err
		}
	}
	return tags, nil
}

type tagNameFilter struct {
	pattern string
}

func (t *tagNameFilter) Filter(tags []*adapter.VTag) ([]*adapter.VTag, error) {
	if len(t.pattern) == 0 {
		return tags, nil
	}
	var result []*adapter.VTag
	for _, tag := range tags {
		match, err := util.Match(t.pattern, tag.Name)
		if err != nil {
			return nil, err
		}
		if match {
			result = append(result, tag)
			continue
		}
	}
	return result, nil
}
*/
