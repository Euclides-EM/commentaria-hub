/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Revision } from '../models/feature_Revision';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class EditionFeatureRevisionsService {
    /**
     * List Edition Feature Revisions
     * Get a list of revisions for a specific edition feature
     * @returns feature_Revision OK
     * @throws ApiError
     */
    public static getFeaturesRevisions({
        featureId,
    }: {
        /**
         * Feature ID
         */
        featureId: string,
    }): CancelablePromise<Array<feature_Revision>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/features/{featureId}/revisions',
            path: {
                'featureId': featureId,
            },
        });
    }
    /**
     * Create Edition Feature Revision
     * Create a new revision for a specific edition feature
     * @returns feature_Revision OK
     * @throws ApiError
     */
    public static postFeaturesRevisions({
        featureId,
        revision,
    }: {
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
            url: '/features/{featureId}/revisions',
            path: {
                'featureId': featureId,
            },
            body: revision,
        });
    }
    /**
     * Get Edition Feature Revision
     * Get details of a specific edition feature revision
     * @returns feature_Revision OK
     * @throws ApiError
     */
    public static getFeaturesRevisions1({
        featureId,
        revisionId,
    }: {
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
            url: '/features/{featureId}/revisions/{revisionId}',
            path: {
                'featureId': featureId,
                'revisionId': revisionId,
            },
        });
    }
}
