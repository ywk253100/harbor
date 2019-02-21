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

// Package ng ...
// TODO rename the package name after removing ng
package ng

import (
	"fmt"

	"github.com/goharbor/harbor/src/replication/ng/scheduler"

	"github.com/goharbor/harbor/src/replication/ng/execution"
	"github.com/goharbor/harbor/src/replication/ng/flow"
	"github.com/goharbor/harbor/src/replication/ng/operation"
	"github.com/goharbor/harbor/src/replication/ng/registry"
)

var (
	// RegistryMgr is a global registry manager
	RegistryMgr registry.Manager
	// ExecutionMgr is a global execution manager
	ExecutionMgr execution.Manager
	// OperationCtl is a global operation controller
	OperationCtl operation.Controller
)

// Init the global variables
func Init() error {
	// TODO init RegistryMgr

	// TODO init ExecutionMgr

	// TODO init scheduler
	var scheduler scheduler.Scheduler

	flowCtl, err := flow.NewController(RegistryMgr, ExecutionMgr, scheduler)
	if err != nil {
		return fmt.Errorf("failed to create the flow controller: %v", err)
	}
	OperationCtl = operation.NewController(flowCtl, ExecutionMgr)
	return nil
}
