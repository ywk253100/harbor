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
import { NgForm } from '@angular/forms';
import { httpStatusCode, AlertType } from './shared.const';
import { MessageService } from '../global-message/message.service';
import { Comparator, State } from 'clarity-angular';

/**
 * To handle the error message body
 * 
 * @export
 * @returns {string}
 */
export const errorHandler = function (error: any): string {
    if (!error) {
        return "UNKNOWN_ERROR";
    }
    if (!(error.statusCode || error.status)) {
        //treat as string message
        return '' + error;
    } else {
        switch (error.statusCode || error.status) {
            case 400:
                return "BAD_REQUEST_ERROR";
            case 401:
                return "UNAUTHORIZED_ERROR";
            case 403:
                return "FORBIDDEN_ERROR";
            case 404:
                return "NOT_FOUND_ERROR";
            case 412:
                return "PRECONDITION_FAILED";
            case 409:
                return "CONFLICT_ERROR";
            case 500:
                return "SERVER_ERROR";
            default:
                return "UNKNOWN_ERROR";
        }
    }
}

/**
 * To check if form is empty
 */
export const isEmptyForm = function (ngForm: NgForm): boolean {
    if (ngForm && ngForm.form) {
        let values = ngForm.form.value;
        if (values) {
            for (var key in values) {
                if (values[key]) {
                    return false;
                }
            }
        }

    }

    return true;
}

/**
 * Hanlde the 401 and 403 code
 * 
 * If handled the 401 or 403, then return true otherwise false
 */
export const accessErrorHandler = function (error: any, msgService: MessageService): boolean {
    if (error && error.status && msgService) {
        if (error.status === httpStatusCode.Unauthorized) {
            msgService.announceAppLevelMessage(error.status, "UNAUTHORIZED_ERROR", AlertType.DANGER);
            return true;
        }
    }

    return false;
}

//Provide capability of reconstructing the query paramter
export const maintainUrlQueryParmas = function (uri: string, key: string, value: string): string {
    let re: RegExp = new RegExp("([?&])" + key + "=.*?(&|#|$)", "i");
    if (value === undefined) {
        if (uri.match(re)) {
            return uri.replace(re, '$1$2');
        } else {
            return uri;
        }
    } else {
        if (uri.match(re)) {
            return uri.replace(re, '$1' + key + "=" + value + '$2');
        } else {
            var hash = '';
            if (uri.indexOf('#') !== -1) {
                hash = uri.replace(/.*#/, '#');
                uri = uri.replace(/#.*/, '');
            }
            var separator = uri.indexOf('?') !== -1 ? "&" : "?";
            return uri + separator + key + "=" + value + hash;
        }
    }
}

//Copy from ui library utils.ts

/**
 * Calculate page number by state
 */
export function calculatePage(state: State): number {
    if (!state || !state.page) {
        return 1;
    }

    return Math.ceil((state.page.to + 1) / state.page.size);
}

/**
 * Comparator for fields with specific type.
 *  
 */
export class CustomComparator<T> implements Comparator<T> {

    fieldName: string;
    type: string;

    constructor(fieldName: string, type: string) {
        this.fieldName = fieldName;
        this.type = type;
    }

    compare(a: { [key: string]: any | any[] }, b: { [key: string]: any | any[] }) {
        let comp = 0;
        if (a && b) {
            let fieldA = a[this.fieldName];
            let fieldB = b[this.fieldName];
            switch (this.type) {
                case "number":
                    comp = fieldB - fieldA;
                    break;
                case "date":
                    comp = new Date(fieldB).getTime() - new Date(fieldA).getTime();
                    break;
            }
        }
        return comp;
    }
}

/**
 * Filter columns via RegExp
 * 
 * @export
 * @param {State} state 
 * @returns {void} 
 */
export function doFiltering<T extends { [key: string]: any | any[] }>(items: T[], state: State): T[] {
    if (!items || items.length === 0) {
        return items;
    }

    if (!state || !state.filters || state.filters.length === 0) {
        return items;
    }

    state.filters.forEach((filter: {
        property: string;
        value: string;
    }) => {
        items = items.filter(item => regexpFilter(filter["value"], item[filter["property"]]));
    });

    return items;
}

/**
 * Match items via RegExp
 * 
 * @export
 * @param {string} terms 
 * @param {*} testedValue 
 * @returns {boolean} 
 */
export function regexpFilter(terms: string, testedValue: any): boolean {
    let reg = new RegExp('.*' + terms + '.*', 'i');
    return reg.test(testedValue);
}

/**
 * Sorting the data by column
 * 
 * @export
 * @template T 
 * @param {T[]} items 
 * @param {State} state 
 * @returns {T[]} 
 */
export function doSorting<T extends { [key: string]: any | any[] }>(items: T[], state: State): T[] {
    if (!items || items.length === 0) {
        return items;
    }
    if (!state || !state.sort) {
        return items;
    }

    return items.sort((a: T, b: T) => {
        let comp: number = 0;
        if (typeof state.sort.by !== "string") {
            comp = state.sort.by.compare(a, b);
        } else {
            let propA = a[state.sort.by.toString()], propB = b[state.sort.by.toString()];
            if (typeof propA === "string") {
                comp = propA.localeCompare(propB);
            } else {
                if (propA > propB) {
                    comp = 1;
                } else if (propA < propB) {
                    comp = -1;
                }
            }
        }

        if (state.sort.reverse) {
            comp = -comp;
        }

        return comp;
    });
}