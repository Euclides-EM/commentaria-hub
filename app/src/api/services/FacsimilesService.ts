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
     * Create Facsimile
     * Create a new facsimile
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static postEditionsFacsimilies({
        editionId,
        facsimile,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Facsimile to create
         */
        facsimile: model_Facsimile,
    }): CancelablePromise<model_Facsimile> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/editions/{editionId}/facsimilies',
            path: {
                'editionId': editionId,
            },
            body: facsimile,
        });
    }
}
