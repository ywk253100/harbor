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
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/filter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"strings"
)

func (a *adapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	repoFilters, err := filter.BuildRepositoryFilters(filters)
	if err != nil {
		return nil, err
	}
	artFilters, err := filter.BuildArtifactFilters(filters)
	if err != nil {
		return nil, err
	}

	projects, err := a.listProjects(filters)
	if err != nil {
		return nil, err
	}

	resources := []*model.Resource{}
	for _, project := range projects {
		repositories, err := a.listRepositories(project, repoFilters)
		if err != nil {
			return nil, err
		}
		if len(repositories) == 0 {
			continue
		}

		var rawResources = make([]*model.Resource, len(repositories))
		runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
		defer runner.Cancel()

		for i, r := range repositories {
			index := i
			repository := r
			runner.AddTask(func() error {
				artifacts, err := a.listArtifacts(repository.Name, artFilters)
				if err != nil {
					return fmt.Errorf("failed to list artifacts of repository '%s': %v", repository.Name, err)
				}
				if len(artifacts) == 0 {
					rawResources[index] = nil
					return nil
				}

				rawResources[index] = &model.Resource{
					Registry: a.registry,
					Repository: &model.Repository{
						Type:     model.RepositoryTypeOCIRegistry,
						Name:     repository.Name,
						Metadata: project.Metadata,
					},
					Artifacts:    artifacts,
					ExtendedInfo: nil,
					Deleted:      false,
					Override:     false,
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

func (a *adapter) listProjects(filters []*model.Filter) ([]*project, error) {
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

func (a *adapter) listRepositories(project *project, filters filter.RepositoryFilters) ([]*model.Repository, error) {
	repositories := []*models.RepoRecord{}
	url := fmt.Sprintf("%s/api/%s/projects/%s/repositories", a.getURL(), api.APIVersion, project.Name)
	if err := a.client.GetAndIteratePagination(url, &repositories); err != nil {
		return nil, err
	}
	var repos []*model.Repository
	for _, repository := range repositories {
		repos = append(repos, &model.Repository{
			Type:     model.RepositoryTypeOCIRegistry,
			Name:     repository.Name,
			Metadata: project.Metadata,
		})
	}
	return filters.Filter(repos)
}

func (a *adapter) listArtifacts(repository string, filters filter.ArtifactFilters) ([]*model.Artifact, error) {
	project, repository := utils.ParseRepository(repository)
	url := fmt.Sprintf("%s/api/%s/projects/%s/repositories/%s/artifacts",
		a.getURL(), api.APIVersion, project, repository)
	artifacts := []*artifact.Artifact{}
	if err := a.client.Get(url, &artifacts); err != nil {
		return nil, err
	}
	var arts []*model.Artifact
	for _, artifact := range artifacts {
		art := &model.Artifact{
			Type:   artifact.Type,
			Digest: artifact.Digest,
		}
		for _, label := range artifact.Labels {
			art.Labels = append(art.Labels, label.Name)
		}
		for _, tag := range artifact.Tags {
			art.Tags = append(art.Tags, tag.Name)
		}
		arts = append(arts, art)
	}
	return filters.Filter(arts)
}
