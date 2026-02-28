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
    }): CancelablePromise<Array<feature_Result>> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations/{id}/results',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            body: result,
        });
    }
}
