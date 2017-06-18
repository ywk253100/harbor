export const LIST_REPLICATION_RULE_TEMPLATE: string = `
<div>
<clr-datagrid [clrDgLoading]="loading">
    <clr-dg-column [clrDgField]="'name'">{{'REPLICATION.NAME' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'project_name'" *ngIf="projectScope">{{'REPLICATION.PROJECT' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'description'">{{'REPLICATION.DESCRIPTION' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgField]="'target_name'">{{'REPLICATION.DESTINATION_NAME' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgSortBy]="startTimeComparator">{{'REPLICATION.LAST_START_TIME' | translate}}</clr-dg-column>
    <clr-dg-column [clrDgSortBy]="enabledComparator">{{'REPLICATION.ACTIVATION' | translate}}</clr-dg-column>
    <clr-dg-row *clrDgItems="let p of changedRules" [clrDgItem]="p" (click)="selectRule(p)" [style.backgroundColor]="(!projectScope && withReplicationJob && selectedId === p.id) ? '#eee' : ''">
        <clr-dg-action-overflow>
            <button class="action-item" (click)="editRule(p)">{{'REPLICATION.EDIT_POLICY' | translate}}</button>
            <button class="action-item" (click)="toggleRule(p)">{{ (p.enabled === 0 ? 'REPLICATION.ENABLE' : 'REPLICATION.DISABLE') | translate}}</button>
            <button class="action-item" (click)="deleteRule(p)">{{'REPLICATION.DELETE_POLICY' | translate}}</button>
        </clr-dg-action-overflow>
        <clr-dg-cell>
          <ng-template [ngIf]="projectScope">
            <a href="javascript:void(0)" (click)="redirectTo(p)">{{p.name}}</a>
          </ng-template>
          <ng-template [ngIf]="!projectScope">
            {{p.name}}
          </ng-template>
        </clr-dg-cell>
        <clr-dg-cell *ngIf="projectScope">{{p.project_name}}</clr-dg-cell>
        <clr-dg-cell>{{p.description ? p.description : '-'}}</clr-dg-cell>
        <clr-dg-cell>{{p.target_name}}</clr-dg-cell>
        <clr-dg-cell>
          <ng-template [ngIf]="p.start_time === nullTime">-</ng-template>
          <ng-template [ngIf]="p.start_time !== nullTime">{{p.start_time | date: 'short'}}</ng-template>
        </clr-dg-cell>
        <clr-dg-cell>
            {{ (p.enabled === 1 ? 'REPLICATION.ENABLED' : 'REPLICATION.DISABLED') | translate}}
        </clr-dg-cell>
    </clr-dg-row>
    <clr-dg-footer>
      {{pagination.firstItem + 1}} - {{pagination.lastItem +1 }} {{'REPLICATION.OF' | translate}} {{pagination.totalItems }} {{'REPLICATION.ITEMS' | translate}}
      <clr-dg-pagination #pagination [clrDgPageSize]="5"></clr-dg-pagination>
    </clr-dg-footer>
</clr-datagrid>
<confirmation-dialog #toggleConfirmDialog (confirmAction)="toggleConfirm($event)"></confirmation-dialog>
<confirmation-dialog #deletionConfirmDialog (confirmAction)="deletionConfirm($event)"></confirmation-dialog>
</div>
`;