package proxy

import (
	"encoding/json"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/clair"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/notary"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/promgr"
	uiutils "github.com/vmware/harbor/src/ui/utils"

	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
)

type contextKey string

const (
	manifestURLPattern = `^/v2/((?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)+)manifests/([\w][\w.:-]{0,127})`
	catalogURLPattern  = `/v2/_catalog`
	imageInfoCtxKey    = contextKey("ImageInfo")
	//TODO: temp solution, remove after vmware/harbor#2242 is resolved.
	tokenUsername = "harbor-ui"
)

// Record the docker deamon raw response.
var rec *httptest.ResponseRecorder

// NotaryEndpoint , exported for testing.
var NotaryEndpoint = config.InternalNotaryEndpoint()

// MatchPullManifest checks if the request looks like a request to pull manifest.  If it is returns the image and tag/sha256 digest as 2nd and 3rd return values
func MatchPullManifest(req *http.Request) (bool, string, string) {
	//TODO: add user agent check.
	if req.Method != http.MethodGet {
		return false, "", ""
	}
	re := regexp.MustCompile(manifestURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 3 {
		s[1] = strings.TrimSuffix(s[1], "/")
		return true, s[1], s[2]
	}
	return false, "", ""
}

// MatchListRepos checks if the request looks like a request to list repositories.
func MatchListRepos(req *http.Request) bool {
	if req.Method != http.MethodGet {
		return false
	}
	re := regexp.MustCompile(catalogURLPattern)
	s := re.FindStringSubmatch(req.URL.Path)
	if len(s) == 1 {
		return true
	}
	return false
}

// policyChecker checks the policy of a project by project name, to determine if it's needed to check the image's status under this project.
type policyChecker interface {
	// contentTrustEnabled returns whether a project has enabled content trust.
	contentTrustEnabled(name string) bool
	// vulnerablePolicy  returns whether a project has enabled vulnerable, and the project's severity.
	vulnerablePolicy(name string) (bool, models.Severity)
}

type pmsPolicyChecker struct {
	pm promgr.ProjectManager
}

func (pc pmsPolicyChecker) contentTrustEnabled(name string) bool {
	project, err := pc.pm.Get(name)
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true
	}
	return project.ContentTrustEnabled()
}
func (pc pmsPolicyChecker) vulnerablePolicy(name string) (bool, models.Severity) {
	project, err := pc.pm.Get(name)
	if err != nil {
		log.Errorf("Unexpected error when getting the project, error: %v", err)
		return true, models.SevUnknown
	}
	return project.VulPrevented(), clair.ParseClairSev(project.Severity())
}

// newPMSPolicyChecker returns an instance of an pmsPolicyChecker
func newPMSPolicyChecker(pm promgr.ProjectManager) policyChecker {
	return &pmsPolicyChecker{
		pm: pm,
	}
}

func getPolicyChecker() policyChecker {
	return newPMSPolicyChecker(config.GlobalProjectMgr)
}

type imageInfo struct {
	repository  string
	reference   string
	projectName string
	digest      string
}

type urlHandler struct {
	next http.Handler
}

func (uh urlHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Debugf("in url handler, path: %s", req.URL.Path)
	req.URL.Path = strings.TrimPrefix(req.URL.Path, RegistryProxyPrefix)
	flag, repository, reference := MatchPullManifest(req)
	if flag {
		components := strings.SplitN(repository, "/", 2)
		if len(components) < 2 {
			http.Error(rw, marshalError(fmt.Sprintf("Bad repository name: %s", repository)), http.StatusBadRequest)
			return
		}

		client, err := uiutils.NewRepositoryClientForUI(tokenUsername, repository)
		if err != nil {
			log.Errorf("Error creating repository Client: %v", err)
			http.Error(rw, marshalError(fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}
		digest, _, err := client.ManifestExist(reference)
		if err != nil {
			log.Errorf("Failed to get digest for reference: %s, error: %v", reference, err)
			http.Error(rw, marshalError(fmt.Sprintf("Failed due to internal Error: %v", err)), http.StatusInternalServerError)
			return
		}

		img := imageInfo{
			repository:  repository,
			reference:   reference,
			projectName: components[0],
			digest:      digest,
		}

		log.Debugf("image info of the request: %#v", img)
		ctx := context.WithValue(req.Context(), imageInfoCtxKey, img)
		req = req.WithContext(ctx)
	}
	uh.next.ServeHTTP(rw, req)
}

type listReposHandler struct {
	next http.Handler
}

func (lrh listReposHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	listReposFlag := MatchListRepos(req)
	if listReposFlag {
		rec = httptest.NewRecorder()
		lrh.next.ServeHTTP(rec, req)
		if rec.Result().StatusCode != http.StatusOK {
			copyResp(rec, rw)
			return
		}
		var ctlg struct {
			Repositories []string `json:"repositories"`
		}
		decoder := json.NewDecoder(rec.Body)
		if err := decoder.Decode(&ctlg); err != nil {
			log.Errorf("Decode repositories error: %v", err)
			copyResp(rec, rw)
			return
		}
		var entries []string
		for repo := range ctlg.Repositories {
			log.Debugf("the repo in the reponse %s", ctlg.Repositories[repo])
			exist := dao.RepositoryExists(ctlg.Repositories[repo])
			if exist {
				entries = append(entries, ctlg.Repositories[repo])
			}
		}
		type Repos struct {
			Repositories []string `json:"repositories"`
		}
		resp := &Repos{Repositories: entries}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			log.Errorf("Encode repositories error: %v", err)
			copyResp(rec, rw)
			return
		}

		for k, v := range rec.Header() {
			rw.Header()[k] = v
		}
		clen := len(respJSON)
		rw.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(clen))
		rw.Write(respJSON)
		return
	}
	lrh.next.ServeHTTP(rw, req)
}

