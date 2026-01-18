/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Edition } from '../models/model_Edition';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class EditionsService {
    /**
     * List Editions
     * Get a list of available editions. Optionally include facsimiles.
     * @returns model_Edition OK
     * @throws ApiError
     */
    public static getEditions({
        expand,
        orderBy,
    }: {
        /**
         * Include related entities
         */
        expand?: 'facsimiles',
        /**
         * Order by field
         */
        orderBy?: 'suggested',
    }): CancelablePromise<Array<model_Edition>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/editions',
            query: {
                'expand': expand,
                'orderBy': orderBy,
            },
        });
    }
    /**
     * Create Edition
     * Create a new edition
     * @returns model_Edition OK
     * @throws ApiError
     */
    public static postEditions({
        edition,
    }: {
        /**
         * Edition to create
         */
        edition: model_Edition,
    }): CancelablePromise<model_Edition> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/editions',
            body: edition,
        });
    }
}
