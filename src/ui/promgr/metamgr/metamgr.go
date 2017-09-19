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

package metamgr

//import (
//"github.com/vmware/harbor/src/common/dao"
//)

type MetaMgr interface {
	Add(projectID int64, meta map[string]string) error
	Delete(projecdtID int64, meta ...[]string) error
	Update(projectID int64, meta map[string]string) error
	Get(projectID int64, meta ...[]string) (map[string]string, error)
}

// TODO add ut

type defaultMetaMgr struct{}

func NewDefaultMetaMgr() MetaMgr {
	return &defaultMetaMgr{}
}

func (d *defaultMetaMgr) Add(projectID int64, meta map[string]string) error {
	//return dao.AddProMeta(projectID, meta)
	return nil
}

func (d *defaultMetaMgr) Delete(projectID int64, meta ...[]string) error {
	// return dao.DeleteProMeta(projectID, meta...)
	return nil
}

func (d *defaultMetaMgr) Update(projectID int64, meta map[string]string) error {
	// return dao.UpdateProMeta(projectID, meta...)
	return nil
}

func (d *defaultMetaMgr) Get(projectID int64, meta ...[]string) (map[string]string, error) {
	//return dao.GetProMeta(projectID, meta...)
	return nil, nil
}
