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

package replication

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/controller/replication/flow"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/replication/model"
)

// Controller defines the operations related with replication
type Controller interface {
	// Start the replication according to the policy
	Start(ctx context.Context, policy *model.Policy, resource *model.Resource, trigger string) (executionID int64, err error)
	// Stop the replication specified by the execution ID
	Stop(ctx context.Context, executionID int64) (err error)
	// ExecutionCount returns the total count of executions according to the query
	ExecutionCount(ctx context.Context, query *q.Query) (count int64, err error)
	// ListExecutions lists the executions according to the query
	ListExecutions(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// GetExecution gets the specific execution
	GetExecution(ctx context.Context, executionID int64) (execution *Execution, err error)
	// TaskCount returns the total count of tasks according to the query
	TaskCount(ctx context.Context, query *q.Query) (count int64, err error)
	// ListTasks lists the tasks according to the query
	ListTasks(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// GetTask gets the specific task
	GetTask(ctx context.Context, taskID int64) (task *Task, err error)
	// GetTaskLog gets the log of the specific task
	GetTaskLog(ctx context.Context, taskID int64) (log []byte, err error)
}

var (
	// Ctl is a global replication controller instance
	Ctl            = NewController()
	_   Controller = &controller{}
)

// NewController creates a new instance of the replication controller
func NewController() Controller {
	return &controller{
		execMgr: task.ExecMgr,
		taskMgr: task.Mgr,
		flowCtl: flow.NewController(),
	}
}

type controller struct {
	execMgr task.ExecutionManager
	taskMgr task.Manager
	flowCtl flow.Controller
}

func (c *controller) Start(ctx context.Context, policy *model.Policy, resource *model.Resource, trigger string) (int64, error) {
	logger := log.GetLogger(ctx)
	if !policy.Enabled {
		return 0, errors.New(nil).WithCode(errors.PreconditionCode).
			WithMessage("the policy %d is disabled", policy.ID)
	}
	// create an execution record
	id, err := c.execMgr.Create(ctx, job.Replication, policy.ID, trigger)
	if err != nil {
		return 0, err
	}
	// start the replication flow in background
	go func() {
		err := c.flowCtl.Start(ctx, id, policy, resource)
		if err == nil {
			// no err, return directly
			return
		}
		// got error, try to stop the execution first in case that some tasks are already created
		if err := c.execMgr.StopAndWait(ctx, id, 10*time.Second); err != nil {
			logger.Errorf("failed to stop the execution %d: %v", id, err)
		}
		if err := c.execMgr.MarkError(ctx, id, err.Error()); err != nil {
			logger.Errorf("failed to mark error for the execution %d: %v", id, err)
		}
	}()
	return id, nil
}

func (c *controller) Stop(ctx context.Context, id int64) error {
	return c.execMgr.Stop(ctx, id)
}

func (c *controller) ExecutionCount(ctx context.Context, query *q.Query) (int64, error) {
	return c.execMgr.Count(ctx, query)
}

func (c *controller) ListExecutions(ctx context.Context, query *q.Query) ([]*Execution, error) {
	// as the following logic may change the content of the query, clone it first
	query = q.MustClone(query)
	// convert the query keyword "PolicyID" or "policy_id" to the "VendorID"
	if value, exist := query.Keywords["PolicyID"]; exist {
		query.Keywords["VendorID"] = value
		delete(query.Keywords, "PolicyID")
	}
	if value, exist := query.Keywords["policy_id"]; exist {
		query.Keywords["VendorID"] = value
		delete(query.Keywords, "policy_id")
	}

	execs, err := c.execMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var executions []*Execution
	for _, exec := range execs {
		executions = append(executions, convertExecution(exec))
	}
	return executions, nil
}

func (c *controller) GetExecution(ctx context.Context, id int64) (*Execution, error) {
	exec, err := c.execMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertExecution(exec), nil
}

func (c *controller) TaskCount(ctx context.Context, query *q.Query) (int64, error) {
	return c.taskMgr.Count(ctx, query)
}

func (c *controller) ListTasks(ctx context.Context, query *q.Query) ([]*Task, error) {
	tks, err := c.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var tasks []*Task
	for _, tk := range tks {
		tasks = append(tasks, convertTask(tk))
	}
	return tasks, nil
}

func (c *controller) GetTask(ctx context.Context, id int64) (*Task, error) {
	tk, err := c.taskMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertTask(tk), nil
}

func (c *controller) GetTaskLog(ctx context.Context, id int64) ([]byte, error) {
	return c.taskMgr.GetLog(ctx, id)
}

func convertExecution(exec *task.Execution) *Execution {
	return &Execution{
		ID:            exec.ID,
		PolicyID:      exec.VendorID,
		Status:        exec.Status,
		StatusMessage: exec.StatusMessage,
		Metrics:       exec.Metrics,
		Trigger:       exec.Trigger,
		StartTime:     exec.StartTime,
		EndTime:       exec.EndTime,
	}
}

func convertTask(task *task.Task) *Task {
	return &Task{
		ID:                  task.ID,
		ExecutionID:         task.ExecutionID,
		Status:              task.Status,
		StatusMessage:       task.StatusMessage,
		RunCount:            task.RunCount,
		ResourceType:        task.GetStringFromExtraAttrs("resource_type"),
		SourceResource:      task.GetStringFromExtraAttrs("source_resource"),
		DestinationResource: task.GetStringFromExtraAttrs("destination_resource"),
		Operation:           task.GetStringFromExtraAttrs("operation"),
		JobID:               task.JobID,
		CreationTime:        task.CreationTime,
		StartTime:           task.StartTime,
		UpdateTime:          task.UpdateTime,
		EndTime:             task.EndTime,
	}
}
