import { ComponentFixture, TestBed, async, fakeAsync, tick } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";
import { DebugElement } from '@angular/core';

import { SharedModule } from '../shared/shared.module';
import { EndpointComponent } from './endpoint.component';
import { FilterComponent } from '../filter/filter.component';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { CreateEditEndpointComponent } from '../create-edit-endpoint/create-edit-endpoint.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import { ErrorHandler } from '../error-handler/error-handler';
import { Endpoint } from '../service/interface';
import { EndpointService, EndpointDefaultService } from '../service/endpoint.service';
import { IServiceConfig, SERVICE_CONFIG } from '../service.config';

import { click } from '../utils';

describe('EndpointComponent (inline template)', () => {

  let mockData: Endpoint[] = [
    {
        "id": 1,
        "endpoint": "https://10.117.4.151",
        "name": "target_01",
        "username": "admin",
        "password": "",
        "insecure": true,
        "type": 0
    },
    {
        "id": 2,
        "endpoint": "https://10.117.5.142",
        "name": "target_02",
        "username": "AAA",
        "password": "",
        "insecure": false,
        "type": 0
    },
    {
        "id": 3,
        "endpoint": "https://101.1.11.111",
        "name": "target_03",
        "username": "admin",
        "password": "",
        "insecure": false,
        "type": 0
    },
    {
        "id": 4,
        "endpoint": "http://4.4.4.4",
        "name": "target_04",
        "username": "",
        "password": "",
        "insecure": false,
        "type": 0
    }
  ];

  let mockOne: Endpoint = {
    "id": 1,
    "endpoint": "https://10.117.4.151",
    "name": "target_01",
    "username": "admin",
    "password": "",
      "insecure": false,
    "type": 0
  };

  let comp: EndpointComponent;
  let fixture: ComponentFixture<EndpointComponent>;
  let config: IServiceConfig = {
    systemInfoEndpoint: '/api/endpoints/testing'
  };

  let endpointService: EndpointService;
  let spy: jasmine.Spy;
  let spyOnRules: jasmine.Spy;
  let spyOne: jasmine.Spy;
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [ 
         SharedModule,
         NoopAnimationsModule
      ],
      declarations: [ 
          FilterComponent, 
          ConfirmationDialogComponent, 
          CreateEditEndpointComponent, 
          InlineAlertComponent,
          EndpointComponent ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: EndpointService, useClass: EndpointDefaultService }
      ]
    });
  }));

  beforeEach(()=>{
    fixture = TestBed.createComponent(EndpointComponent);
    comp = fixture.componentInstance;
    
    endpointService = fixture.debugElement.injector.get(EndpointService);

    spy = spyOn(endpointService, 'getEndpoints').and.returnValues(Promise.resolve(mockData));
    spyOnRules = spyOn(endpointService, 'getEndpointWithReplicationRules').and.returnValue([]);
    spyOne = spyOn(endpointService, 'getEndpoint').and.returnValue(Promise.resolve(mockOne));
    fixture.detectChanges();
  });  

  it('should retrieve endpoint data', () => {
    fixture.detectChanges();
    expect(spy.calls.any()).toBeTruthy();
  });

  it('should endpoint be initialized', () => {
    fixture.detectChanges();
    expect(config.systemInfoEndpoint).toEqual('/api/endpoints/testing');
  });

  it('should open create endpoint modal', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(()=>{      
      fixture.detectChanges();
      comp.editTarget(mockOne);
      fixture.detectChanges();
      expect(comp.target.name).toEqual('target_01');
    });
  }));

  it('should filter endpoints by keyword', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      comp.doSearchTargets('target_02');
      fixture.detectChanges();
      expect(comp.targets.length).toEqual(1);
    });
  }));

   it('should render data', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css('datagrid-cell'));
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el.textContent).toEqual('target_01');
    });
  }));

  it('should open creation endpoint', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      let de: DebugElement = fixture.debugElement.query(By.css('btn-link'));
      expect(de).toBeTruthy();
      fixture.detectChanges();
      click(de);
      fixture.detectChanges();
      let deInput: DebugElement = fixture.debugElement.query(By.css('input'));
      expect(deInput).toBeTruthy();
    });
  }));


  it('should open to edit existing endpoint', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      let de: DebugElement = fixture.debugElement.query(del=>del.classes['action-item']);
      expect(de).toBeTruthy();
      fixture.detectChanges();
      click(de);
      fixture.detectChanges();
      let deInput: DebugElement = fixture.debugElement.query(By.css('input'));
      expect(deInput).toBeTruthy();
      let elInput: HTMLElement = deInput.nativeElement;
      expect(elInput).toBeTruthy();
      expect(elInput.textContent).toEqual('target_01');
    });
  }));

});