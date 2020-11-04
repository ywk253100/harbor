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

package scheduler

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/task"
)

var (
	registry = make(map[string]CallbackFunc)
)

// CallbackFunc defines the function that the scheduler calls when triggered
type CallbackFunc func(interface{}) error

func init() {
	if err := task.RegisterCheckInProcessor(JobNameScheduler, triggerCallback); err != nil {
		log.Errorf("failed to register check in processor for scheduler: %v", err)
	}
}

// RegisterCallbackFunc registers the callback function which will be called when the scheduler is triggered
func RegisterCallbackFunc(name string, callbackFunc CallbackFunc) error {
	if len(name) == 0 {
		return errors.New("empty name")
	}
	if callbackFunc == nil {
		return errors.New("callback function is nil")
	}

	_, exist := registry[name]
	if exist {
		return fmt.Errorf("callback function %s already exists", name)
	}
	registry[name] = callbackFunc

	return nil
}

func getCallbackFunc(name string) (CallbackFunc, error) {
	f, exist := registry[name]
	if !exist {
		return nil, fmt.Errorf("callback function %s not found", name)
	}
	return f, nil
}

func callbackFuncExist(name string) bool {
	_, exist := registry[name]
	return exist
}

func triggerCallback(ctx context.Context, task *task.Task, data string) (err error) {
	execution, err := Sched.(*scheduler).execMgr.Get(ctx, task.ExecutionID)
	if err != nil {
		return err
	}
	if execution.VendorType != JobNameScheduler {
		return fmt.Errorf("the vendor type of execution %d isn't %s: %s",
			task.ExecutionID, JobNameScheduler, execution.VendorType)
	}
	schedule, err := Sched.(*scheduler).dao.Get(ctx, execution.VendorID)
	if err != nil {
		return err
	}
	callbackFunc, err := getCallbackFunc(schedule.CallbackFuncName)
	if err != nil {
		return err
	}
	return callbackFunc(schedule.CallbackFuncParam)
}
