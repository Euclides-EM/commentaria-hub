/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { job_Platform } from '../models/job_Platform';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class IntegrationsService {
    /**
     * List Integration Platforms
     * Get a list of supported integration platforms.
     * @returns job_Platform OK
     * @throws ApiError
     */
    public static getIntegrationsPlatforms(): CancelablePromise<Array<job_Platform>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/integrations/platforms',
        });
    }
}
