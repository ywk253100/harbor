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

package harbor

import (
	"fmt"
	art "github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/filter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"github.com/goharbor/harbor/src/server"
	"strings"
)

func (a *adapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	projects, err := a.listCandidateProjects(filters)
	if err != nil {
		return nil, err
	}

	resources := []*model.Resource{}
	for _, project := range projects {
		repositories, err := a.getRepositories(project.Name)
		if err != nil {
			return nil, err
		}
		if len(repositories) == 0 {
			continue
		}
		for _, filter := range filters {
			if err = filter.DoFilter(&repositories); err != nil {
				return nil, err
			}
		}

		var rawResources = make([]*model.Resource, len(repositories))
		runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
		defer runner.Cancel()

		for i, r := range repositories {
			index := i
			repo := r
			runner.AddTask(func() error {
				artifacts, err := a.listArtifacts(repo.Name)
				if err != nil {
					return fmt.Errorf("failed to list artifacts of repository '%s': %v", repo.Name, err)
				}
				if len(artifacts) == 0 {
					rawResources[index] = nil
					return nil
				}
				for _, filter := range filters {
					if err = filter.DoFilter(&artifacts); err != nil {
						return fmt.Errorf("failed to filter the artifacts: %v", err)
					}
				}
				if len(artifacts) == 0 {
					rawResources[index] = nil
					return nil
				}
				var vTags []*adp.VTag
				for _, artifact := range artifacts {
					for _, tag := range artifact.Tags {
						vTags = append(vTags, &adp.VTag{
							ResourceType: string(model.ResourceTypeImage),
							Name:         tag.Name,
						})
					}
				}
				for _, filter := range filters {
					if err = filter.DoFilter(&vTags); err != nil {
						return fmt.Errorf("failed to filter the vtags: %v", err)
					}
				}
				tags := []string{}
				for _, vTag := range vTags {
					tags = append(tags, vTag.Name)
				}
				rawResources[index] = &model.Resource{
					Type:     model.ResourceTypeImage,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name:     repo.Name,
							Metadata: project.Metadata,
						},
						Vtags: tags,
					},
				}

				return nil
			})
		}
		runner.Wait()

		if runner.IsCancelled() {
			return nil, fmt.Errorf("FetchImages error when collect tags for repos")
		}

		for _, r := range rawResources {
			if r != nil {
				resources = append(resources, r)
			}
		}
	}

	return resources, nil
}

func (a *adapter) listCandidateProjects(filters []*model.Filter) ([]*project, error) {
	pattern := ""
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			pattern = filter.Value.(string)
			break
		}
	}
	projects := []*project{}
	if len(pattern) > 0 {
		substrings := strings.Split(pattern, "/")
		projectPattern := substrings[0]
		names, ok := util.IsSpecificPathComponent(projectPattern)
		if ok {
			for _, name := range names {
				project, err := a.getProject(name)
				if err != nil {
					return nil, err
				}
				if project == nil {
					continue
				}
				projects = append(projects, project)
			}
		}
	}
	if len(projects) > 0 {
		names := []string{}
		for _, project := range projects {
			names = append(names, project.Name)
		}
		log.Debugf("parsed the projects %v from pattern %s", names, pattern)
		return projects, nil
	}
	return a.getProjects("")
}

func (a *adapter) listArtifacts(repository string) ([]*artifact, error) {
	project, repository := utils.ParseRepository(repository)
	url := fmt.Sprintf("%s/api/%s/projects/%s/repositories/%s/artifacts",
		a.getURL(), server.APIVersion, project, repository)
	artifacts := []*artifact{}
	if err := a.client.Get(url, &artifacts); err != nil {
		return nil, err
	}
	return artifacts, nil
}

type artifact struct {
	art.Artifact
}

func (a *artifact) GetFilterableType() filter.FilterableType {
	return filter.FilterableTypeArtifact
}
func (a *artifact) GetResourceType() string {
	return string(model.ResourceTypeImage)
}
func (a *artifact) GetName() string {
	return ""
}
func (a *artifact) GetLabels() []string {
	// TODO set labels
	return nil
}
