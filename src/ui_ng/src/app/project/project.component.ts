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
import { Component, OnInit, ViewChild, OnDestroy} from '@angular/core';
import { Router } from '@angular/router';
import { Project } from './project';
import { CreateProjectComponent } from './create-project/create-project.component';
import { ListProjectComponent } from './list-project/list-project.component';
import { AppConfigService } from '../app-config.service';
import { SessionService } from '../shared/session.service';
import { ProjectTypes } from '../shared/shared.const';

@Component({
  selector: 'project',
  templateUrl: 'project.component.html',
  styleUrls: ['./project.component.css']
})
export class ProjectComponent implements OnInit {
  projectTypes = ProjectTypes;

  @ViewChild(CreateProjectComponent)
  creationProject: CreateProjectComponent;

  @ViewChild(ListProjectComponent)
  listProject: ListProjectComponent;

  currentFilteredType: number = 0;//all projects
  projectName: string = "";

  loading: boolean = true;

  get selecteType(): number {
    return this.currentFilteredType;
  }
  set selecteType(_project: number) {
    this.currentFilteredType = _project;
    if (window.sessionStorage) {
      window.sessionStorage['projectTypeValue'] = _project;
    }
  }

  constructor(
    private appConfigService: AppConfigService,
    private sessionService: SessionService) {
  }

  ngOnInit(): void {
    if (window.sessionStorage && window.sessionStorage['projectTypeValue'] && window.sessionStorage['fromDetails']) {
      this.currentFilteredType = +window.sessionStorage['projectTypeValue'];
      window.sessionStorage.removeItem('fromDetails');
    }
  }

  get projectCreationRestriction(): boolean {
    let account = this.sessionService.getCurrentUser();
    if (account) {
      switch (this.appConfigService.getConfig().project_creation_restriction) {
        case 'adminonly':
          return (account.has_admin_role === 1);
        case 'everyone':
          return true;
      }
    }
    return false;
  }

  openModal(): void {
    this.creationProject.newProject();
  }

  createProject(created: boolean) {
    if (created) {
      this.refresh();
    }
  }

  doSearchProjects(projectName: string): void {
    this.projectName = projectName;
    this.listProject.doSearchProject(this.projectName);
  }

  doFilterProjects(): void {
    this.listProject.doFilterProject(this.selecteType);
  }

  refresh(): void {
    this.currentFilteredType = 0;
    this.projectName = "";
    this.listProject.refresh();
  }

}