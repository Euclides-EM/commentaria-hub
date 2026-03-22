/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Result } from '../models/feature_Result';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class FeatureResultsService {
    /**
     * List feature results
     * Get a list of feature results
     * @returns feature_Result OK
     * @throws ApiError
     */
    public static getDatasetsAnnotationsResults({
        dataSetId,
        id,
        keys,
        features,
        fallbackToOrigin,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Comma-separated list of keys to filter results
         */
        keys?: string,
        /**
         * Comma-separated list of feature names to filter results
         */
        features?: string,
        /**
         * Whether to fallback to results of the origin annotation if no feature results are found.
         */
        fallbackToOrigin?: boolean,
    }): CancelablePromise<Array<feature_Result>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/results',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'keys': keys,
                'features': features,
                'fallback_to_origin': fallbackToOrigin,
            },
        });
    }
    /**
     * Create feature results
     * Create new feature results (batch)
     * @returns feature_Result OK
     * @throws ApiError
     */
    public static postDatasetsAnnotationsResults({
        dataSetId,
        id,
        result,
        pushToOrigin,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Feature results data
         */
        result: Array<feature_Result>,
        /**
         * Whether to push the created results to the origin annotation.
         */
        pushToOrigin?: boolean,
    }): CancelablePromise<Array<feature_Result>> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations/{id}/results',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'push_to_origin': pushToOrigin,
            },
            body: result,
        });
    }
    /**
     * Get feature results SQL dump
     * Get a SQL dump of feature results for a specific annotation
     * @returns string SQL dump of feature results
     * @throws ApiError
     */
    public static getDatasetsAnnotationsResultsSqldump({
        dataSetId,
        id,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
    }): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/results/sqldump',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
        });
    }
}
