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

package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

var (
	// ExecMgr is a global execution manager instance
	ExecMgr = NewExecutionManager()
)

// ExecutionManager manages executions.
// The execution and task managers provide an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by them
type ExecutionManager interface {
	// Create an execution. The "vendorType" specifies the type of vendor (e.g. replication, scan, gc, retention, etc.),
	// and the "vendorID" specifies the ID of vendor if needed(e.g. policy ID for replication and retention).
	// The "extraAttrs" can be used to set the customized attributes
	Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
		extraAttrs ...map[string]interface{}) (id int64, err error)
	// MarkDone marks the status of the specified execution as success.
	// It must be called to update the execution status if the created execution contains no tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly
	MarkDone(ctx context.Context, id int64, message string) (err error)
	// MarkError marks the status of the specified execution as error.
	// It must be called to update the execution status when failed to create tasks.
	// In other cases, the execution status can be calculated from the referenced tasks automatically
	// and no need to update it explicitly
	MarkError(ctx context.Context, id int64, message string) (err error)
	// Stop all linked tasks of the specified execution
	Stop(ctx context.Context, id int64) (err error)
	// StopAndWait stops all linked tasks of the specified execution and waits until all tasks are stopped
	// or get an error
	StopAndWait(ctx context.Context, id int64, timeout time.Duration) (err error)
	// Delete the specified execution and its tasks
	Delete(ctx context.Context, id int64) (err error)
	// Get the specified execution
	Get(ctx context.Context, id int64) (execution *Execution, err error)
	// List executions according to the query
	List(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// Count counts total.
	Count(ctx context.Context, query *q.Query) (int64, error)
}

// NewExecutionManager return an instance of the default execution manager
func NewExecutionManager() ExecutionManager {
	return &executionManager{
		executionDAO: dao.NewExecutionDAO(),
		taskMgr:      Mgr,
		taskDAO:      dao.NewTaskDAO(),
	}
}

type executionManager struct {
	executionDAO dao.ExecutionDAO
	taskMgr      Manager
	taskDAO      dao.TaskDAO
}

func (e *executionManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return e.executionDAO.Count(ctx, query)
}

func (e *executionManager) Create(ctx context.Context, vendorType string, vendorID int64, trigger string,
	extraAttrs ...map[string]interface{}) (int64, error) {
	extras := map[string]interface{}{}
	if len(extraAttrs) > 0 && extraAttrs[0] != nil {
		extras = extraAttrs[0]
	}
	data, err := json.Marshal(extras)
	if err != nil {
		return 0, err
	}

	execution := &dao.Execution{
		VendorType: vendorType,
		VendorID:   vendorID,
		Status:     job.RunningStatus.String(),
		Trigger:    trigger,
		ExtraAttrs: string(data),
		StartTime:  time.Now(),
	}
	return e.executionDAO.Create(ctx, execution)
}

func (e *executionManager) MarkDone(ctx context.Context, id int64, message string) error {
	return e.executionDAO.Update(ctx, &dao.Execution{
		ID:            id,
		Status:        job.SuccessStatus.String(),
		StatusMessage: message,
		EndTime:       time.Now(),
	}, "Status", "StatusMessage", "EndTime")
}

func (e *executionManager) MarkError(ctx context.Context, id int64, message string) error {
	return e.executionDAO.Update(ctx, &dao.Execution{
		ID:            id,
		Status:        job.ErrorStatus.String(),
		StatusMessage: message,
		EndTime:       time.Now(),
	}, "Status", "StatusMessage", "EndTime")
}

func (e *executionManager) Stop(ctx context.Context, id int64) error {
	execution, err := e.executionDAO.Get(ctx, id)
	if err != nil {
		return err
	}

	// when an execution is in final status, if it contains task that is a periodic or retrying job it will
	// run again in the near future, so we must operate the stop action
	tasks, err := e.taskDAO.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID": id,
		},
	})
	if err != nil {
		return err
	}
	// contains no task and the status isn't final, update the status to stop directly
	if len(tasks) == 0 && !job.Status(execution.Status).Final() {
		return e.executionDAO.Update(ctx, &dao.Execution{
			ID:      id,
			Status:  job.StoppedStatus.String(),
			EndTime: time.Now(),
		}, "Status", "EndTime")
	}

	for _, task := range tasks {
		if err = e.taskMgr.Stop(ctx, task.ID); err != nil {
			log.Errorf("failed to stop task %d: %v", task.ID, err)
			continue
		}
	}
	return nil
}

func (e *executionManager) StopAndWait(ctx context.Context, id int64, timeout time.Duration) error {
	var (
		overtime bool
		errChan  = make(chan error)
		lock     = sync.RWMutex{}
	)
	go func() {
		// stop the execution
		if err := e.Stop(ctx, id); err != nil {
			errChan <- err
			return
		}
		// check the status of the execution
		interval := 100 * time.Millisecond
		stop := false
		for !stop {
			execution, err := e.executionDAO.Get(ctx, id)
			if err != nil {
				errChan <- err
				return
			}
			// if the status is final, return
			if job.Status(execution.Status).Final() {
				errChan <- nil
				return
			}
			time.Sleep(interval)
			if interval < 1*time.Second {
				interval = interval * 2
			}
			lock.RLock()
			stop = overtime
			lock.RUnlock()
		}
	}()

	select {
	case <-time.After(timeout):
		lock.Lock()
		overtime = true
		lock.Unlock()
		return fmt.Errorf("stopping the execution %d timeout", id)
	case err := <-errChan:
		return err
	}
}

func (e *executionManager) Delete(ctx context.Context, id int64) error {
	tasks, err := e.taskDAO.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID": id,
		},
	})
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if !job.Status(task.Status).Final() {
			return errors.New(nil).WithCode(errors.PreconditionCode).
				WithMessage("the execution %d has tasks that aren't in final status, stop the tasks first", id)
		}
		if err = e.taskDAO.Delete(ctx, task.ID); err != nil {
			return err
		}
	}

	return e.executionDAO.Delete(ctx, id)
}

func (e *executionManager) Get(ctx context.Context, id int64) (*Execution, error) {
	execution, err := e.executionDAO.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return e.populateExecution(ctx, execution), nil
}

func (e *executionManager) List(ctx context.Context, query *q.Query) ([]*Execution, error) {
	executions, err := e.executionDAO.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var execs []*Execution
	for _, execution := range executions {
		execs = append(execs, e.populateExecution(ctx, execution))
	}
	return execs, nil
}

func (e *executionManager) populateExecution(ctx context.Context, execution *dao.Execution) *Execution {
	exec := &Execution{
		ID:            execution.ID,
		VendorType:    execution.VendorType,
		VendorID:      execution.VendorID,
		Status:        execution.Status,
		StatusMessage: execution.StatusMessage,
		Metrics:       nil,
		Trigger:       execution.Trigger,
		StartTime:     execution.StartTime,
		EndTime:       execution.EndTime,
	}

	if len(execution.ExtraAttrs) > 0 {
		extras := map[string]interface{}{}
		if err := json.Unmarshal([]byte(execution.ExtraAttrs), &extras); err != nil {
			log.Errorf("failed to unmarshal the extra attributes of execution %d: %v", execution.ID, err)
		} else {
			exec.ExtraAttrs = extras
		}
	}

	// populate task metrics
	metrics, err := e.executionDAO.GetMetrics(ctx, execution.ID)
	if err != nil {
		log.Errorf("failed to get metrics of the execution %d: %v", execution.ID, err)
	} else {
		exec.Metrics = metrics
	}

	return exec
}
