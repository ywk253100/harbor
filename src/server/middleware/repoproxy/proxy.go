//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"github.com/go-openapi/errors"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/registry"
	serror "github.com/goharbor/harbor/src/server/error"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/middleware"
)

// BlobGetMiddleware handle get blob request
func BlobGetMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		log.Debugf("Request url is %v", r.URL)
		urlStr := r.URL.String()
		log.Debugf("getting blob with url: %v\n", urlStr)
		ctx := r.Context()
		art := lib.GetArtifactInfo(ctx)
		p, err := project.Ctl.GetByName(ctx, art.ProjectName, project.Metadata(false))
		if err != nil {
			serror.SendError(w, err)
			return
		}

		proxyCtl, err := proxy.ControllerInstance()
		if err != nil {
			serror.SendError(w, err)
			return
		}

		if canProxy(p) == false || proxyCtl.UseLocal(ctx, art.Digest) {
			next.ServeHTTP(w, r)
			return
		}
		err = proxyCtl.ProxyBlob(ctx, p, art, w)
		if err != nil {
			serror.SendError(w, err)
			return
		}
		return
	})
}

// ManifestGetMiddleware middleware handle request for get blob request
func ManifestGetMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		ctx := r.Context()
		art := lib.GetArtifactInfo(ctx)
		p, err := project.Ctl.GetByName(ctx, art.ProjectName)
		if err != nil {
			serror.SendError(w, err)
			return
		}
		proxyCtl, err := proxy.ControllerInstance()
		if err != nil {
			serror.SendError(w, err)
			return
		}

		if !canProxy(p) || proxyCtl.UseLocal(ctx, art.Digest) {
			next.ServeHTTP(w, r)
			return
		}

		if !validProxyRepo(art.Repository) {
			serror.SendError(w, errors.NotFound("repository %v not found", art.Repository))
			return
		}

		log.Debugf("the digest is %v", string(art.Digest))
		err = proxyCtl.ProxyManifest(ctx, p, art, w)
		if err != nil {
			serror.SendError(w, err)
			return
		}
	})
}

func canProxy(p *models.Project) bool {
	if p.RegistryID < 1 {
		return false
	}
	reg, err := registry.NewDefaultManager().Get(p.RegistryID)
	if err != nil {
		log.Errorf("failed to get registry, error:%v", err)
		return false
	}
	return reg.Status == model.Healthy
}

func validProxyRepo(repo string) bool {
	return strings.Count(repo, "/") > 1
}
