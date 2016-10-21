/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/vmware/harbor/utils/log"
)

type FD struct {
	BaseAPI
}

func (f *FD) Get() {
	client := &http.Client{
		Transport: &http.Transport{},
	}

	resp, err := client.Get("http://cn.bing.com")
	if err != nil {
		log.Errorf("%v", err)
		f.CustomAbort(http.StatusInternalServerError, "")
	}
	defer resp.Body.Close()
	if _, err = io.Copy(ioutil.Discard, resp.Body); err != nil {
		log.Errorf("%v", err)
		f.CustomAbort(http.StatusInternalServerError, "")
	}
	f.Data["json"] = "success"
	f.ServeJSON()
}
