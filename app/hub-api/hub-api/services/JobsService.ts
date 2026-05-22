/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { job_Job } from '../models/job_Job';
import type { job_Jobs } from '../models/job_Jobs';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class JobsService {
    /**
     * List jobs
     * Retrieves a list of all jobs
     * @returns job_Job OK
     * @throws ApiError
     */
    public static getJobs(): CancelablePromise<Array<job_Job>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/jobs',
            errors: {
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Create new jobs
     * Creates new jobs based on the provided details
     * @returns job_Jobs Created
     * @throws ApiError
     */
    public static postJobs({
        job,
    }: {
        /**
         * Job details
         */
        job: job_Jobs,
    }): CancelablePromise<job_Jobs> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/jobs',
            body: job,
        });
    }
    /**
     * Get job details
     * Retrieves details of a specific job by ID
     * @returns job_Job OK
     * @throws ApiError
     */
    public static getJobs1({
        jobId,
    }: {
        /**
         * Job ID
         */
        jobId: string,
    }): CancelablePromise<job_Job> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/jobs/{jobId}',
            path: {
                'jobId': jobId,
            },
        });
    }
}
