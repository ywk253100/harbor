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

package artifact

import (
	"testing"
	"time"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/project"
	rep "github.com/goharbor/harbor/src/controller/replication"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/model"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	replicationtesting "github.com/goharbor/harbor/src/testing/controller/replication"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedNotificationPolicyMgr struct {
}

type fakedReplicationPolicyMgr struct {
}

type fakedReplicationRegistryMgr struct {
}

func (f *fakedNotificationPolicyMgr) Create(*models.NotificationPolicy) (int64, error) {
	return 0, nil
}

// List the policies, returns the policy list and error
func (f *fakedNotificationPolicyMgr) List(int64) ([]*models.NotificationPolicy, error) {
	return nil, nil
}

// Get policy with specified ID
func (f *fakedNotificationPolicyMgr) Get(int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

// GetByNameAndProjectID get policy by the name and projectID
func (f *fakedNotificationPolicyMgr) GetByNameAndProjectID(string, int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

// Update the specified policy
func (f *fakedNotificationPolicyMgr) Update(*models.NotificationPolicy) error {
	return nil
}

// Delete the specified policy
func (f *fakedNotificationPolicyMgr) Delete(int64) error {
	return nil
}

// Test the specified policy
func (f *fakedNotificationPolicyMgr) Test(*models.NotificationPolicy) error {
	return nil
}

// GetRelatedPolices get event type related policies in project
func (f *fakedNotificationPolicyMgr) GetRelatedPolices(int64, string) ([]*models.NotificationPolicy, error) {
	return []*models.NotificationPolicy{
		{
			ID: 0,
		},
	}, nil
}

// Create new policy
func (f *fakedReplicationPolicyMgr) Create(*model.Policy) (int64, error) {
	return 0, nil
}

// List the policies, returns the total count, policy list and error
func (f *fakedReplicationPolicyMgr) List(...*model.PolicyQuery) (int64, []*model.Policy, error) {
	return 0, nil, nil
}

// Get policy with specified ID
func (f *fakedReplicationPolicyMgr) Get(int64) (*model.Policy, error) {
	return &model.Policy{
		ID: 1,
		SrcRegistry: &model.Registry{
			ID: 0,
		},
		DestRegistry: &model.Registry{
			ID: 0,
		},
	}, nil
}

// Get policy by the name
func (f *fakedReplicationPolicyMgr) GetByName(string) (*model.Policy, error) {
	return nil, nil
}

// Update the specified policy
func (f *fakedReplicationPolicyMgr) Update(policy *model.Policy) error {
	return nil
}

// Remove the specified policy
func (f *fakedReplicationPolicyMgr) Remove(int64) error {
	return nil
}

// Add new registry
func (f *fakedReplicationRegistryMgr) Add(*model.Registry) (int64, error) {
	return 0, nil
}

// List registries, returns total count, registry list and error
func (f *fakedReplicationRegistryMgr) List(query *q.Query) (int64, []*model.Registry, error) {
	return 0, nil, nil
}

// Get the specified registry
func (f *fakedReplicationRegistryMgr) Get(int64) (*model.Registry, error) {
	return &model.Registry{
		Type: "harbor",
		Credential: &model.Credential{
			Type: "local",
		},
	}, nil
}

// GetByName gets registry by name
func (f *fakedReplicationRegistryMgr) GetByName(name string) (*model.Registry, error) {
	return nil, nil
}

// Update the registry, the "props" are the properties of registry
// that need to be updated
func (f *fakedReplicationRegistryMgr) Update(registry *model.Registry, props ...string) error {
	return nil
}

// Remove the registry with the specified ID
func (f *fakedReplicationRegistryMgr) Remove(int64) error {
	return nil
}

// HealthCheck checks health status of all registries and update result in database
func (f *fakedReplicationRegistryMgr) HealthCheck() error {
	return nil
}

func TestReplicationHandler_Handle(t *testing.T) {
	common_dao.PrepareTestForPostgresSQL()
	config.Init()

	PolicyMgr := notification.PolicyMgr
	rpPolicy := replication.PolicyCtl
	rpRegistry := replication.RegistryMgr
	prj := project.Ctl
	repCtl := rep.Ctl

	defer func() {
		notification.PolicyMgr = PolicyMgr
		replication.PolicyCtl = rpPolicy
		replication.RegistryMgr = rpRegistry
		project.Ctl = prj
		rep.Ctl = repCtl
	}()
	notification.PolicyMgr = &fakedNotificationPolicyMgr{}
	replication.PolicyCtl = &fakedReplicationPolicyMgr{}
	replication.RegistryMgr = &fakedReplicationRegistryMgr{}
	projectCtl := &projecttesting.Controller{}
	project.Ctl = projectCtl
	mockRepCtl := &replicationtesting.Controller{}
	rep.Ctl = mockRepCtl
	mockRepCtl.On("GetTask", mock.Anything, mock.Anything).Return(&rep.Task{}, nil)
	mockRepCtl.On("GetExecution", mock.Anything, mock.Anything).Return(&rep.Execution{}, nil)

	mock.OnAnything(projectCtl, "GetByName").Return(&models.Project{ProjectID: 1}, nil)

	handler := &ReplicationHandler{}

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ReplicationHandler Want Error 1",
			args: args{
				data: "",
			},
			wantErr: true,
		},
		{
			name: "ReplicationHandler 1",
			args: args{
				data: &event.ReplicationEvent{
					OccurAt: time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "ReplicationHandler with docker registry",
			args: args{
				data: &event.ReplicationEvent{
					OccurAt:           time.Now(),
					ReplicationTaskID: 1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Nil(t, err)
		})
	}

}

func TestReplicationHandler_IsStateful(t *testing.T) {
	handler := &ReplicationHandler{}
	assert.False(t, handler.IsStateful())
}

func TestReplicationHandler_Name(t *testing.T) {
	handler := &ReplicationHandler{}
	assert.Equal(t, "ReplicationWebhook", handler.Name())
}
