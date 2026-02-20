/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Execution } from '../models/feature_Execution';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ExecutionsService {
    /**
     * List Executions
     * Get a list of all executions for a specific edition
     * @returns feature_Execution OK
     * @throws ApiError
     */
    public static getFeaturesExecutions({
        dataset,
        features,
        statuses,
    }: {
        /**
         * Filter by dataset ID
         */
        dataset?: string,
        /**
         * Filter by delimited list of feature IDs
         */
        features?: string,
        /**
         * Filter by delimited list of execution statuses
         */
        statuses?: 'pending' | 'running' | 'completed' | 'failed',
    }): CancelablePromise<Array<feature_Execution>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/features/executions',
            query: {
                'dataset': dataset,
                'features': features,
                'statuses': statuses,
            },
        });
    }
    /**
     * Create Execution
     * Create a new execution for a specific feature
     * @returns feature_Execution OK
     * @throws ApiError
     */
    public static postFeaturesExecutions({
        execution,
    }: {
        /**
         * Execution to create
         */
        execution: feature_Execution,
    }): CancelablePromise<feature_Execution> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/features/executions',
            body: execution,
        });
    }
    /**
     * Get Execution
     * Get details of a specific execution by ID
     * @returns feature_Execution OK
     * @throws ApiError
     */
    public static getFeaturesExecutions1({
        executionId,
    }: {
        /**
         * Execution ID
         */
        executionId: string,
    }): CancelablePromise<feature_Execution> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/features/executions/{executionId}',
            path: {
                'executionId': executionId,
            },
        });
    }
    /**
     * Cancel Execution
     * Cancel a running execution by ID
     * @returns string status: cancelled
     * @throws ApiError
     */
    public static putFeaturesExecutionsCancel({
        executionId,
    }: {
        /**
         * Execution ID
         */
        executionId: string,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/features/executions/{executionId}/cancel',
            path: {
                'executionId': executionId,
            },
        });
    }
}
