/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_USTC } from '../models/model_USTC';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ThirdPartyCatalogsService {
    /**
     * Lookup USTC metadata by ID
     * Fetches metadata from the USTC catalog based on the provided USTC ID.
     * @returns model_USTC OK
     * @throws ApiError
     */
    public static postCatalogsUstcLookup({
        ustc,
        ustcId,
    }: {
        /**
         * JSON with ustc_id
         */
        ustc?: model_USTC,
        /**
         * USTC ID (alternative to body)
         */
        ustcId?: number,
    }): CancelablePromise<model_USTC> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/catalogs/ustc/lookup',
            query: {
                'ustc_id': ustcId,
            },
            body: ustc,
        });
    }
}
