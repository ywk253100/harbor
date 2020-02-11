// Copyright 2018 Project Harbor Authors
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

package route

import (
	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/config"
)

// RegisterRoutes for Harbor legacy APIs
// TODO bump up the version of APIs called by clients
func registerLegacyRoutes() {
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/members/?:pmid([0-9]+)", &api.ProjectMemberAPI{})
	beego.Router("/api/"+APIVersion+"/projects/", &api.ProjectAPI{}, "head:Head")
	beego.Router("/api/"+APIVersion+"/projects/:id([0-9]+)", &api.ProjectAPI{})
	beego.Router("/api/"+APIVersion+"/users/:id", &api.UserAPI{}, "get:Get;delete:Delete;put:Put")
	beego.Router("/api/"+APIVersion+"/users", &api.UserAPI{}, "get:List;post:Post")
	beego.Router("/api/"+APIVersion+"/users/search", &api.UserAPI{}, "get:Search")
	beego.Router("/api/"+APIVersion+"/users/:id([0-9]+)/password", &api.UserAPI{}, "put:ChangePassword")
	beego.Router("/api/"+APIVersion+"/users/:id/permissions", &api.UserAPI{}, "get:ListUserPermissions")
	beego.Router("/api/"+APIVersion+"/users/:id/sysadmin", &api.UserAPI{}, "put:ToggleUserAdminRole")
	beego.Router("/api/"+APIVersion+"/users/:id/cli_secret", &api.UserAPI{}, "put:SetCLISecret")
	beego.Router("/api/"+APIVersion+"/usergroups/?:ugid([0-9]+)", &api.UserGroupAPI{})
	beego.Router("/api/"+APIVersion+"/ldap/ping", &api.LdapAPI{}, "post:Ping")
	beego.Router("/api/"+APIVersion+"/ldap/users/search", &api.LdapAPI{}, "get:Search")
	beego.Router("/api/"+APIVersion+"/ldap/groups/search", &api.LdapAPI{}, "get:SearchGroup")
	beego.Router("/api/"+APIVersion+"/ldap/users/import", &api.LdapAPI{}, "post:ImportUser")
	beego.Router("/api/"+APIVersion+"/email/ping", &api.EmailAPI{}, "post:Ping")
	beego.Router("/api/"+APIVersion+"/health", &api.HealthAPI{}, "get:CheckHealth")
	beego.Router("/api/"+APIVersion+"/ping", &api.SystemInfoAPI{}, "get:Ping")
	beego.Router("/api/"+APIVersion+"/search", &api.SearchAPI{})
	beego.Router("/api/"+APIVersion+"/projects/", &api.ProjectAPI{}, "get:List;post:Post")
	beego.Router("/api/"+APIVersion+"/projects/:id([0-9]+)/summary", &api.ProjectAPI{}, "get:Summary")
	beego.Router("/api/"+APIVersion+"/projects/:id([0-9]+)/logs", &api.ProjectAPI{}, "get:Logs")
	beego.Router("/api/"+APIVersion+"/projects/:id([0-9]+)/_deletable", &api.ProjectAPI{}, "get:Deletable")
	beego.Router("/api/"+APIVersion+"/projects/:id([0-9]+)/metadatas/?:name", &api.MetadataAPI{}, "get:Get")
	beego.Router("/api/"+APIVersion+"/projects/:id([0-9]+)/metadatas/", &api.MetadataAPI{}, "post:Post")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/robots", &api.RobotAPI{}, "post:Post;get:List")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/robots/:id([0-9]+)", &api.RobotAPI{}, "get:Get;put:Put;delete:Delete")

	beego.Router("/api/"+APIVersion+"/quotas", &api.QuotaAPI{}, "get:List")
	beego.Router("/api/"+APIVersion+"/quotas/:id([0-9]+)", &api.QuotaAPI{}, "get:Get;put:Put")

	beego.Router("/api/"+APIVersion+"/system/gc", &api.GCAPI{}, "get:List")
	beego.Router("/api/"+APIVersion+"/system/gc/:id", &api.GCAPI{}, "get:GetGC")
	beego.Router("/api/"+APIVersion+"/system/gc/:id([0-9]+)/log", &api.GCAPI{}, "get:GetLog")
	beego.Router("/api/"+APIVersion+"/system/gc/schedule", &api.GCAPI{}, "get:Get;put:Put;post:Post")
	beego.Router("/api/"+APIVersion+"/system/scanAll/schedule", &api.ScanAllAPI{}, "get:Get;put:Put;post:Post")
	beego.Router("/api/"+APIVersion+"/system/CVEWhitelist", &api.SysCVEWhitelistAPI{}, "get:Get;put:Put")
	beego.Router("/api/"+APIVersion+"/system/oidc/ping", &api.OIDCAPI{}, "post:Ping")

	beego.Router("/api/"+APIVersion+"/logs", &api.LogAPI{})

	beego.Router("/api/"+APIVersion+"/replication/adapters", &api.ReplicationAdapterAPI{}, "get:List")
	beego.Router("/api/"+APIVersion+"/replication/adapterinfos", &api.ReplicationAdapterAPI{}, "get:ListAdapterInfos")
	beego.Router("/api/"+APIVersion+"/replication/executions", &api.ReplicationOperationAPI{}, "get:ListExecutions;post:CreateExecution")
	beego.Router("/api/"+APIVersion+"/replication/executions/:id([0-9]+)", &api.ReplicationOperationAPI{}, "get:GetExecution;put:StopExecution")
	beego.Router("/api/"+APIVersion+"/replication/executions/:id([0-9]+)/tasks", &api.ReplicationOperationAPI{}, "get:ListTasks")
	beego.Router("/api/"+APIVersion+"/replication/executions/:id([0-9]+)/tasks/:tid([0-9]+)/log", &api.ReplicationOperationAPI{}, "get:GetTaskLog")
	beego.Router("/api/"+APIVersion+"/replication/policies", &api.ReplicationPolicyAPI{}, "get:List;post:Create")
	beego.Router("/api/"+APIVersion+"/replication/policies/:id([0-9]+)", &api.ReplicationPolicyAPI{}, "get:Get;put:Update;delete:Delete")

	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/webhook/policies", &api.NotificationPolicyAPI{}, "get:List;post:Post")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/webhook/policies/:id([0-9]+)", &api.NotificationPolicyAPI{})
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/webhook/policies/test", &api.NotificationPolicyAPI{}, "post:Test")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/webhook/lasttrigger", &api.NotificationPolicyAPI{}, "get:ListGroupByEventType")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/webhook/jobs/", &api.NotificationJobAPI{}, "get:List")

	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/immutabletagrules", &api.ImmutableTagRuleAPI{}, "get:List;post:Post")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/immutabletagrules/:id([0-9]+)", &api.ImmutableTagRuleAPI{})

	beego.Router("/api/"+APIVersion+"/configurations", &api.ConfigAPI{}, "get:Get;put:Put")
	beego.Router("/api/"+APIVersion+"/statistics", &api.StatisticAPI{})
	beego.Router("/api/"+APIVersion+"/labels", &api.LabelAPI{}, "post:Post;get:List")
	beego.Router("/api/"+APIVersion+"/labels/:id([0-9]+)", &api.LabelAPI{}, "get:Get;put:Put;delete:Delete")
	beego.Router("/api/"+APIVersion+"/labels/:id([0-9]+)/resources", &api.LabelAPI{}, "get:ListResources")

	beego.Router("/api/"+APIVersion+"/systeminfo", &api.SystemInfoAPI{}, "get:GetGeneralInfo")
	beego.Router("/api/"+APIVersion+"/systeminfo/volumes", &api.SystemInfoAPI{}, "get:GetVolumeInfo")
	beego.Router("/api/"+APIVersion+"/systeminfo/getcert", &api.SystemInfoAPI{}, "get:GetCert")

	beego.Router("/api/"+APIVersion+"/registries", &api.RegistryAPI{}, "get:List;post:Post")
	beego.Router("/api/"+APIVersion+"/registries/:id([0-9]+)", &api.RegistryAPI{}, "get:Get;put:Put;delete:Delete")
	beego.Router("/api/"+APIVersion+"/registries/ping", &api.RegistryAPI{}, "post:Ping")
	// we use "0" as the ID of the local Harbor registry, so don't add "([0-9]+)" in the path
	beego.Router("/api/"+APIVersion+"/registries/:id/info", &api.RegistryAPI{}, "get:GetInfo")
	beego.Router("/api/"+APIVersion+"/registries/:id/namespace", &api.RegistryAPI{}, "get:GetNamespace")

	beego.Router("/api/"+APIVersion+"/retentions/metadatas", &api.RetentionAPI{}, "get:GetMetadatas")
	beego.Router("/api/"+APIVersion+"/retentions/:id", &api.RetentionAPI{}, "get:GetRetention")
	beego.Router("/api/"+APIVersion+"/retentions", &api.RetentionAPI{}, "post:CreateRetention")
	beego.Router("/api/"+APIVersion+"/retentions/:id", &api.RetentionAPI{}, "put:UpdateRetention")
	beego.Router("/api/"+APIVersion+"/retentions/:id/executions", &api.RetentionAPI{}, "post:TriggerRetentionExec")
	beego.Router("/api/"+APIVersion+"/retentions/:id/executions/:eid", &api.RetentionAPI{}, "patch:OperateRetentionExec")
	beego.Router("/api/"+APIVersion+"/retentions/:id/executions", &api.RetentionAPI{}, "get:ListRetentionExecs")
	beego.Router("/api/"+APIVersion+"/retentions/:id/executions/:eid/tasks", &api.RetentionAPI{}, "get:ListRetentionExecTasks")
	beego.Router("/api/"+APIVersion+"/retentions/:id/executions/:eid/tasks/:tid", &api.RetentionAPI{}, "get:GetRetentionExecTaskLog")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/immutabletagrules", &api.ImmutableTagRuleAPI{}, "get:List;post:Post")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/immutabletagrules/:id([0-9]+)", &api.ImmutableTagRuleAPI{})

	// APIs for chart repository
	if config.WithChartMuseum() {
		// Charts are controlled under projects
		chartRepositoryAPIType := &api.ChartRepositoryAPI{}
		beego.Router("/api/"+APIVersion+"/chartrepo/health", chartRepositoryAPIType, "get:GetHealthStatus")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts", chartRepositoryAPIType, "get:ListCharts")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts/:name", chartRepositoryAPIType, "get:ListChartVersions")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts/:name", chartRepositoryAPIType, "delete:DeleteChart")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts/:name/:version", chartRepositoryAPIType, "get:GetChartVersion")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts/:name/:version", chartRepositoryAPIType, "delete:DeleteChartVersion")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts", chartRepositoryAPIType, "post:UploadChartVersion")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/prov", chartRepositoryAPIType, "post:UploadChartProvFile")
		beego.Router("/api/"+APIVersion+"/chartrepo/charts", chartRepositoryAPIType, "post:UploadChartVersion")

		// Labels for chart
		chartLabelAPIType := &api.ChartLabelAPI{}
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts/:name/:version/labels", chartLabelAPIType, "get:GetLabels;post:MarkLabel")
		beego.Router("/api/"+APIVersion+"/chartrepo/:repo/charts/:name/:version/labels/:id([0-9]+)", chartLabelAPIType, "delete:RemoveLabel")
	}

	// Add routes for plugin scanner management
	scannerAPI := &api.ScannerAPI{}
	beego.Router("/api/"+APIVersion+"/scanners", scannerAPI, "post:Create;get:List")
	beego.Router("/api/"+APIVersion+"/scanners/:uuid", scannerAPI, "get:Get;delete:Delete;put:Update;patch:SetAsDefault")
	beego.Router("/api/"+APIVersion+"/scanners/:uuid/metadata", scannerAPI, "get:Metadata")
	beego.Router("/api/"+APIVersion+"/scanners/ping", scannerAPI, "post:Ping")

	// Add routes for project level scanner
	proScannerAPI := &api.ProjectScannerAPI{}
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/scanner", proScannerAPI, "get:GetProjectScanner;put:SetProjectScanner")
	beego.Router("/api/"+APIVersion+"/projects/:pid([0-9]+)/scanner/candidates", proScannerAPI, "get:GetProScannerCandidates")

	// Add routes for scan all metrics
	scanAllAPI := &api.ScanAllAPI{}
	beego.Router("/api/"+APIVersion+"/scans/all/metrics", scanAllAPI, "get:GetScanAllMetrics")
	beego.Router("/api/"+APIVersion+"/scans/schedule/metrics", scanAllAPI, "get:GetScheduleMetrics")
}
