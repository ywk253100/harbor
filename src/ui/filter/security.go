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

package filter

import (
	"context"
	"strings"

	beegoctx "github.com/astaxie/beego/context"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/security/rbac"
	"github.com/vmware/harbor/src/common/security/secret"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/projectmanager"
)

type key string

const (
	// HarborSecurityContext is the name of security context passed to handlers
	HarborSecurityContext key = "harbor_security_context"
	// HarborProjectManager is the name of project manager passed to handlers
	HarborProjectManager key = "harbor_project_manager"
)

// SecurityFilter authenticates the request and passes a security context with it
// which can be used to do some authorization
func SecurityFilter(ctx *beegoctx.Context) {
	if ctx == nil {
		return
	}

	req := ctx.Request
	if req == nil {
		return
	}

	if !strings.HasPrefix(req.URL.RequestURI(), "/api/") &&
		!strings.HasPrefix(req.URL.RequestURI(), "/service/") {
		return
	}

	// fill ctx with security context and project manager
	fillContext(ctx)
}

func fillContext(ctx *beegoctx.Context) {
	// secret
	scrt := ctx.GetCookie("secret")
	if len(scrt) != 0 {
		ct := context.WithValue(ctx.Request.Context(),
			HarborProjectManager,
			getProjectManager(ctx))

		log.Info("creating a secret security context...")
		ct = context.WithValue(ct, HarborSecurityContext,
			secret.NewSecurityContext(scrt, config.SecretStore))

		ctx.Request = ctx.Request.WithContext(ct)

		return
	}

	var user *models.User
	var err error

	// basic auth
	username, password, ok := ctx.Request.BasicAuth()
	if ok {
		// TODO the return data contains other params when integrated
		// with vic
		user, err = auth.Login(models.AuthModel{
			Principal: username,
			Password:  password,
		})
		if err != nil {
			log.Errorf("failed to authenticate %s: %v", username, err)
		}
		if user != nil {
			log.Info("got user information via basic auth")
		}
	}

	// session
	if user == nil {
		username := ctx.Input.Session("username")
		isSysAdmin := ctx.Input.Session("isSysAdmin")
		if username != nil {
			user = &models.User{
				Username: username.(string),
			}

			if isSysAdmin != nil && isSysAdmin.(bool) {
				user.HasAdminRole = 1
			}
			log.Info("got user information from session")
		}

		// TODO maybe need to get token from session
	}

	if user == nil {
		log.Info("user information is nil")
	}

	pm := getProjectManager(ctx)
	ct := context.WithValue(ctx.Request.Context(), HarborProjectManager, pm)

	log.Info("creating a rbac security context...")
	ct = context.WithValue(ct, HarborSecurityContext,
		rbac.NewSecurityContext(user, pm))
	ctx.Request = ctx.Request.WithContext(ct)

	return
}

func getProjectManager(ctx *beegoctx.Context) projectmanager.ProjectManager {
	if len(config.DeployMode()) == 0 ||
		config.DeployMode() == common.DeployModeStandAlone {
		log.Info("filling a project manager based on database...")
		return config.DBProjectManager
	}

	// TODO create project manager based on pms
	log.Info("filling a project manager based on pms...")
	return nil
}
