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

package core

import (
	"fmt"
	"strings"

	"github.com/vmware/harbor/src/common/http/client"
	"github.com/vmware/harbor/src/common/http/client/auth"
	common_models "github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice/api"
<<<<<<< HEAD
	"github.com/vmware/harbor/src/jobservice/client"
=======
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	"github.com/vmware/harbor/src/replication"
	"github.com/vmware/harbor/src/replication/models"
	"github.com/vmware/harbor/src/replication/policy"
	"github.com/vmware/harbor/src/replication/replicator"
	"github.com/vmware/harbor/src/replication/source"
	"github.com/vmware/harbor/src/replication/target"
	"github.com/vmware/harbor/src/replication/trigger"
	"github.com/vmware/harbor/src/ui/config"
)

// Controller defines the methods that a replicatoin controllter should implement
type Controller interface {
	policy.Manager
	Init() error
	Replicate(policyID int64, metadata ...map[string]interface{}) error
}

//DefaultController is core module to cordinate and control the overall workflow of the
//replication modules.
type DefaultController struct {
	//Indicate whether the controller has been initialized or not
	initialized bool

	//Manage the policies
	policyManager policy.Manager

	//Manage the targets
	targetManager target.Manager

	//Handle the things related with source
	sourcer *source.Sourcer

	//Manage the triggers of policies
	triggerManager *trigger.Manager

	//Handle the replication work
	replicator replicator.Replicator
}

//Keep controller as singleton instance
var (
	GlobalController Controller = NewDefaultController(ControllerConfig{}) //Use default data
)

//ControllerConfig includes related configurations required by the controller
type ControllerConfig struct {
	//The capacity of the cache storing enabled triggers
	CacheCapacity int
}

//NewDefaultController is the constructor of DefaultController.
func NewDefaultController(cfg ControllerConfig) *DefaultController {
	//Controller refer the default instances
	ctl := &DefaultController{
		policyManager:  policy.NewDefaultManager(),
		targetManager:  target.NewDefaultManager(),
		sourcer:        source.NewSourcer(),
		triggerManager: trigger.NewManager(cfg.CacheCapacity),
	}

<<<<<<< HEAD
	// TODO read from configuration
	endpoint := "http://jobservice:8080"
	ctl.replicator = replicator.NewDefaultReplicator(endpoint,
		&client.Config{
			Secret: config.UISecret(),
		})
=======
	endpoint := config.InternalJobServiceURL()
	client := client.NewAuthorizedClient(auth.NewSecretAuthorizer(config.UISecret()))
	ctl.replicator = replicator.NewDefaultReplicator(endpoint, client)
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice

	return ctl
}

//Init will initialize the controller and the sub components
func (ctl *DefaultController) Init() error {
	if ctl.initialized {
		return nil
	}

	//Build query parameters
	query := models.QueryParameter{
		TriggerType: replication.TriggerKindSchedule,
	}

	policies, err := ctl.policyManager.GetPolicies(query)
	if err != nil {
		return err
	}
	if policies != nil && len(policies) > 0 {
		for _, policy := range policies {
			if err := ctl.triggerManager.SetupTrigger(&policy); err != nil {
				//TODO: Log error
				fmt.Printf("Error: %s", err)
				//TODO:Update the status of policy
			}
		}
	}

	//Initialize sourcer
	ctl.sourcer.Init()

	ctl.initialized = true

	return nil
}

//CreatePolicy is used to create a new policy and enable it if necessary
func (ctl *DefaultController) CreatePolicy(newPolicy models.ReplicationPolicy) (int64, error) {
	id, err := ctl.policyManager.CreatePolicy(newPolicy)
	if err != nil {
		return 0, err
	}

	newPolicy.ID = id
	if err = ctl.triggerManager.SetupTrigger(&newPolicy); err != nil {
		return 0, err
	}

	return id, nil
}

//UpdatePolicy will update the policy with new content.
//Parameter updatedPolicy must have the ID of the updated policy.
func (ctl *DefaultController) UpdatePolicy(updatedPolicy models.ReplicationPolicy) error {
	// TODO check pre-conditions

	id := updatedPolicy.ID
	originPolicy, err := ctl.policyManager.GetPolicy(id)
	if err != nil {
		return err
	}

	if originPolicy.ID == 0 {
		return fmt.Errorf("policy %d not found", id)
	}

	reset := false
	if updatedPolicy.Trigger.Kind != originPolicy.Trigger.Kind {
		reset = true
	} else {
		switch updatedPolicy.Trigger.Kind {
		case replication.TriggerKindSchedule:
			if !originPolicy.Trigger.ScheduleParam.Equal(updatedPolicy.Trigger.ScheduleParam) {
				reset = true
			}
		case replication.TriggerKindImmediate:
			// Always reset immediate trigger as it is relevent with namespaces
			reset = true
		default:
			// manual trigger, no need to reset
		}
	}

	if err = ctl.policyManager.UpdatePolicy(updatedPolicy); err != nil {
		return err
	}

	if reset {
		if err = ctl.triggerManager.UnsetTrigger(&originPolicy); err != nil {
			return err
		}

		return ctl.triggerManager.SetupTrigger(&updatedPolicy)
	}

	return nil
}

