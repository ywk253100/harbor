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

package orm

import (
	"context"
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/q"
)

// GenerateQuerySetter generates the query setter according to the query
// "Limit" and "offset" will be set if the pagination is true
func GenerateQuerySetter(ctx context.Context, model interface{}, query *q.Query, pagination bool) orm.QuerySeter {
	qs := GetOrmer(ctx).QueryTable(model)
	if query != nil {
		for k, v := range query.Keywords {
			qs = qs.Filter(k, v)
		}
		if pagination {
			if query.PageSize > 0 {
				qs = qs.Limit(query.PageSize)
				if query.PageNumber > 0 {
					qs = qs.Offset(query.PageSize * (query.PageNumber - 1))
				}
			}
		}
	}
	return qs
}

// GetOrmer returns an ormer
// TODO remove it after weiwei's PR merged
func GetOrmer(ctx context.Context) orm.Ormer {
	return dao.GetOrmer()
}
