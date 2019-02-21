import { Injectable } from '@angular/core';
import { Http } from '@angular/http';
import { Observable, Subscription, Subject, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { GcApiRepository } from './gc.api.repository';
import { ErrorHandler } from '../../error-handler/index';
import { GcJobData } from './gcLog';


@Injectable()
export class GcRepoService {

    constructor(private http: Http,
        private gcApiRepository: GcApiRepository,
        private errorHandler: ErrorHandler) {
    }

    public manualGc(): Observable <any> {
        let param = {
            "schedule": {
                "type": "Manual"
            }
        };
        return this.gcApiRepository.postSchedule(param);
    }

    public getJobs(): Observable <GcJobData []> {
        return this.gcApiRepository.getJobs();
    }

    public getLog(id): Observable <any> {
        return this.gcApiRepository.getLog(id);
    }

    public getSchedule(): Observable <any> {
        return this.gcApiRepository.getSchedule();
    }

    public postScheduleGc(type, offTime, weekday ?): Observable <any> {
        let param = {
            "schedule": {
                "type": type,
                "offtime": offTime,
            }
        };
        if (weekday) {
            param.schedule["weekday"] = weekday;
        }
        return this.gcApiRepository.postSchedule(param);
    }

    public putScheduleGc(type, offTime, weekday ?): Observable <any> {
        let param = {
            "schedule": {
                "type": type,
                "offtime": offTime,
            }
        };
        if (weekday) {
            param.schedule["weekday"] = weekday;
        }
        return this.gcApiRepository.putSchedule(param);
    }
}
