/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Feature } from '../models/feature_Feature';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class FeaturesService {
    /**
     * List Edition Features
     * Get a list of available features for the global editions scope
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static getEditionsFeatures({
        expand,
    }: {
        /**
         * Include related entities
         */
        expand?: Array<string>,
    } = {}): CancelablePromise<Array<feature_Feature>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/editions/features',
            query: {
                'expand': expand,
            },
        });
    }

    /**
     * List Features
     * Get a list of available features for the dataset
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static getDatasetsFeatures({
        dataSetId,
        expand,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Include related entities
         */
        expand?: Array<string>,
    }): CancelablePromise<Array<feature_Feature>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/features',
            path: {
                'dataSetId': dataSetId,
            },
            query: {
                'expand': expand,
            },
        });
    }
    /**
     * Create Feature
     * Create a new feature for the dataset
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static postDatasetsFeatures({
        dataSetId,
        feature,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Feature to create
         */
        feature: feature_Feature,
    }): CancelablePromise<feature_Feature> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/features',
            path: {
                'dataSetId': dataSetId,
            },
            body: feature,
        });
    }
    /**
     * Get Feature
     * Get details of a specific feature from the dataset
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static getDatasetsFeatures1({
        dataSetId,
        featureId,
        expand,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Feature ID
         */
        featureId: string,
        /**
         * Include related entities
         */
        expand?: Array<string>,
    }): CancelablePromise<feature_Feature> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/features/{featureId}',
            path: {
                'dataSetId': dataSetId,
                'featureId': featureId,
            },
            query: {
                'expand': expand,
            },
        });
    }
    /**
     * Update Feature
     * Update an existing feature in the dataset.
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static putDatasetsFeatures({
        dataSetId,
        featureId,
        feature,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Feature ID
         */
        featureId: string,
        /**
         * Updated feature data
         */
        feature: feature_Feature,
    }): CancelablePromise<feature_Feature> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/features/{featureId}',
            path: {
                'dataSetId': dataSetId,
                'featureId': featureId,
            },
            body: feature,
        });
    }
    /**
     * Delete Feature
     * Delete a feature from the dataset.
     * @returns void
     * @throws ApiError
     */
    public static deleteDatasetsFeatures({
        dataSetId,
        featureId,
        force,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Feature ID
         */
        featureId: string,
        /**
         * Force deletion
         */
        force?: boolean,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/datasets/{dataSetId}/features/{featureId}',
            path: {
                'dataSetId': dataSetId,
                'featureId': featureId,
            },
            query: {
                'force': force,
            },
        });
    }
}