//RemovePolicy will remove the specified policy and clean the related settings
func (ctl *DefaultController) RemovePolicy(policyID int64) error {
	// TODO check pre-conditions

	policy, err := ctl.policyManager.GetPolicy(policyID)
	if err != nil {
		return err
	}

	if policy.ID == 0 {
		return fmt.Errorf("policy %d not found", policyID)
	}

	if err = ctl.triggerManager.UnsetTrigger(&policy); err != nil {
		return err
	}

	return ctl.policyManager.RemovePolicy(policyID)
}

//GetPolicy is delegation of GetPolicy of Policy.Manager
func (ctl *DefaultController) GetPolicy(policyID int64) (models.ReplicationPolicy, error) {
	return ctl.policyManager.GetPolicy(policyID)
}

//GetPolicies is delegation of GetPoliciemodels.ReplicationPolicy{}s of Policy.Manager
func (ctl *DefaultController) GetPolicies(query models.QueryParameter) ([]models.ReplicationPolicy, error) {
	return ctl.policyManager.GetPolicies(query)
}

//Replicate starts one replication defined in the specified policy;
//Can be launched by the API layer and related triggers.
func (ctl *DefaultController) Replicate(policyID int64, metadata ...map[string]interface{}) error {
	policy, err := ctl.GetPolicy(policyID)
	if err != nil {
		return err
	}
	if policy.ID == 0 {
		return fmt.Errorf("policy %d not found", policyID)
	}

	// prepare candidates for replication
	candidates := getCandidates(&policy, ctl.sourcer, metadata...)

	/*
		targets := []*common_models.RepTarget{}
		for _, targetID := range policy.TargetIDs {
			target, err := ctl.targetManager.GetTarget(targetID)
			if err != nil {
				return err
			}
			targets = append(targets, target)
		}
	*/

	// submit the replication
	return replicate(ctl.replicator, policyID, candidates)
}

func getCandidates(policy *models.ReplicationPolicy, sourcer *source.Sourcer,
	metadata ...map[string]interface{}) []models.FilterItem {
	candidates := []models.FilterItem{}
	if len(metadata) > 0 {
		meta := metadata[0]["candidates"]
		if meta != nil {
			cands, ok := meta.([]models.FilterItem)
			if ok {
				candidates = append(candidates, cands...)
			}
		}
	}

<<<<<<< HEAD
=======
	// prepare candidates for replication
	candidates = getCandidates(&policy, ctl.sourcer, candidates...)

	// TODO
	/*
		targets := []*common_models.RepTarget{}
		for _, targetID := range policy.TargetIDs {
			target, err := ctl.targetManager.GetTarget(targetID)
			if err != nil {
				return err
			}
			targets = append(targets, target)
		}
	*/

	// TODO merge tags whose repository is same into one struct

	// submit the replication
	return replicate(ctl.replicator, policyID, candidates)
}

func getCandidates(policy *models.ReplicationPolicy, sourcer *source.Sourcer, candidates ...models.FilterItem) []models.FilterItem {
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	if len(candidates) == 0 {
		for _, namespace := range policy.Namespaces {
			candidates = append(candidates, models.FilterItem{
				Kind:      replication.FilterItemKindProject,
				Value:     namespace,
				Operation: common_models.RepOpTransfer,
			})
		}
	}

	filterChain := buildFilterChain(policy, sourcer)

	return filterChain.DoFilter(candidates)
}

func buildFilterChain(policy *models.ReplicationPolicy, sourcer *source.Sourcer) source.FilterChain {
	filters := []source.Filter{}

	patterns := map[string]string{}
	for _, f := range policy.Filters {
		patterns[f.Kind] = f.Pattern
	}

	registry := sourcer.GetAdaptor(replication.AdaptorKindHarbor)
	// only support repository and tag filter for now
	filters = append(filters,
		source.NewRepositoryFilter(patterns[replication.FilterItemKindRepository], registry))
	filters = append(filters,
		source.NewTagFilter(patterns[replication.FilterItemKindTag], registry))

	return source.NewDefaultFilterChain(filters)
}

func replicate(replicator replicator.Replicator, policyID int64, candidates []models.FilterItem) error {
<<<<<<< HEAD
	if len(candidates) == 0 {
		log.Debugf("replicaton candidates are null, no further action needed")
	}

=======
>>>>>>> a982d8f... Create replicator to submit replication job to jobservice
	repositories := map[string][]string{}
	// TODO the operation of all candidates are same for now. Update it after supporting
	// replicate deletion
	operation := ""
	for _, candidate := range candidates {
		strs := strings.SplitN(candidate.Value, ":", 2)
		repositories[strs[0]] = append(repositories[strs[0]], strs[1])
		operation = candidate.Operation
	}

	for repository, tags := range repositories {
		replication := &api.ReplicationReq{
			PolicyID:  policyID,
			Repo:      repository,
			Operation: operation,
			TagList:   tags,
		}
		log.Debugf("submiting replication job to jobservice: %v", replication)
		if err := replicator.Replicate(replication); err != nil {
			return err
		}
	}
	return nil
}
