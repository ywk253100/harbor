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
import { Component, OnInit, ViewChild, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';
import { Endpoint, ReplicationRule } from '../service/interface';
import { EndpointService } from '../service/endpoint.service';

import { TranslateService } from '@ngx-translate/core';

import { ErrorHandler } from '../error-handler/index';

import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';

import { ConfirmationTargets, ConfirmationState, ConfirmationButtons } from '../shared/shared.const';

import { Subscription } from 'rxjs/Subscription';

import { CreateEditEndpointComponent } from '../create-edit-endpoint/create-edit-endpoint.component';

import { ENDPOINT_STYLE } from './endpoint.component.css';
import { ENDPOINT_TEMPLATE } from './endpoint.component.html';

import { toPromise, CustomComparator } from '../utils';

import { State, Comparator } from 'clarity-angular';

@Component({
    selector: 'hbr-endpoint',
    template: ENDPOINT_TEMPLATE,
    styles: [ENDPOINT_STYLE],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class EndpointComponent implements OnInit {

    @ViewChild(CreateEditEndpointComponent)
    createEditEndpointComponent: CreateEditEndpointComponent;


    @ViewChild('confirmationDialog')
    confirmationDialogComponent: ConfirmationDialogComponent;

    targets: Endpoint[];
    target: Endpoint;

    targetName: string;
    subscription: Subscription;

    loading: boolean = false;

    creationTimeComparator: Comparator<Endpoint> = new CustomComparator<Endpoint>('creation_time', 'date');

    timerHandler: any;

    get initEndpoint(): Endpoint {
        return {
            endpoint: "",
            name: "",
            username: "",
            password: "",
            insecure: false,
            type: 0
        };
    }

    constructor(
        private endpointService: EndpointService,
        private errorHandler: ErrorHandler,
        private translateService: TranslateService,
        private ref: ChangeDetectorRef) {
        this.forceRefreshView(1000);
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (message &&
            message.source === ConfirmationTargets.TARGET &&
            message.state === ConfirmationState.CONFIRMED) {

            let targetId = message.data;
            toPromise<number>(this.endpointService
                .deleteEndpoint(targetId))
                .then(
                response => {
                    this.translateService.get('DESTINATION.DELETED_SUCCESS')
                        .subscribe(res => this.errorHandler.info(res));
                    this.reload(true);
                }).catch(
                error => {
                    if (error && error.status === 412) {
                        this.translateService.get('DESTINATION.FAILED_TO_DELETE_TARGET_IN_USED')
                            .subscribe(res => this.errorHandler.error(res));
                    } else {
                        this.errorHandler.error(error);
                    }
                });
        }
    }

    ngOnInit(): void {
        this.targetName = '';
        this.retrieve();
    }

    ngOnDestroy(): void {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    }

    retrieve(): void {
        this.loading = true;
        toPromise<Endpoint[]>(this.endpointService
            .getEndpoints(this.targetName))
            .then(
            targets => {
                this.targets = targets || [];
                this.forceRefreshView(1000);
                this.loading = false;
            }).catch(error => {
                this.errorHandler.error(error);
                this.loading = false;
            });
    }

    doSearchTargets(targetName: string) {
        this.targetName = targetName;
        this.retrieve();
    }

    refreshTargets() {
        this.retrieve();
    }

    reload($event: any) {
        this.targetName = '';
        this.retrieve();
    }

    openModal() {
        this.createEditEndpointComponent.openCreateEditTarget(true);
        this.target = this.initEndpoint;
    }

    editTarget(target: Endpoint) {
        if (target) {
            let editable = true;
            if (!target.id) {
                return;
            }
            let id: number | string = target.id;
            toPromise<ReplicationRule[]>(this.endpointService
                .getEndpointWithReplicationRules(id))
                .then(
                rules => {
                    if (rules && rules.length > 0) {
                        rules.forEach((rule) => editable = (rule && rule.enabled !== 1));
                    }
                    this.createEditEndpointComponent.openCreateEditTarget(editable, id);
                    this.forceRefreshView(1000);
                })
                .catch(error => this.errorHandler.error(error));
        }
    }

    deleteTarget(target: Endpoint) {
        if (target) {
            let targetId = target.id;
            let deletionMessage = new ConfirmationMessage(
                'REPLICATION.DELETION_TITLE_TARGET',
                'REPLICATION.DELETION_SUMMARY_TARGET',
                target.name || '',
                target.id,
                ConfirmationTargets.TARGET,
                ConfirmationButtons.DELETE_CANCEL);
            this.confirmationDialogComponent.open(deletionMessage);
        }
    }

    //Forcely refresh the view
    forceRefreshView(duration: number): void {
        //Reset timer
        if (this.timerHandler) {
            clearInterval(this.timerHandler);
        }
        this.timerHandler = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => {
            if (this.timerHandler) {
                clearInterval(this.timerHandler);
                this.timerHandler = null;
            }
        }, duration);
    }
}