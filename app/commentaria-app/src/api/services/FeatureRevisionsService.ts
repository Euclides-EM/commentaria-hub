/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Revision } from '../models/feature_Revision';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class FeatureRevisionsService {
    /**
     * List Feature Revisions
     * Get a list of revisions for a specific feature
     * @returns feature_Revision OK
     * @throws ApiError
     */
    public static getDatasetsFeaturesRevisions({
        dataSetId,
        featureId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Feature ID
         */
        featureId: string,
    }): CancelablePromise<Array<feature_Revision>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/features/{featureId}/revisions',
            path: {
                'dataSetId': dataSetId,
                'featureId': featureId,
            },
        });
    }
    /**
     * Create Feature Revision
     * Create a new revision for a specific feature
     * @returns feature_Revision OK
     * @throws ApiError
     */
    public static postDatasetsFeaturesRevisions({
        dataSetId,
        featureId,
        revision,
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
         * Revision data
         */
        revision: feature_Revision,
    }): CancelablePromise<feature_Revision> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/features/{featureId}/revisions',
            path: {
                'dataSetId': dataSetId,
                'featureId': featureId,
            },
            body: revision,
        });
    }
    /**
     * Get Feature Revision
     * Get details of a specific feature revision
     * @returns feature_Revision OK
     * @throws ApiError
     */
    public static getDatasetsFeaturesRevisions1({
        dataSetId,
        featureId,
        revisionId,
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
         * Revision ID
         */
        revisionId: string,
    }): CancelablePromise<feature_Revision> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/features/{featureId}/revisions/{revisionId}',
            path: {
                'dataSetId': dataSetId,
                'featureId': featureId,
                'revisionId': revisionId,
            },
        });
    }
}
