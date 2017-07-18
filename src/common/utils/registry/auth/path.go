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

package auth

import (
	"regexp"

	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/reference"
	"github.com/vmware/harbor/src/common/utils/log"
)

/*
	get, "/v2/",
	get "/v2/{name:" + reference.NameRegexp.String() + "}/tags/list",
	get, put, delete "/v2/{name:" + reference.NameRegexp.String() + "}/manifests/{reference:" + reference.TagRegexp.String() + "|" + digest.DigestRegexp.String() + "}",
	get, delete "/v2/{name:" + reference.NameRegexp.String() + "}/blobs/{digest:" + digest.DigestRegexp.String() + "}",
	post "/v2/{name:" + reference.NameRegexp.String() + "}/blobs/uploads/",
	get, patch, put, delete"/v2/{name:" + reference.NameRegexp.String() + "}/blobs/uploads/{uuid:[a-zA-Z0-9-_.=]+}",
	get "/v2/_catalog",

*/

var (
	base            = regexp.MustCompile("/v2")
	catalog         = regexp.MustCompile("/v2/_catalog")
	tag             = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/tags/list")
	manifest        = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/manifests/" + reference.TagRegexp.String() + "|" + digest.DigestRegexp.String())
	blob            = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/blobs/" + digest.DigestRegexp.String())
	blobUpload      = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/blobs/uploads/")
	blobUploadChunk = regexp.MustCompile("/v2/(" + reference.NameRegexp.String() + ")/blobs/uploads/[a-zA-Z0-9-_.=]+")

	repoRegExps = []*regexp.Regexp{tag, manifest, blob, blobUpload, blobUploadChunk}
)

// parse the repository name from path, if the path doesn't match any
// regular expressions in repoRegExps, nil string will be returned
func parseRepository(path string) string {
	for _, regExp := range repoRegExps {
		subs := regExp.FindStringSubmatch(path)
		// no match
		if subs == nil {
			continue
		}

		// match
		if len(subs) < 2 {
			log.Warningf("unexpected length of sub matches: %d, should >= 2 ", len(subs))
			continue
		}
		return subs[1]
	}
	return ""
}