type contentTrustHandler struct {
	next http.Handler
}

func (cth contentTrustHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(imageInfoCtxKey)
	if imgRaw == nil || !config.WithNotary() {
		cth.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(imageInfoCtxKey).(imageInfo)
	if img.digest == "" {
		cth.next.ServeHTTP(rw, req)
		return
	}
	if !getPolicyChecker().contentTrustEnabled(img.projectName) {
		cth.next.ServeHTTP(rw, req)
		return
	}
	match, err := matchNotaryDigest(img)
	if err != nil {
		http.Error(rw, marshalError("Failed in communication with Notary please check the log"), http.StatusInternalServerError)
		return
	}
	if !match {
		log.Debugf("digest mismatch, failing the response.")
		http.Error(rw, marshalError("The image is not signed in Notary."), http.StatusPreconditionFailed)
		return
	}
	cth.next.ServeHTTP(rw, req)
}

type vulnerableHandler struct {
	next http.Handler
}

func (vh vulnerableHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(imageInfoCtxKey)
	if imgRaw == nil || !config.WithClair() {
		vh.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(imageInfoCtxKey).(imageInfo)
	if img.digest == "" {
		vh.next.ServeHTTP(rw, req)
		return
	}
	projectVulnerableEnabled, projectVulnerableSeverity := getPolicyChecker().vulnerablePolicy(img.projectName)
	if !projectVulnerableEnabled {
		vh.next.ServeHTTP(rw, req)
		return
	}
	overview, err := dao.GetImgScanOverview(img.digest)
	if err != nil {
		log.Errorf("failed to get ImgScanOverview with repo: %s, reference: %s, digest: %s. Error: %v", img.repository, img.reference, img.digest, err)
		http.Error(rw, marshalError("Failed to get ImgScanOverview."), http.StatusPreconditionFailed)
		return
	}
	// severity is 0 means that the image fails to scan or not scanned successfully.
	if overview == nil || overview.Sev == 0 {
		log.Debugf("cannot get the image scan overview info, failing the response.")
		http.Error(rw, marshalError("Cannot get the image severity."), http.StatusPreconditionFailed)
		return
	}
	imageSev := overview.Sev
	if imageSev >= int(projectVulnerableSeverity) {
		log.Debugf("the image severity: %q is higher then project setting: %q, failing the response.", models.Severity(imageSev), projectVulnerableSeverity)
		http.Error(rw, marshalError(fmt.Sprintf("The severity of vulnerability of the image: %q is equal or higher than the threshold in project setting: %q.", models.Severity(imageSev), projectVulnerableSeverity)), http.StatusPreconditionFailed)
		return
	}
	vh.next.ServeHTTP(rw, req)
}

func matchNotaryDigest(img imageInfo) (bool, error) {
	targets, err := notary.GetInternalTargets(NotaryEndpoint, tokenUsername, img.repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if isDigest(img.reference) {
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			if img.digest == d {
				return true, nil
			}
		} else {
			if t.Tag == img.reference {
				log.Debugf("found reference: %s in notary, try to match digest.", img.reference)
				d, err := notary.DigestFromTarget(t)
				if err != nil {
					return false, err
				}
				if img.digest == d {
					return true, nil
				}
			}
		}
	}
	log.Debugf("image: %#v, not found in notary", img)
	return false, nil
}

//A sha256 is a string with 64 characters.
func isDigest(ref string) bool {
	return strings.HasPrefix(ref, "sha256:") && len(ref) == 71
}

func copyResp(rec *httptest.ResponseRecorder, rw http.ResponseWriter) {
	for k, v := range rec.Header() {
		rw.Header()[k] = v
	}
	rw.WriteHeader(rec.Result().StatusCode)
	rw.Write(rec.Body.Bytes())
}

func marshalError(msg string) string {
	var tmpErrs struct {
		Errors []JSONError `json:"errors,omitempty"`
	}
	tmpErrs.Errors = append(tmpErrs.Errors, JSONError{
		Code:    "PROJECT_POLICY_VIOLATION",
		Message: msg,
		Detail:  msg,
	})

	str, err := json.Marshal(tmpErrs)
	if err != nil {
		log.Debugf("failed to marshal json error, %v", err)
		return msg
	}
	return string(str)
}

// JSONError wraps a concrete Code and Message, it's readable for docker deamon.
type JSONError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}
