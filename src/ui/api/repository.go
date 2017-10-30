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

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/clair"
	registry_error "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/notary"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/ui/config"
	uiutils "github.com/vmware/harbor/src/ui/utils"
)

// RepositoryAPI handles request to /api/repositories /api/repositories/tags /api/repositories/manifests, the parm has to be put
// in the query string as the web framework can not parse the URL if it contains veriadic sectors.
type RepositoryAPI struct {
	BaseController
}

type repoResp struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	ProjectID    int64     `json:"project_id"`
	Description  string    `json:"description"`
	PullCount    int64     `json:"pull_count"`
	StarCount    int64     `json:"star_count"`
	TagsCount    int64     `json:"tags_count"`
	CreationTime time.Time `json:"creation_time"`
	UpdateTime   time.Time `json:"update_time"`
}

type tagDetail struct {
	Digest        string    `json:"digest"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	Architecture  string    `json:"architecture"`
	OS            string    `json:"os"`
	DockerVersion string    `json:"docker_version"`
	Author        string    `json:"author"`
	Created       time.Time `json:"created"`
}

type tagResp struct {
	tagDetail
	Signature    *notary.Target          `json:"signature"`
	ScanOverview *models.ImgScanOverview `json:"scan_overview,omitempty"`
}

type manifestResp struct {
	Manifest interface{} `json:"manifest"`
	Config   interface{} `json:"config,omitempty" `
}

// Get ...
func (ra *RepositoryAPI) Get() {
	projectID, err := ra.GetInt64("project_id")
	if err != nil || projectID <= 0 {
		ra.HandleBadRequest(fmt.Sprintf("invalid project_id %s", ra.GetString("project_id")))
		return
	}

	exist, err := ra.ProjectMgr.Exists(projectID)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %d",
			projectID), err)
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %d not found", projectID))
		return
	}

	if !ra.SecurityCtx.HasReadPerm(projectID) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	keyword := ra.GetString("q")

	total, err := dao.GetTotalOfRepositoriesByProject(
		[]int64{projectID}, keyword)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to get total of repositories of project %d: %v",
			projectID, err))
		return
	}

	page, pageSize := ra.GetPaginationParams()

	repositories, err := getRepositories(projectID,
		keyword, pageSize, pageSize*(page-1))
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to get repository: %v", err))
		return
	}

	ra.SetPaginationHeader(total, page, pageSize)
	ra.Data["json"] = repositories
	ra.ServeJSON()
}

func getRepositories(projectID int64, keyword string,
	limit, offset int64) ([]*repoResp, error) {
	repositories, err := dao.GetRepositoriesByProject(projectID, keyword, limit, offset)
	if err != nil {
		return nil, err
	}

	return populateTagsCount(repositories)
}

func populateTagsCount(repositories []*models.RepoRecord) ([]*repoResp, error) {
	result := []*repoResp{}
	for _, repository := range repositories {
		repo := &repoResp{
			ID:           repository.RepositoryID,
			Name:         repository.Name,
			ProjectID:    repository.ProjectID,
			Description:  repository.Description,
			PullCount:    repository.PullCount,
			StarCount:    repository.StarCount,
			CreationTime: repository.CreationTime,
			UpdateTime:   repository.UpdateTime,
		}

		tags, err := getTags(repository.Name)
		if err != nil {
			return nil, err
		}
		repo.TagsCount = int64(len(tags))
		result = append(result, repo)
	}
	return result, nil
}

// Delete ...
func (ra *RepositoryAPI) Delete() {
	// using :splat to get * part in path
	repoName := ra.GetString(":splat")

	projectName, _ := utils.ParseRepository(repoName)
	project, err := ra.ProjectMgr.Get(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to get the project %s",
			projectName), err)
		return
	}

	if project == nil {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}

	if !ra.SecurityCtx.IsAuthenticated() {
		ra.HandleUnauthorized()
		return
	}

	if !ra.SecurityCtx.HasAllPerm(projectName) {
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	rc, err := uiutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	tags := []string{}
	tag := ra.GetString(":tag")
	if len(tag) == 0 {
		tagList, err := rc.ListTag()
		if err != nil {
			if regErr, ok := err.(*registry_error.HTTPError); ok {
				ra.CustomAbort(regErr.StatusCode, regErr.Detail)
			}

			log.Errorf("error occurred while listing tags of %s: %v", repoName, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}

		// TODO remove the logic if the bug of registry is fixed
		if len(tagList) == 0 {
			ra.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		tags = append(tags, tagList...)
	} else {
		tags = append(tags, tag)
	}

	if config.WithNotary() {
		signedTags, err := getSignatures(ra.SecurityCtx.GetUsername(), repoName)
		if err != nil {
			ra.HandleInternalServerError(fmt.Sprintf(
				"failed to get signatures for repository %s: %v", repoName, err))
			return
		}

		for _, t := range tags {
			digest, _, err := rc.ManifestExist(t)
			if err != nil {
				log.Errorf("Failed to Check the digest of tag: %s, error: %v", t, err.Error())
				ra.CustomAbort(http.StatusInternalServerError, err.Error())
			}
			log.Debugf("Tag: %s, digest: %s", t, digest)
			if _, ok := signedTags[digest]; ok {
				log.Errorf("Found signed tag, repostory: %s, tag: %s, deletion will be canceled", repoName, t)
				ra.CustomAbort(http.StatusPreconditionFailed, fmt.Sprintf("tag %s is signed", t))
			}
		}
	}

	for _, t := range tags {
		if err = rc.DeleteTag(t); err != nil {
			if regErr, ok := err.(*registry_error.HTTPError); ok {
				if regErr.StatusCode == http.StatusNotFound {
					continue
				}
				ra.CustomAbort(regErr.StatusCode, regErr.Detail)
			}
			log.Errorf("error occurred while deleting tag %s:%s: %v", repoName, t, err)
			ra.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		log.Infof("delete tag: %s:%s", repoName, t)

		go TriggerReplicationByRepository(project.ProjectID, repoName, []string{t}, models.RepOpDelete)

		go func(tag string) {
			if err := dao.AddAccessLog(models.AccessLog{
				Username:  ra.SecurityCtx.GetUsername(),
				ProjectID: project.ProjectID,
				RepoName:  repoName,
				RepoTag:   tag,
				Operation: "delete",
				OpTime:    time.Now(),
			}); err != nil {
				log.Errorf("failed to add access log: %v", err)
			}
		}(t)
	}

	exist, err := repositoryExist(repoName, rc)
	if err != nil {
		log.Errorf("failed to check the existence of repository %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}
	if !exist {
		if err = dao.DeleteRepository(repoName); err != nil {
			log.Errorf("failed to delete repository %s: %v", repoName, err)
			ra.CustomAbort(http.StatusInternalServerError, "")
		}
	}
}

// GetTag returns the tag of a repository
func (ra *RepositoryAPI) GetTag() {
	repository := ra.GetString(":splat")
	tag := ra.GetString(":tag")
	exist, _, err := ra.checkExistence(repository, tag)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of resource, error: %v", err))
		return
	}
	if !exist {
		ra.HandleNotFound(fmt.Sprintf("resource: %s:%s not found", repository, tag))
		return
	}
	project, _ := utils.ParseRepository(repository)
	if !ra.SecurityCtx.HasReadPerm(project) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	client, err := uiutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repository)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to initialize the client for %s: %v",
			repository, err))
		return
	}

	_, exist, err = client.ManifestExist(tag)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of %s:%s: %v", repository, tag, err))
		return
	}
	if !exist {
		ra.HandleNotFound(fmt.Sprintf("%s not found", tag))
		return
	}

	result := assemble(client, repository, []string{tag},
		ra.SecurityCtx.GetUsername())
	ra.Data["json"] = result[0]
	ra.ServeJSON()
}

// GetTags returns tags of a repository
func (ra *RepositoryAPI) GetTags() {
	repoName := ra.GetString(":splat")

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exists(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			projectName), err)
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}

	if !ra.SecurityCtx.HasReadPerm(projectName) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	client, err := uiutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	tags, err := client.ListTag()
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to get tag of %s: %v", repoName, err))
		return
	}

	ra.Data["json"] = assemble(client, repoName, tags, ra.SecurityCtx.GetUsername())
	ra.ServeJSON()
}

// get config, signature and scan overview and assemble them into one
// struct for each tag in tags
func assemble(client *registry.Repository, repository string,
	tags []string, username string) []*tagResp {

	var err error
	signatures := map[string][]notary.Target{}
	if config.WithNotary() {
		signatures, err = getSignatures(username, repository)
		if err != nil {
			signatures = map[string][]notary.Target{}
			log.Errorf("failed to get signatures of %s: %v", repository, err)
		}
	}

	result := []*tagResp{}
	for _, t := range tags {
		item := &tagResp{}

		// the detail information of tag
		tagDetail, err := getTagDetail(client, t)
		if err != nil {
			log.Errorf("failed to get v2 manifest of %s:%s: %v", repository, t, err)
		}
		if tagDetail != nil {
			item.tagDetail = *tagDetail
		}

		// scan overview
		if config.WithClair() {
			item.ScanOverview = getScanOverview(item.Digest, item.Name)
		}

		// signature, compare both digest and tag
		if config.WithNotary() {
			if sigs, ok := signatures[item.Digest]; ok {
				for _, sig := range sigs {
					if item.Name == sig.Tag {
						item.Signature = &sig
					}
				}
			}
		}

		result = append(result, item)
	}

	return result
}

// getTagDetail returns the detail information for v2 manifest image
// The information contains architecture, os, author, size, etc.
func getTagDetail(client *registry.Repository, tag string) (*tagDetail, error) {
	detail := &tagDetail{
		Name: tag,
	}

	digest, _, payload, err := client.PullManifest(tag, []string{schema2.MediaTypeManifest})
	if err != nil {
		return detail, err
	}
	detail.Digest = digest

	manifest := &schema2.DeserializedManifest{}
	if err = manifest.UnmarshalJSON(payload); err != nil {
		return detail, err
	}

	// size of manifest + size of layers
	detail.Size = int64(len(payload))
	for _, ref := range manifest.References() {
		detail.Size += ref.Size
	}

	_, reader, err := client.PullBlob(manifest.Target().Digest.String())
	if err != nil {
		return detail, err
	}

	configData, err := ioutil.ReadAll(reader)
	if err != nil {
		return detail, err
	}

	if err = json.Unmarshal(configData, detail); err != nil {
		return detail, err
	}

	return detail, nil
}

// GetManifests returns the manifest of a tag
func (ra *RepositoryAPI) GetManifests() {
	repoName := ra.GetString(":splat")
	tag := ra.GetString(":tag")

	version := ra.GetString("version")
	if len(version) == 0 {
		version = "v2"
	}

	if version != "v1" && version != "v2" {
		ra.CustomAbort(http.StatusBadRequest, "version should be v1 or v2")
	}

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exists(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			projectName), err)
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}

	if !ra.SecurityCtx.HasReadPerm(projectName) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}

		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	rc, err := uiutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("error occurred while initializing repository client for %s: %v", repoName, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	manifest, err := getManifest(rc, tag, version)
	if err != nil {
		if regErr, ok := err.(*registry_error.HTTPError); ok {
			ra.CustomAbort(regErr.StatusCode, regErr.Detail)
		}

		log.Errorf("error occurred while getting manifest of %s:%s: %v", repoName, tag, err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	ra.Data["json"] = manifest
	ra.ServeJSON()
}

func getManifest(client *registry.Repository,
	tag, version string) (*manifestResp, error) {
	result := &manifestResp{}

	mediaTypes := []string{}
	switch version {
	case "v1":
		mediaTypes = append(mediaTypes, schema1.MediaTypeManifest)
	case "v2":
		mediaTypes = append(mediaTypes, schema2.MediaTypeManifest)
	}

	_, mediaType, payload, err := client.PullManifest(tag, mediaTypes)
	if err != nil {
		return nil, err
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		return nil, err
	}

	result.Manifest = manifest

	deserializedmanifest, ok := manifest.(*schema2.DeserializedManifest)
	if ok {
		_, data, err := client.PullBlob(deserializedmanifest.Target().Digest.String())
		if err != nil {
			return nil, err
		}

		b, err := ioutil.ReadAll(data)
		if err != nil {
			return nil, err
		}

		result.Config = string(b)
	}

	return result, nil
}

//GetTopRepos returns the most populor repositories
func (ra *RepositoryAPI) GetTopRepos() {
	count, err := ra.GetInt("count", 10)
	if err != nil || count <= 0 {
		ra.CustomAbort(http.StatusBadRequest, "invalid count")
	}

	projectIDs := []int64{}
	projects, err := ra.ProjectMgr.GetPublic()
	if err != nil {
		ra.ParseAndHandleError("failed to get public projects", err)
		return
	}
	if ra.SecurityCtx.IsAuthenticated() {
		list, err := ra.SecurityCtx.GetMyProjects()
		if err != nil {
			ra.HandleInternalServerError(fmt.Sprintf("failed to get projects which the user %s is a member of: %v",
				ra.SecurityCtx.GetUsername(), err))
			return
		}
		projects = append(projects, list...)
	}

	for _, project := range projects {
		projectIDs = append(projectIDs, project.ProjectID)
	}

	repos, err := dao.GetTopRepos(projectIDs, count)
	if err != nil {
		log.Errorf("failed to get top repos: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal server error")
	}

	result, err := populateTagsCount(repos)
	if err != nil {
		log.Errorf("failed to popultate tags count to repositories: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal server error")
	}

	ra.Data["json"] = result
	ra.ServeJSON()
}

//GetSignatures returns signatures of a repository
func (ra *RepositoryAPI) GetSignatures() {
	repoName := ra.GetString(":splat")

	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exists(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			projectName), err)
		return
	}

	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}

	if !ra.SecurityCtx.HasReadPerm(projectName) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	targets, err := notary.GetInternalTargets(config.InternalNotaryEndpoint(),
		ra.SecurityCtx.GetUsername(), repoName)
	if err != nil {
		log.Errorf("Error while fetching signature from notary: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "internal error")
	}
	ra.Data["json"] = targets
	ra.ServeJSON()
}

//ScanImage handles request POST /api/repository/$repository/tags/$tag/scan to trigger image scan manually.
func (ra *RepositoryAPI) ScanImage() {
	if !config.WithClair() {
		log.Warningf("Harbor is not deployed with Clair, scan is disabled.")
		ra.RenderError(http.StatusServiceUnavailable, "")
		return
	}
	repoName := ra.GetString(":splat")
	tag := ra.GetString(":tag")
	projectName, _ := utils.ParseRepository(repoName)
	exist, err := ra.ProjectMgr.Exists(projectName)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			projectName), err)
		return
	}
	if !exist {
		ra.HandleNotFound(fmt.Sprintf("project %s not found", projectName))
		return
	}
	if !ra.SecurityCtx.IsAuthenticated() {
		ra.HandleUnauthorized()
		return
	}
	if !ra.SecurityCtx.HasAllPerm(projectName) {
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}
	err = uiutils.TriggerImageScan(repoName, tag)
	//TODO better check existence
	if err != nil {
		log.Errorf("Error while calling job service to trigger image scan: %v", err)
		ra.HandleInternalServerError("Failed to scan image, please check log for details")
		return
	}
}

// VulnerabilityDetails fetch vulnerability info from clair, transform to Harbor's format and return to client.
func (ra *RepositoryAPI) VulnerabilityDetails() {
	if !config.WithClair() {
		log.Warningf("Harbor is not deployed with Clair, it's not impossible to get vulnerability details.")
		ra.RenderError(http.StatusServiceUnavailable, "")
		return
	}
	repository := ra.GetString(":splat")
	tag := ra.GetString(":tag")
	exist, digest, err := ra.checkExistence(repository, tag)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to check the existence of resource, error: %v", err))
		return
	}
	if !exist {
		ra.HandleNotFound(fmt.Sprintf("resource: %s:%s not found", repository, tag))
		return
	}
	project, _ := utils.ParseRepository(repository)
	if !ra.SecurityCtx.HasReadPerm(project) {
		if !ra.SecurityCtx.IsAuthenticated() {
			ra.HandleUnauthorized()
			return
		}
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}
	res := []*models.VulnerabilityItem{}
	overview, err := dao.GetImgScanOverview(digest)
	if err != nil {
		ra.HandleInternalServerError(fmt.Sprintf("failed to get the scan overview, error: %v", err))
		return
	}
	if overview != nil && len(overview.DetailsKey) > 0 {
		clairClient := clair.NewClient(config.ClairEndpoint(), nil)
		log.Debugf("The key for getting details: %s", overview.DetailsKey)
		details, err := clairClient.GetResult(overview.DetailsKey)
		if err != nil {
			ra.HandleInternalServerError(fmt.Sprintf("Failed to get scan details from Clair, error: %v", err))
			return
		}
		res = transformVulnerabilities(details)
	}
	ra.Data["json"] = res
	ra.ServeJSON()
}

// ScanAll handles the api to scan all images on Harbor.
func (ra *RepositoryAPI) ScanAll() {
	if !config.WithClair() {
		log.Warningf("Harbor is not deployed with Clair, it's not possible to scan images.")
		ra.RenderError(http.StatusServiceUnavailable, "")
		return
	}
	if !ra.SecurityCtx.IsAuthenticated() {
		ra.HandleUnauthorized()
		return
	}
	projectIDStr := ra.GetString("project_id")
	if len(projectIDStr) > 0 { //scan images under the project only.
		pid, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || pid <= 0 {
			ra.HandleBadRequest(fmt.Sprintf("Invalid project_id %s", projectIDStr))
			return
		}
		if !ra.SecurityCtx.HasAllPerm(pid) {
			ra.HandleForbidden(ra.SecurityCtx.GetUsername())
			return
		}
		if err := uiutils.ScanImagesByProjectID(pid); err != nil {
			log.Errorf("Failed triggering scan images in project: %d, error: %v", pid, err)
			ra.HandleInternalServerError(fmt.Sprintf("Error: %v", err))
			return
		}
	} else { //scan all images in Harbor
		if !ra.SecurityCtx.IsSysAdmin() {
			ra.HandleForbidden(ra.SecurityCtx.GetUsername())
			return
		}
		if !utils.ScanAllMarker().Check() {
			log.Warningf("There is a scan all scheduled at: %v, the request will not be processed.", utils.ScanAllMarker().Next())
			ra.RenderError(http.StatusPreconditionFailed, "Unable handle frequent scan all requests")
			return
		}

		if err := uiutils.ScanAllImages(); err != nil {
			log.Errorf("Failed triggering scan all images, error: %v", err)
			ra.HandleInternalServerError(fmt.Sprintf("Error: %v", err))
			return
		}
		utils.ScanAllMarker().Mark()
	}
	ra.Ctx.ResponseWriter.WriteHeader(http.StatusAccepted)
}

func getSignatures(username, repository string) (map[string][]notary.Target, error) {
	targets, err := notary.GetInternalTargets(config.InternalNotaryEndpoint(),
		username, repository)
	if err != nil {
		return nil, err
	}

	signatures := map[string][]notary.Target{}
	for _, tgt := range targets {
		digest, err := notary.DigestFromTarget(tgt)
		if err != nil {
			return nil, err
		}
		signatures[digest] = append(signatures[digest], tgt)
	}

	return signatures, nil
}

func (ra *RepositoryAPI) checkExistence(repository, tag string) (bool, string, error) {
	project, _ := utils.ParseRepository(repository)
	exist, err := ra.ProjectMgr.Exists(project)
	if err != nil {
		return false, "", err
	}
	if !exist {
		log.Errorf("project %s not found", project)
		return false, "", nil
	}
	client, err := uiutils.NewRepositoryClientForUI(ra.SecurityCtx.GetUsername(), repository)
	if err != nil {
		return false, "", fmt.Errorf("failed to initialize the client for %s: %v", repository, err)
	}
	digest, exist, err := client.ManifestExist(tag)
	if err != nil {
		return false, "", fmt.Errorf("failed to check the existence of %s:%s: %v", repository, tag, err)
	}
	if !exist {
		log.Errorf("%s not found", tag)
		return false, "", nil
	}
	return true, digest, nil
}

//will return nil when it failed to get data.  The parm "tag" is for logging only.
func getScanOverview(digest string, tag string) *models.ImgScanOverview {
	if len(digest) == 0 {
		log.Debug("digest is nil")
		return nil
	}
	data, err := dao.GetImgScanOverview(digest)
	if err != nil {
		log.Errorf("Failed to get scan result for tag:%s, digest: %s, error: %v", tag, digest, err)
	}
	if data == nil {
		return nil
	}
	job, err := dao.GetScanJob(data.JobID)
	if err != nil {
		log.Errorf("Failed to get scan job for id:%d, error: %v", data.JobID, err)
		return nil
	} else if job == nil { //job does not exist
		log.Errorf("The scan job with id: %d does not exist, returning nil", data.JobID)
		return nil
	}
	data.Status = job.Status
	if data.Status != models.JobFinished {
		log.Debugf("Unsetting vulnerable related historical values, job status: %s", data.Status)
		data.Sev = 0
		data.CompOverview = nil
		data.DetailsKey = ""
	}
	return data
}
