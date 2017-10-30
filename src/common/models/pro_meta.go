// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package models

import (
	"time"
)

// keys of project metadata and severity values
const (
	ProMetaPublic             = "public"
	ProMetaEnableContentTrust = "enable_content_trust"
	ProMetaPreventVul         = "prevent_vul" //prevent vulnerable images from being pulled
	ProMetaSeverity           = "severity"
	ProMetaAutoScan           = "auto_scan"
	SeverityNone              = "negligible"
	SeverityLow               = "low"
	SeverityMedium            = "medium"
	SeverityHigh              = "high"
)

// ProjectMetadata holds the metadata of a project.
type ProjectMetadata struct {
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	Name         string    `orm:"column(name)" json:"name"`
	Value        string    `orm:"column(value)" json:"value"`
	CreationTime time.Time `orm:"column(creation_time)" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time)" json:"update_time"`
	Deleted      int       `orm:"column(deleted)" json:"deleted"`
}
