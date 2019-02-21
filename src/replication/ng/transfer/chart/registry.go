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

package chart

import (
	"io"
	"io/ioutil"
	"net/http"

	com_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/util"
)

type Registry interface {
	Upload(path string, chart io.Reader) error
	Download(path string) (io.ReadCloser, error)
	Exist(path string) (bool, error)
	Delete(path string) error
}

// NewRegistry return an instance of the implement of Registry
func NewRegistry(reg *model.Registry) Registry {
	modifiers := []modifier.Modifier{}
	if reg.Credential != nil {
		authorizer := auth.NewBasicAuthorizer(
			reg.Credential.AccessKey,
			reg.Credential.AccessSecret)
		modifiers = append(modifiers, authorizer)
	}

	client := com_http.NewClient(&http.Client{
		Transport: util.GetTransport(reg.Insecure),
	}, modifiers...)

	return &registry{
		url:    reg.URL,
		client: client,
	}
}

type registry struct {
	url    string
	client *com_http.Client
}

func (r *registry) Upload(path string, chart io.Reader) error {
	return r.client.Post(r.url+path, chart)
}
func (r *registry) Download(path string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, r.url+path, nil)
	if err != nil {
		return nil, err
	}
	// only close the resp.Body when got error as we need to
	// return a ReadCloser. The caller is responsible to close it
	resp, err := r.client.Do(req)
	if err != nil {
		defer resp.Body.Close()
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		errMsg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, &com_http.Error{
			Code:    resp.StatusCode,
			Message: string(errMsg),
		}
	}
	return resp.Body, nil
}
func (r *registry) Exist(path string) (bool, error) {
	err := r.client.Get(r.url + path)
	// exist
	if err == nil {
		return true, nil
	}
	// not exist
	if httpErr, ok := err.(*com_http.Error); ok && httpErr.Code == http.StatusNotFound {
		return false, nil
	}
	// got error
	return false, err
}
func (r *registry) Delete(path string) error {
	return r.client.Delete(r.url + path)
}
