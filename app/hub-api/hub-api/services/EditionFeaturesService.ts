/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Feature } from '../models/feature_Feature';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class EditionFeaturesService {
    /**
     * List Edition Features
     * Get a list of available features for the global editions scope
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static getFeatures({
        scope,
        expand,
        dataset,
    }: {
        /**
         * Filter by feature execution scope
         */
        scope: 'dataset' | 'editions',
        /**
         * Include related entities
         */
        expand?: Array<string>,
        /**
         * Filter by dataset ID, relevant only for the dataset scope
         */
        dataset?: string,
    }): CancelablePromise<Array<feature_Feature>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/features',
            query: {
                'expand': expand,
                'scope': scope,
                'dataset': dataset,
            },
        });
    }
    /**
     * Create Edition Feature
     * Create a new feature for the global editions scope
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static postFeatures({
        feature,
    }: {
        /**
         * Feature to create
         */
        feature: feature_Feature,
    }): CancelablePromise<feature_Feature> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/features',
            body: feature,
        });
    }
    /**
     * Get Edition Feature
     * Get details of a specific feature from the global editions scope
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static getFeatures1({
        featureId,
        expand,
    }: {
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
            url: '/features/{featureId}',
            path: {
                'featureId': featureId,
            },
            query: {
                'expand': expand,
            },
        });
    }
    /**
     * Update Edition Feature
     * Update an existing feature in the global editions scope.
     * @returns feature_Feature OK
     * @throws ApiError
     */
    public static putFeatures({
        featureId,
        feature,
    }: {
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
            url: '/features/{featureId}',
            path: {
                'featureId': featureId,
            },
            body: feature,
        });
    }
    /**
     * Delete Edition Feature
     * Delete a feature from the global editions scope.
     * @returns void
     * @throws ApiError
     */
    public static deleteFeatures({
        featureId,
        force,
    }: {
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
            url: '/features/{featureId}',
            path: {
                'featureId': featureId,
            },
            query: {
                'force': force,
            },
        });
    }
}
