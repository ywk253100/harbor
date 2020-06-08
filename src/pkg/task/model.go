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
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

// const definitions
const (
	ExecutionTypeReplication       = "REPLICATION"
	ExecutionTypeGarbageCollection = "GARBAGE_COLLECTION"
	ExecutionTypeRetention         = "RETENTION"
	ExecutionTypeScan              = "SCAN"
	ExecutionTypeScanAll           = "SCAN_ALL"
	ExecutionTypeScheduler         = "SCHEDULER"

	ExecutionTriggerManual   = "MANUAL"
	ExecutionTriggerSchedule = "SCHEDULE"
	ExecutionTriggerEvent    = "EVENT"

	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusScheduled = "SCHEDULED"
	StatusStopped   = "STOPPED"
)

// Execution is one run for one action. It contains one or more tasks and provides the summary view of the tasks
type Execution struct {
	ID int64 `json:"id"`
	// indicate the execution type: replication/GC/retention/scan/etc.
	Type   string `json:"type"`
	Status string `json:"status"`
	// the detail message to explain the status in some cases. e.g.
	// 1. After creating the execution, there may be some errors before creating tasks, the
	// "StatusMessage" can contain the error message
	// 2. The execution may contain no tasks, "StatusMessage" can be used to explain the case
	StatusMessage      string `json:"status_message"`
	TaskCount          int64  `json:"task_count"`
	SucceededTaskCount int64  `json:"succeeded_task_count"`
	FailedTaskCount    int64  `json:"failed_task_count"`
	PendingTaskCount   int64  `json:"pending_task_count"`
	RunningTaskCount   int64  `json:"running_task_count"`
	ScheduledTaskCount int64  `json:"scheduled_task_count"`
	StoppedTaskCount   int64  `json:"stopped_task_count"`
	// trigger type: manual/schedule/event
	Trigger string `json:"trigger"`
	// the customized attributes for different kinds of consumers
	ExtraAttrs map[string]interface{} `json:"extra_attrs"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
}

// Task is the unit for running. It stores the jobservice job records and related information
type Task struct {
	ID          int64  `json:"id"`
	ExecutionID int64  `json:"execution_id"`
	Status      string `json:"status"`
	// the detail message to explain the status in some cases. e.g.
	// When the job is failed to submit to jobservice, this field can be used to explain the reason
	StatusMessage string `json:"status_message"`
	RetryCount    int    `json:"retry_count"`
	// the customized attributes for different kinds of consumers
	ExtraAttrs   map[string]interface{} `json:"extra_attrs"`
	CheckInDatas []*dao.CheckInData     `json:"check_in_datas"`
	StartTime    time.Time              `json:"start_time"`
	UpdateTime   time.Time              `json:"update_time"`
	EndTime      time.Time              `json:"end_time"`
}

// Job is the model represents the requested jobservice job
type Job struct {
	Name              string
	Parameters        job.Parameters
	Metadata          *job.Metadata
	AppendCheckInData bool // true: store all data, false: the new data overrides the last one
}
