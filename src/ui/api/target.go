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

package api

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	registry_error "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/ui/config"
)

// TargetAPI handles request to /api/targets/ping /api/targets/{}
type TargetAPI struct {
	BaseController
	secretKey string
}

// Prepare validates the user
func (t *TargetAPI) Prepare() {
	t.BaseController.Prepare()
	if !t.SecurityCtx.IsAuthenticated() {
		t.HandleUnauthorized()
		return
	}

	if !t.SecurityCtx.IsSysAdmin() {
		t.HandleForbidden(t.SecurityCtx.GetUsername())
		return
	}

	var err error
	t.secretKey, err = config.SecretKey()
	if err != nil {
		log.Errorf("failed to get secret key: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func (t *TargetAPI) ping(endpoint, username, password string, insecure bool) {
	registry, err := newRegistryClient(endpoint, insecure, username, password)
	if err != nil {
		// timeout, dns resolve error, connection refused, etc.
		if urlErr, ok := err.(*url.Error); ok {
			if netErr, ok := urlErr.Err.(net.Error); ok {
				t.CustomAbort(http.StatusBadRequest, netErr.Error())
			}

			t.CustomAbort(http.StatusBadRequest, urlErr.Error())
		}

		log.Errorf("failed to create registry client: %#v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if err = registry.Ping(); err != nil {
		if regErr, ok := err.(*registry_error.HTTPError); ok {
			t.CustomAbort(regErr.StatusCode, regErr.Detail)
		}

		log.Errorf("failed to ping registry %s: %v", registry.Endpoint.String(), err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Ping validates whether the target is reachable and whether the credential is valid
func (t *TargetAPI) Ping() {
	req := struct {
		ID       *int64  `json:"id"`
		Endpoint *string `json:"endpoint"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Insecure *bool   `json:"insecure"`
	}{}
	t.DecodeJSONReq(&req)

	target := &models.RepTarget{}
	if req.ID != nil {
		var err error
		target, err = dao.GetRepTarget(*req.ID)
		if err != nil {
			t.HandleInternalServerError(fmt.Sprintf("failed to get target %d: %v", *req.ID, err))
			return
		}
		if target == nil {
			t.HandleNotFound(fmt.Sprintf("target %d not found", *req.ID))
			return
		}
		if len(target.Password) != 0 {
			target.Password, err = utils.ReversibleDecrypt(target.Password, t.secretKey)
			if err != nil {
				t.HandleInternalServerError(fmt.Sprintf("failed to decrypt password: %v", err))
				return
			}
		}
	}

	if req.Endpoint != nil {
		target.URL = *req.Endpoint
	}
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Password != nil {
		target.Password = *req.Password
	}
	if req.Insecure != nil {
		target.Insecure = *req.Insecure
	}

	if len(target.URL) == 0 {
		t.HandleBadRequest("empty endpoint")
		return
	}

	t.ping(target.URL, target.Username, target.Password, target.Insecure)
}

// Get ...
func (t *TargetAPI) Get() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	target.Password = ""

	t.Data["json"] = target
	t.ServeJSON()
}

// List ...
func (t *TargetAPI) List() {
	name := t.GetString("name")
	targets, err := dao.FilterRepTargets(name)
	if err != nil {
		log.Errorf("failed to filter targets %s: %v", name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	for _, target := range targets {
		target.Password = ""
	}

	t.Data["json"] = targets
	t.ServeJSON()
	return
}

// Post ...
func (t *TargetAPI) Post() {
	target := &models.RepTarget{}
	t.DecodeJSONReqAndValidate(target)

	ta, err := dao.GetRepTargetByName(target.Name)
	if err != nil {
		log.Errorf("failed to get target %s: %v", target.Name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if ta != nil {
		t.CustomAbort(http.StatusConflict, "name is already used")
	}

	ta, err = dao.GetRepTargetByEndpoint(target.URL)
	if err != nil {
		log.Errorf("failed to get target [ %s ]: %v", target.URL, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if ta != nil {
		t.CustomAbort(http.StatusConflict, fmt.Sprintf("the target whose endpoint is %s already exists", target.URL))
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	id, err := dao.AddRepTarget(*target)
	if err != nil {
		log.Errorf("failed to add target: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	t.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put ...
func (t *TargetAPI) Put() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	hasEnabledPolicy := false
	for _, policy := range policies {
		if policy.Enabled == 1 {
			hasEnabledPolicy = true
			break
		}
	}

	if hasEnabledPolicy {
		t.CustomAbort(http.StatusBadRequest, "the target is associated with policy which is enabled")
	}
	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleDecrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to decrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	req := struct {
		Name     *string `json:"name"`
		Endpoint *string `json:"endpoint"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Insecure *bool   `json:"insecure"`
	}{}
	t.DecodeJSONReq(&req)

	originalName := target.Name
	originalURL := target.URL

	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.Endpoint != nil {
		target.URL = *req.Endpoint
	}
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Password != nil {
		target.Password = *req.Password
	}
	if req.Insecure != nil {
		target.Insecure = *req.Insecure
	}

	t.Validate(target)

	if target.Name != originalName {
		ta, err := dao.GetRepTargetByName(target.Name)
		if err != nil {
			log.Errorf("failed to get target %s: %v", target.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.CustomAbort(http.StatusConflict, "name is already used")
		}
	}

	if target.URL != originalURL {
		ta, err := dao.GetRepTargetByEndpoint(target.URL)
		if err != nil {
			log.Errorf("failed to get target [ %s ]: %v", target.URL, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.CustomAbort(http.StatusConflict, fmt.Sprintf("the target whose endpoint is %s already exists", target.URL))
		}
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	if err := dao.UpdateRepTarget(*target); err != nil {
		log.Errorf("failed to update target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Delete ...
func (t *TargetAPI) Delete() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if len(policies) > 0 {
		t.CustomAbort(http.StatusPreconditionFailed, "the target is used by policies, can not be deleted")
	}

	if err = dao.DeleteRepTarget(id); err != nil {
		log.Errorf("failed to delete target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func newRegistryClient(endpoint string, insecure bool, username, password string) (*registry.Registry, error) {
	transport := registry.GetHTTPTransport(insecure)
	credential := auth.NewBasicAuthCredential(username, password)
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential)
	return registry.NewRegistry(endpoint, &http.Client{
		Transport: registry.NewTransport(transport, authorizer),
	})
}

// ListPolicies ...
func (t *TargetAPI) ListPolicies() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	t.Data["json"] = policies
	t.ServeJSON()
}
