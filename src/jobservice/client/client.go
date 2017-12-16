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
<<<<<<< HEAD
	"github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/http/modifier/auth"
=======
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	commonhttp "github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/http/client"
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	"github.com/vmware/harbor/src/jobservice/api"
)

// Client defines the methods that a jobservice client should implement
type Client interface {
	SubmitReplicationJob(*api.ReplicationReq) error
}

// DefaultClient provides a default implement for the interface Client
type DefaultClient struct {
	endpoint string
<<<<<<< HEAD
	client   *http.Client
}

// Config contains configuration items needed for DefaultClient
type Config struct {
	Secret string
}

// NewDefaultClient returns an instance of DefaultClient
func NewDefaultClient(endpoint string, cfg *Config) *DefaultClient {
=======
	client   client.Client
}

// NewDefaultClient returns an instance of DefaultClient
func NewDefaultClient(endpoint string, client ...client.Client) *DefaultClient {
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	c := &DefaultClient{
		endpoint: endpoint,
	}

<<<<<<< HEAD
	if cfg != nil {
		c.client = http.NewClient(nil, auth.NewSecretAuthorizer(cfg.Secret))
=======
	if len(client) > 0 {
		c.client = client[0]
	}

	if c.client == nil {
		c.client = &http.Client{}
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	}

	return c
}

// SubmitReplicationJob submits a replication job to the jobservice
func (d *DefaultClient) SubmitReplicationJob(replication *api.ReplicationReq) error {
	url := d.endpoint + "/api/jobs/replication"
<<<<<<< HEAD
	return d.client.Post(url, replication)
=======

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(replication); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, buffer)
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(message),
		}
	}

	return nil
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
}
