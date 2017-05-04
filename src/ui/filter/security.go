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
	"strings"

	"github.com/astaxie/beego/context"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/security/rbac"
	"github.com/vmware/harbor/src/common/security/secret"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/projectmanager"
)

const (
	// HarborSecurityContext is the name of security context passed to handlers
	HarborSecurityContext = "harbor_security_context"
	// HarborProjectManager is the name of project manager passed to handlers
	HarborProjectManager = "harbor_project_manager"
)

// SecurityFilter authenticates the request and passes a security context with it
// which can be used to do some authorization
func SecurityFilter(ctx *context.Context) {
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

func fillContext(ctx *context.Context) {
	// secret
	scrt := ctx.GetCookie("secret")
	if len(scrt) != 0 {
		ctx.Input.SetData(HarborProjectManager,
			getProjectManager())

		log.Info("creating a secret security context...")
		ctx.Input.SetData(HarborSecurityContext,
			secret.NewSecurityContext(scrt, config.SecretStore))

		return
	}

	var user *models.User
	var err error

	// session
	username := ctx.Input.Session("username")
	if username != nil {
		user, err = dao.GetUser(models.User{
			Username: username.(string),
		})
		if err != nil {
			log.Errorf("failed to get user %s: %v", username.(string), err)
		}

		if user != nil {
			log.Info("get credential from session")
		}
	}

	// basic auth
	if user == nil {
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
				log.Info("get credential from basic auth")
			}
		}
	}

	if user == nil {
		log.Info("get no credential")
	}

	pm := getProjectManager()
	ctx.Input.SetData(HarborProjectManager, pm)

	log.Info("creating a rbac security context...")
	ctx.Input.SetData(HarborSecurityContext,
		rbac.NewSecurityContext(user, pm))

	return
}

func getProjectManager(token ...string) projectmanager.ProjectManager {
	if len(config.DeployMode()) == 0 ||
		config.DeployMode() == common.DeployModeStandAlone {
		log.Info("filling a project manager based on database...")
		return config.DBProjectManager
	}

	// TODO create project manager based on pms
	log.Info("filling a project manager based on pms...")
	return nil
}
