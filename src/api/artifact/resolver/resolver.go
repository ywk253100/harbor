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

package resolver

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/api/artifact/blob"
	"github.com/goharbor/harbor/src/api/artifact/resolver/image"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

var (
	registry = map[string]Resolver{}
)

// register resolvers for artifacts
func init() {
	register(image.NewManifestV2Resolver(), image.ManifestV2ResolverMediaTypes...)
	register(image.NewManifestV1Resolver(), image.ManifestV1ResolverMediaTypes...)
	register(image.NewIndexResolver(), image.IndexResolverMediaTypes...)
}

// Resolver resolves the detail information for a specific kind of artifact
type Resolver interface {
	// ArtifactType returns the type of artifact that the resolver handles
	ArtifactType(context.Context) string
	// Resolve receives the manifest content, resolves the detail information
	// from the manifest or the layers referenced by the manifest, and populates
	// the detail information into the artifact
	// The layers can be read by reader
	// The "repoFullName" is the "project_name/repository_name" that the artifact belongs to
	Resolve(ctx context.Context, manifest []byte, repoFullName string, artifact *artifact.Artifact, reader blob.Reader) error
}

func register(resolver Resolver, mediaTypes ...string) {
	if err := Register(resolver, mediaTypes...); err != nil {
		log.Errorf("failed to register resolver for artifact %s: %v", resolver.ArtifactType(nil), err)
	}
}

// Register resolver, one resolver can handle multiple media types for one kind of artifact
func Register(resolver Resolver, mediaTypes ...string) error {
	for _, mediaType := range mediaTypes {
		_, exist := registry[mediaType]
		if exist {
			return fmt.Errorf("resolver to handle media type %s already exists", mediaType)
		}
		registry[mediaType] = resolver
		log.Debugf("resolver to handle media type %s registered", mediaType)
	}
	return nil
}

// Get resolver according to the artifact media type
func Get(mediaType string) (Resolver, error) {
	resolver := registry[mediaType]
	if resolver == nil {
		return nil, fmt.Errorf("resolver for artifact media type %s not found", mediaType)
	}
	return resolver, nil
}
