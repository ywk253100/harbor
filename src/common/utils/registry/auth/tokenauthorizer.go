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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
)

const (
	latency int = 10 //second, the network latency when token is received
	scheme      = "bearer"
)

// Scope ...
type Scope struct {
	Type    string
	Name    string
	Actions []string
}

func (s *Scope) string() string {
	return fmt.Sprintf("%s:%s:%s", s.Type, s.Name, strings.Join(s.Actions, ","))
}

type rawTokenAuthorizer struct {
	token string
}

// NewRawTokenAuthorizer returns an instance of rawTokenAuthorizer
func NewRawTokenAuthorizer(token string) registry.Modifier {
	return &rawTokenAuthorizer{
		token: token,
	}
}

func (r rawTokenAuthorizer) Modify(req *http.Request) error {
	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", r.token))
	return nil
}

// Implements interface Authorizer
type standardTokenAuthorizer struct {
	realm                  string
	service                string
	customizedTokenService string
	registryURL            *url.URL // used to filter request
	cachedTokens           map[string]*models.Token
	client                 *http.Client
	credential             Credential
	sync.RWMutex
}

// NewStandardTokenAuthorizer returns a standard token authorizer. The authorizer will request a token
// from token server and add it to the origin request
// If tokenServiceEndpoint is set, the token request will be sent to it instead of the server get from authorizer
// The usage please refer to the function tokenURL
func NewStandardTokenAuthorizer(credential Credential, insecure bool,
	customizedTokenService ...string) registry.Modifier {
	authorizer := &standardTokenAuthorizer{
		client: &http.Client{
			Transport: registry.GetHTTPTransport(insecure),
			Timeout:   30 * time.Second,
		},
		credential:   credential,
		cachedTokens: make(map[string]*models.Token),
	}

	if len(customizedTokenService) > 0 {
		authorizer.customizedTokenService = customizedTokenService[0]
	}

	return authorizer
}

func (s *standardTokenAuthorizer) Modify(req *http.Request) error {
	// ping first if registry URL, realm or service is null
	if s.registryURL == nil || len(s.realm) == 0 || len(s.service) == 0 {
		if err := s.ping(req.URL.Scheme + req.URL.Host); err != nil {
			return err
		}
	}

	//only handle requests sent to registry
	goon, err := s.filterReq(req)
	if err != nil {
		return err
	}
	if !goon {
		log.Debugf("the request %s is not sent to registry, skip", req.URL.String())
		return nil
	}

	scopes := []*Scope{}
	scope, err := parseScope(req)
	if err != nil {
		return err
	}
	scopes = append(scopes, scope)

	from := req.URL.Query().Get("from")
	if len(from) != 0 {
		scopes = append(scopes, &Scope{
			Type:    "repository",
			Name:    from,
			Actions: []string{"pull"},
		})
	}

	var token *models.Token
	// try to get token from cache if the request is for single scope
	if len(scopes) == 1 {
		token = s.getCachedToken(scopes[0])
	}

	// request a new token if token is null
	if token == nil {
		token, err = s.requestToken(scopes)
		if err != nil {
			return err
		}
		// only cache the token for single scope request
		if len(scopes) == 1 {
			s.updateCachedToken(scopes[0], token)
		}
	}

	req.Header.Add(http.CanonicalHeaderKey("Authorization"), fmt.Sprintf("Bearer %s", token.Token))

	return nil
}

// only handle requests sent to registry
func (s *standardTokenAuthorizer) filterReq(req *http.Request) (bool, error) {
	v2Index := strings.Index(req.URL.Path, "/v2/")
	if v2Index == -1 {
		return false, nil
	}

	if req.URL.Host != s.registryURL.Host || req.URL.Scheme != s.registryURL.Scheme ||
		req.URL.Path[:v2Index+4] != s.registryURL.Path {
		return false, nil
	}

	return true, nil
}

func (s *standardTokenAuthorizer) getCachedToken(scope *Scope) *models.Token {
	s.RLock()
	defer s.RUnlock()
	token := s.cachedTokens[scope.string()]
	if token == nil {
		return nil
	}

	issueAt, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", token.IssuedAt)
	if err != nil {
		log.Errorf("failed parse %s: %v", token.IssuedAt, err)
		return nil
	}

	if issueAt.Add(time.Duration(token.ExpiresIn-latency) * time.Second).Before(time.Now().UTC()) {
		return nil
	}

	log.Debug("get token from cache")
	return token
}

func (s *standardTokenAuthorizer) updateCachedToken(scope *Scope, token *models.Token) {
	s.Lock()
	defer s.Unlock()
	s.cachedTokens[scope.string()] = token
}

func (s *standardTokenAuthorizer) requestToken(scopes []*Scope) (*models.Token, error) {
	realm := s.tokenServiceURL()
	return getToken(s.client, s.credential, realm, s.service, scopes)
}

// TODO
func parseScope(req *http.Request) (*Scope, error) {
	return nil, nil
}

func (s *standardTokenAuthorizer) ping(endpoint string) error {
	pingURL := buildPingURL(utils.FormatEndpoint(endpoint))
	resp, err := s.client.Get(pingURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	ping, err := url.Parse(pingURL)
	if err != nil {
		return err
	}
	s.registryURL = ping

	challenges := ParseChallengeFromResponse(resp)
	for _, challenge := range challenges {
		if scheme == challenge.Scheme {
			s.realm = challenge.Parameters["realm"]
			s.service = challenge.Parameters["service"]
			return nil
		}
	}
	return fmt.Errorf("%s is unsupportted", scheme)
}

// when the registry client is used inside Harbor, the token request
// can be posted to token service directly rather than going through nginx.
// If realm is set as the internal url of token service, this can resolve
// two problems:
// 1. performance issue
// 2. the realm field returned by registry is an IP which can not reachable
// inside Harbor
func (s *standardTokenAuthorizer) tokenServiceURL() string {
	if len(s.customizedTokenService) != 0 {
		return s.customizedTokenService
	}
	return s.realm
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}
