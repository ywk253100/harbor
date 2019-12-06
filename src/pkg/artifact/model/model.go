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

package model

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Artifact is the abstract object managed by Harbor. It hides the
// underlying concrete detail and provides an unified artifact view
// for all users.
type Artifact struct {
	ID                      int64
	Type                    string // image, chart, etc
	Repository              *models.RepoRecord
	Tags                    []*Tag // the list of tags that attached to the artifact
	MediaType               string // the specific media type
	Digest                  string
	Size                    int64
	UploadTime              time.Time
	ExtraAttrs              map[string]interface{}     // only contains the simple attributes specific for the different artifact type, most of them should come from the config layer
	AdditionalResourceLinks map[string][]*ResourceLink // the resource link for build history(image), values.yaml(chart), dependency(chart), etc
	Annotations             map[string]string
	References              []*Reference // child artifacts referenced by the parent artifact if the artifact is an index
	// TODO: As the labels and signature aren't handled inside the artifact module,
	// we should move it to the API level artifact model rather than
	// keeping it here. The same to scan information
	// Labels                  []*models.Label
	// Signature               *Signature                 // add the signature in the artifact level rather than tag level as we cannot make sure the signature always apply to tag
}

// ResourceLink is a link via that a resource can be fetched
type ResourceLink struct {
	HREF     string
	Absolute bool // specify the href is an absolute URL or not
}

// Reference records the child artifact information referenced by
// other parent artifact
type Reference struct {
	ArtifactID int64
	Platform   *v1.Platform
}

// TODO: move it to the API level artifact model
// Signature information
// type Signature struct {
// 	Signatures map[string]bool // tag: signed or not
// }

// Tag belongs to one repository and can only be attached to a single
// one artifact under the repository
type Tag struct {
	ID                 int64
	Name               string
	UploadTime         time.Time
	LatestDownloadTime time.Time
}

// Query condition for query artifacts
type Query struct {
	RepositoryID int64
	Digest       string
	Tag          string
	Type         string // image, chart, etc
	Page         int64
	PageSize     int64
}

// Option for pruning artifact records
type Option struct {
	KeepUntagged bool // keep the untagged artifacts or not
}
