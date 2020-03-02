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

// the resource type
const (
	RepositoryTypeChartMuseum = "CHART_MUSEUM"
	RepositoryTypeOCIRegistry = "OCI_REGISTRY"
)

// the resource type
const (
	ResourceTypeImage ResourceType = "image"
	ResourceTypeChart ResourceType = "chart"
)

// ResourceType represents the type of the resource
type ResourceType string

// Valid indicates whether the ResourceType is a valid value
func (r ResourceType) Valid() bool {
	return len(r) > 0
}

// Resource represents the general replicating content
type Resource struct {
	Registry     *Registry              `json:"registry"`
	Repository   *Repository            `json:"repository"`
	Artifacts    []*Artifact            `json:"artifact"`
	ExtendedInfo map[string]interface{} `json:"extended_info"`
	Deleted      bool                   `json:"deleted"`  // Indicate if the resource is a deleted resource
	Override     bool                   `json:"override"` // indicate whether the resource can be overridden
}

// Repository info of the resource
type Repository struct {
	Type     string                 `json:"type"` // chartmuseum repo or registry repo
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

type Artifact struct {
	Type   string   `json:"type"`
	Digest string   `json:"digest"`
	Labels []string `json:"labels"`
	Tags   []string `json:"tags"`
}
