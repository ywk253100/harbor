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

	"github.com/goharbor/harbor/src/lib/q"
)

// Manager provides an execution-task model to abstract the interactive with jobservice.
// All of the operations with jobservice should be delegated by this manager
type Manager interface {
	// CreateExecution creates an execution. The "typee" is used to specify the type(execution, scan, etc.)
	// of the execution, and the "extraAttrs" can be used to set the customized attributes
	CreateExecution(ctx context.Context, typee, trigger string, extraAttrs ...map[string]interface{}) (id int64, err error)
	// UpdateExecutionStatus updates the status of the execution.
	// In most cases, the execution status can be calculated from the referenced tasks automatically.
	// When the execution contains no tasks or failed to create tasks, the status should be set manually
	UpdateExecutionStatus(ctx context.Context, id int64, status, message string) (err error)
	// StopExecution stops all linked tasks of the specified execution
	StopExecution(ctx context.Context, id int64) (err error)
	// DeleteExecution deletes the specified execution and its tasks
	DeleteExecution(ctx context.Context, id int64) (err error)
	// GetExecution gets the specified execution
	GetExecution(ctx context.Context, id int64) (execution *Execution, err error)
	// ListExecutions lists executions according to the query
	ListExecutions(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// CreateTask submits the job to jobservice and creates a corresponding task record.
	// An execution must be created first and the task will be linked to it.
	// The "extraAttrs" can be used to set the customized attributes
	CreateTask(ctx context.Context, executionID int64, job *Job, extraAttrs ...map[string]interface{}) (id int64, err error)
	// StopTask stops the specified task
	StopTask(ctx context.Context, id int64) (err error)
	// GetTask gets the specified task
	GetTask(ctx context.Context, id int64) (task *Task, err error)
	// ListTasks lists the tasks according to the query
	ListTasks(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// GetTaskLog gets log of the specified task
	GetTaskLog(ctx context.Context, id int64) (log []byte, err error)
}
