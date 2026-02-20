/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { integration_Job } from '../models/integration_Job';
import type { integration_Jobs } from '../models/integration_Jobs';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class IntegrationService {
    /**
     * List integration jobs
     * Retrieves a list of all integration jobs
     * @returns integration_Job OK
     * @throws ApiError
     */
    public static getIntegrationsJobs(): CancelablePromise<Array<integration_Job>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/integrations/jobs',
            errors: {
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Create new integration jobs
     * Creates new integration jobs based on the provided details
     * @returns integration_Jobs Created
     * @throws ApiError
     */
    public static postIntegrationsJobs({
        job,
    }: {
        /**
         * Integration Job Details
         */
        job: integration_Jobs,
    }): CancelablePromise<integration_Jobs> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/integrations/jobs',
            body: job,
        });
    }
    /**
     * Get integration job details
     * Retrieves details of a specific integration job by ID
     * @returns integration_Job OK
     * @throws ApiError
     */
    public static getIntegrationsJobs1({
        jobId,
    }: {
        /**
         * Integration Job ID
         */
        jobId: string,
    }): CancelablePromise<integration_Job> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/integrations/jobs/{jobId}',
            path: {
                'jobId': jobId,
            },
        });
    }
}
