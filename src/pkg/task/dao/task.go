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

package dao

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

// TaskDAO is the data access object interface for task
type TaskDAO interface {
	// Count returns the total count of tasks according to the query
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List the tasks according to the query
	List(ctx context.Context, query *q.Query) (tasks []*Task, err error)
	// Get the specified task
	Get(ctx context.Context, id int64) (task *Task, err error)
	// Create a task
	Create(ctx context.Context, task *Task) (id int64, err error)
	// Update the specified task. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, task *Task, props ...string) (err error)
	// Delete the specified task
	Delete(ctx context.Context, id int64) (err error)
	// ListStatusCount lists the status count for the tasks reference the specified execution
	ListStatusCount(ctx context.Context, executionID int64) (statusCounts []*StatusCount, err error)
	// GetMaxEndTime gets the max end time for the tasks references the specified execution
	GetMaxEndTime(ctx context.Context, executionID int64) (endTime time.Time, err error)
}

// NewTaskDAO returns an instance of TaskDAO
func NewTaskDAO() TaskDAO {
	return &taskDAO{}
}

type taskDAO struct{}

func (t *taskDAO) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	qs, err := orm.QuerySetter(ctx, &Task{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (t *taskDAO) List(ctx context.Context, query *q.Query) ([]*Task, error) {
	tasks := []*Task{}
	qs, err := orm.QuerySetter(ctx, &Task{}, query)
	if err != nil {
		return nil, err
	}
	qs = qs.OrderBy("-StartTime")
	if _, err = qs.All(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (t *taskDAO) Get(ctx context.Context, id int64) (*Task, error) {
	task := &Task{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(task); err != nil {
		if e := orm.AsNotFoundError(err, "task %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return task, nil
}

func (t *taskDAO) Create(ctx context.Context, task *Task) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(task)
	if err != nil {
		if e := orm.AsForeignKeyError(err,
			"the task tries to reference a non existing execution %d", task.ExecutionID); e != nil {
			err = e
		}
		return 0, err
	}
	return id, nil
}

func (t *taskDAO) Update(ctx context.Context, task *Task, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(task, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("task %d not found", task.ID)
	}
	return nil
}

func (t *taskDAO) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&Task{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("task %d not found", id)
	}
	return nil
}

func (t *taskDAO) ListStatusCount(ctx context.Context, executionID int64) ([]*StatusCount, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	statusCounts := []*StatusCount{}
	_, err = ormer.Raw("select status, count(*) as count from task where execution_id=? group by status", executionID).
		QueryRows(&statusCounts)
	if err != nil {
		return nil, err
	}
	return statusCounts, nil
}

func (t *taskDAO) GetMaxEndTime(ctx context.Context, executionID int64) (time.Time, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return time.Time{}, err
	}
	var endTime time.Time
	err = ormer.Raw("select max(end_time) from task where execution_id = ?", executionID).
		QueryRow(&endTime)
	return endTime, nil
}
