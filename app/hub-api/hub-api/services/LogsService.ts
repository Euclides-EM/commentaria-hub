/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { common_LogTail } from '../models/common_LogTail';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class LogsService {
    /**
     * Tail service logs
     * Returns the last n lines from the deployed API service logs.
     * @returns common_LogTail OK
     * @throws ApiError
     */
    public static getLogs({
        n,
    }: {
        /**
         * Number of log lines to return
         */
        n?: number,
    }): CancelablePromise<common_LogTail> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/logs',
            query: {
                'n': n,
            },
        });
    }
}
