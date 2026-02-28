/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class FeaturePropertiesService {
    /**
     * List available feature properties
     * Get a list of all feature property keys
     * @returns string OK
     * @throws ApiError
     */
    public static getFeaturesProperties(): CancelablePromise<Array<string>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/features/properties',
        });
    }
}
