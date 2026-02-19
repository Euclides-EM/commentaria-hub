/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { common_HealthStatus } from '../models/common_HealthStatus';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class HealthService {
    /**
     * Health check
     * Returns service and DB status
     * @returns common_HealthStatus OK
     * @throws ApiError
     */
    public static getHealth(): CancelablePromise<common_HealthStatus> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/health',
        });
    }
}
