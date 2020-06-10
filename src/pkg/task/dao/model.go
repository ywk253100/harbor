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
	"time"

	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(&Execution{})
	orm.RegisterModel(&Task{})
	orm.RegisterModel(&CheckInData{})
}

// Execution database model
type Execution struct {
	ID   int64  `orm:"pk;auto;column(id)"`
	Type string `orm:"column(type)"`
	// In most of cases, the status should be calculated from the referenced tasks.
	// When the execution contains no task or failed to create tasks, the status should
	// be set manually
	Status        string    `orm:"column(status)"`
	StatusMessage string    `orm:"column(status_message)"`
	Trigger       string    `orm:"column(trigger)"`
	ExtraAttrs    string    `orm:"column(extra_attrs)"` // json string
	StartTime     time.Time `orm:"column(start_time)"`
	EndTime       time.Time `orm:"column(end_time)"`
}

// Task database model
type Task struct {
	ID             int64     `orm:"pk;auto;column(id)"`
	ExecutionID    int64     `orm:"column(execution_id)"`
	JobID          int64     `orm:"column(job_id)"`
	Status         string    `orm:"column(status)"`
	StatusCode     int       `orm:"column(status_code)"`
	StatusRevision int64     `orm:"column(status_revision)"`
	StatusMessage  string    `orm:"column(status_message)"`
	RetryCount     int       `orm:"column(retry_count)"`
	ExtraAttrs     string    `orm:"column(extra_attrs)"` // json string
	StartTime      time.Time `orm:"column(start_time)"`
	UpdateTime     time.Time `orm:"column(update_time)"`
	EndTime        time.Time `orm:"column(end_time)"`
}

// CheckInData database model
type CheckInData struct {
	ID           int64     `json:"id" orm:"pk;auto;column(id)"`
	TaskID       int64     `json:"task_id" orm:"column(task_id)"`
	Data         string    `json:"data" orm:"column(data)"`
	CreationTime time.Time `json:"creation_time" orm:"column(creation_time)"`
	UpdateTime   time.Time `json:"update_time" orm:"column(update_time)"`
}

// StatusCount model
type StatusCount struct {
	Status string `orm:"column(status)"`
	Count  int64  `orm:"column(count)"`
}
