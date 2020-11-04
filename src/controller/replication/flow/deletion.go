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

package flow

import (
	"context"
	"encoding/json"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/replication/model"
)

type deletionFlow struct {
	executionID  int64
	policy       *model.Policy
	executionMgr task.ExecutionManager
	taskMgr      task.Manager
	resources    []*model.Resource
}

// NewDeletionFlow returns an instance of the delete flow which deletes the resources
// on the destination registry
func NewDeletionFlow(executionID int64, policy *model.Policy, resources ...*model.Resource) Flow {
	return &deletionFlow{
		executionMgr: task.ExecMgr,
		taskMgr:      task.Mgr,
		executionID:  executionID,
		policy:       policy,
		resources:    resources,
	}
}

func (d *deletionFlow) Run(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	srcResources, err := filterResources(d.resources, d.policy.Filters)
	if err != nil {
		return err
	}
	if len(srcResources) == 0 {
		// no candidates, mark the execution as done directly
		if err := d.executionMgr.MarkDone(ctx, d.executionID, "no resources need to be replicated"); err != nil {
			logger.Errorf("failed to mark done for the execution %d: %v", d.executionID, err)
		}
		return nil
	}

	srcResources = assembleSourceResources(srcResources, d.policy)
	dstResources := assembleDestinationResources(srcResources, d.policy)

	return d.createTasks(ctx, srcResources, dstResources)
}

func (d *deletionFlow) createTasks(ctx context.Context, srcResources, dstResources []*model.Resource) error {
	for i, resource := range srcResources {
		src, err := json.Marshal(resource)
		if err != nil {
			return err
		}
		dest, err := json.Marshal(dstResources[i])
		if err != nil {
			return err
		}

		job := &task.Job{
			Name: job.Replication,
			Metadata: &job.Metadata{
				JobKind: job.KindGeneric,
			},
			Parameters: map[string]interface{}{
				"src_resource": string(src),
				"dst_resource": string(dest),
			},
		}

		operation := "deletion"
		if dstResources[i].IsDeleteTag {
			operation = "tag deletion"
		}

		if _, err = d.taskMgr.Create(ctx, d.executionID, job, map[string]interface{}{
			"operation":            operation,
			"resource_type":        string(resource.Type),
			"source_resource":      getResourceName(resource),
			"destination_resource": getResourceName(dstResources[i])}); err != nil {
			return err
		}
	}
	return nil
}
