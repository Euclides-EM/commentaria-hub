/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_DiagramCrops } from '../models/model_DiagramCrops';
import type { model_Edition } from '../models/model_Edition';
import type { model_EditionListResult } from '../models/model_EditionListResult';
import type { model_ImageUpload } from '../models/model_ImageUpload';
import type { model_Note } from '../models/model_Note';
import type { search_Query } from '../models/search_Query';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class EditionsService {
    /**
     * Upload Edition Image
     * Upload an image for a specific edition identified by key. The image file is provided as multipart form data.
     * @returns model_ImageUpload OK
     * @throws ApiError
     */
    public static postDatasetsImagesUpload({
        dataSetId,
        key,
        type,
        file,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Edition key
         */
        key: string,
        /**
         * Type of image (e.g., 'cover', 'facsimile')
         */
        type: string,
        /**
         * Image file to upload
         */
        file: Blob,
    }): CancelablePromise<model_ImageUpload> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/images/upload',
            path: {
                'dataSetId': dataSetId,
            },
            formData: {
                'key': key,
                'type': type,
                'file': file,
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
    /**
     * List Editions
     * Get a paginated list of editions. Filter by corpus; use offset/limit for paging.
     * @returns model_EditionListResult OK
     * @throws ApiError
     */
    public static postEditionsSearch({
        edition,
    }: {
        /**
         * Filter, ordering, and pagination options
         */
        edition?: search_Query,
    }): CancelablePromise<model_EditionListResult> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/editions/search',
            body: edition,
        });
    }
    /**
     * Get Edition by ID
     * Get a single edition by its ID.
     * @returns model_Edition OK
     * @throws ApiError
     */
    public static getEditions({
        editionId,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
    }): CancelablePromise<model_Edition> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/editions/{editionId}',
            path: {
                'editionId': editionId,
            },
            errors: {
                404: `Edition not found`,
            },
        });
    }
    /**
     * Update Edition
     * Update an existing edition identified by key. The edition data is provided in the JSON body.
     * @returns model_Edition OK
     * @throws ApiError
     */
    public static putEditions({
        editionId,
        edition,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Edition data to update
         */
        edition: model_Edition,
    }): CancelablePromise<model_Edition> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/editions/{editionId}',
            path: {
                'editionId': editionId,
            },
            body: edition,
        });
    }
    /**
     * Delete Edition
     * Delete an edition identified by ID.
     * @returns any OK
     * @throws ApiError
     */
    public static deleteEditions({
        editionId,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
    }): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/editions/{editionId}',
            path: {
                'editionId': editionId,
            },
        });
    }
    /**
     * Get Edition Diagrams
     * Get diagram image URLs for a specific edition key.
     * @returns model_DiagramCrops OK
     * @throws ApiError
     */
    public static getEditionsDiagrams({
        editionId,
    }: {
        /**
         * Edition key
         */
        editionId: string,
    }): CancelablePromise<model_DiagramCrops> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/editions/{editionId}/diagrams',
            path: {
                'editionId': editionId,
            },
        });
    }
    /**
     * Update Edition Notes
     * Update the notes for an edition identified by id. The note content is provided in the JSON body.
     * @returns model_Edition OK
     * @throws ApiError
     */
    public static postEditionsNotes({
        editionId,
        note,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Note content
         */
        note: model_Note,
    }): CancelablePromise<model_Edition> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/editions/{editionId}/notes',
            path: {
                'editionId': editionId,
            },
            body: note,
        });
    }
}
