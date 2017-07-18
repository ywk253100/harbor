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

// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"strings"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/service/token"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// ScanAllImages scans all images of Harbor by submiting jobs to jobservice, the whole process will move one if failed to subit any job of a single image.
func ScanAllImages() error {
	repos, err := dao.GetAllRepositories()
	if err != nil {
		log.Errorf("Failed to list all repositories, error: %v", err)
		return err
	}
	log.Infof("Rescanning all images.")

	go func() {
		var repoClient *registry.Repository
		var err error
		var tags []string
		for _, r := range repos {
			repoClient, err = NewRepositoryClientForUI("harbor-ui", r.Name, "pull")
			if err != nil {
				log.Errorf("Failed to initialize client for repository: %s, error: %v, skip scanning", r.Name, err)
				continue
			}
			tags, err = repoClient.ListTag()
			if err != nil {
				log.Errorf("Failed to get tags for repository: %s, error: %v, skip scanning.", r.Name, err)
				continue
			}
			for _, t := range tags {
				if err = TriggerImageScan(r.Name, t); err != nil {
					log.Errorf("Failed to scan image with repository: %s, tag: %s, error: %v.", r.Name, t, err)
				} else {
					log.Debugf("Triggered scan for image with repository: %s, tag: %s", r.Name, t)
				}
			}
		}
	}()
	return nil
}

// RequestAsUI is a shortcut to make a request attach UI secret and send the request.
// Do not use this when you want to handle the response
// TODO: add a response handler to replace expectSC *when needed*
func RequestAsUI(method, url string, body io.Reader, expectSC int) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	AddUISecret(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectSC {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("Unexpected status code: %d, text: %s", resp.StatusCode, string(b))
	}
	return nil
}

//AddUISecret add secret cookie to a request
func AddUISecret(req *http.Request) {
	if req != nil {
		req.AddCookie(&http.Cookie{
			Name:  models.UISecretCookie,
			Value: config.UISecret(),
		})
	}
}

// TriggerImageScan triggers an image scan job on jobservice.
func TriggerImageScan(repository string, tag string) error {
	data := &models.ImageScanReq{
		Repo: repository,
		Tag:  tag,
	}
	b, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/jobs/scan", config.InternalJobServiceURL())
	return RequestAsUI("POST", url, bytes.NewBuffer(b), http.StatusOK)
}

// NewRepositoryClientForUI ...
func NewRepositoryClientForUI(username, repository string, actions ...string) (*registry.Repository, error) {
	url, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}
	token, err := token.RegistryTokenForUI(username, "harbor-registry",
		[]string{fmt.Sprintf("repository:%s:%s", repository, strings.Join(actions, ","))})
	authorizer := auth.NewRawTokenAuthorizer(token.Token)
	return registry.NewRepositoryWithModifiers(repository, url,
		true, authorizer)
}
