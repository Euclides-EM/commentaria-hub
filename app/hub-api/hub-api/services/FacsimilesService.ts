/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Facsimile } from '../models/model_Facsimile';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class FacsimilesService {
    /**
     * List Facsimiles (bulk get)
     * Get facsimiles, optionally filtered by edition ID.
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static getFacsimilies({
        editionId,
    }: {
        /**
         * Filter by edition ID
         */
        editionId?: Array<string>,
    }): CancelablePromise<Array<model_Facsimile>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/facsimilies',
            query: {
                'edition_id': editionId,
            },
        });
    }
    /**
     * Create Facsimile
     * Create a new facsimile
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static postFacsimilies({
        facsimile,
    }: {
        /**
         * Facsimile to create
         */
        facsimile: model_Facsimile,
    }): CancelablePromise<model_Facsimile> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/facsimilies',
            body: facsimile,
        });
    }
    /**
     * Get Facsimile by ID
     * Get a single facsimile by its ID.
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static getFacsimilies1({
        id,
    }: {
        /**
         * Facsimile ID
         */
        id: string,
    }): CancelablePromise<model_Facsimile> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/facsimilies/{id}',
            path: {
                'id': id,
            },
            errors: {
                404: `Facsimile not found`,
            },
        });
    }
    /**
     * Update Facsimile
     * Update an existing facsimile identified by ID.
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static putFacsimilies({
        id,
        facsimile,
    }: {
        /**
         * Facsimile ID
         */
        id: string,
        /**
         * Facsimile data to update
         */
        facsimile: model_Facsimile,
    }): CancelablePromise<model_Facsimile> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/facsimilies/{id}',
            path: {
                'id': id,
            },
            body: facsimile,
        });
    }
}
