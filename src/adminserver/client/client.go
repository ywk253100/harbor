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

package client

import (
	"strings"

	"github.com/vmware/harbor/src/adminserver/systeminfo/imagestorage"
<<<<<<< HEAD
	"github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/http/modifier/auth"
=======
	httpclient "github.com/vmware/harbor/src/common/http/client"
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	"github.com/vmware/harbor/src/common/utils"
)

// Client defines methods that an Adminserver client should implement
type Client interface {
	// Ping tests the connection with server
	Ping() error
	// GetCfgs returns system configurations
	GetCfgs() (map[string]interface{}, error)
	// UpdateCfgs updates system configurations
	UpdateCfgs(map[string]interface{}) error
	// ResetCfgs resets system configuratoins form environment variables
	ResetCfgs() error
	// Capacity returns the capacity of image storage
	Capacity() (*imagestorage.Capacity, error)
}

// NewClient return an instance of Adminserver client
<<<<<<< HEAD
func NewClient(baseURL string, cfg *Config) Client {
=======
func NewClient(baseURL string, c httpclient.Client) Client {
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.Contains(baseURL, "://") {
		baseURL = "http://" + baseURL
	}
<<<<<<< HEAD
	client := &client{
		baseURL: baseURL,
=======
	return &client{
		baseURL: baseURL,
		client:  c,
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	}
	if cfg != nil {
		authorizer := auth.NewSecretAuthorizer(cfg.Secret)
		client.client = http.NewClient(nil, authorizer)
	}
	return client
}

type client struct {
	baseURL string
<<<<<<< HEAD
	client  *http.Client
}

// Config contains configurations needed for client
type Config struct {
	Secret string
=======
	client  httpclient.Client
}

// do creates request and authorizes it if authorizer is not nil
func (c *client) do(method, relativePath string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + relativePath
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return c.client.Do(req)
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
}

func (c *client) Ping() error {
	addr := strings.Split(c.baseURL, "://")[1]
	if !strings.Contains(addr, ":") {
		addr = addr + ":80"
	}

	return utils.TestTCPConn(addr, 60, 2)
}

// GetCfgs ...
func (c *client) GetCfgs() (map[string]interface{}, error) {
	url := c.baseURL + "/api/configurations"
	cfgs := map[string]interface{}{}
	if err := c.client.Get(url, &cfgs); err != nil {
		return nil, err
	}
	return cfgs, nil
}

// UpdateCfgs ...
func (c *client) UpdateCfgs(cfgs map[string]interface{}) error {
	url := c.baseURL + "/api/configurations"
	return c.client.Put(url, cfgs)
}

// ResetCfgs ...
func (c *client) ResetCfgs() error {
	url := c.baseURL + "/api/configurations/reset"
	return c.client.Post(url)
}

// Capacity ...
func (c *client) Capacity() (*imagestorage.Capacity, error) {
	url := c.baseURL + "/api/systeminfo/capacity"
	capacity := &imagestorage.Capacity{}
	if err := c.client.Get(url, capacity); err != nil {
		return nil, err
	}
	return capacity, nil
}
