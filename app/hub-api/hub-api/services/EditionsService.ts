/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_ApplyReprints } from '../models/model_ApplyReprints';
import type { model_DiagramCrops } from '../models/model_DiagramCrops';
import type { model_Edition } from '../models/model_Edition';
import type { model_EditionListResult } from '../models/model_EditionListResult';
import type { model_Note } from '../models/model_Note';
import type { model_ReprintDetection } from '../models/model_ReprintDetection';
import type { search_Query } from '../models/search_Query';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class EditionsService {
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
     * Apply reviewed reprint relationships in bulk
     * @returns model_ApplyReprints OK
     * @throws ApiError
     */
    public static postEditionsReprintsApply({
        request,
    }: {
        /**
         * Approved relationships
         */
        request: model_ApplyReprints,
    }): CancelablePromise<model_ApplyReprints> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/editions/reprints/apply',
            body: request,
        });
    }
    /**
     * Detect likely reprints without modifying the catalog
     * @returns model_ReprintDetection OK
     * @throws ApiError
     */
    public static postEditionsReprintsDetect(): CancelablePromise<model_ReprintDetection> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/editions/reprints/detect',
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
    /**
     * Get Edition TEI
     * Get the TEI representation of a specific edition for a specific page.
     * @returns string TEI XML content
     * @throws ApiError
     */
    public static getEditionsTei({
        editionId,
        pageNum,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Page Number
         */
        pageNum: string,
    }): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/editions/{editionId}/tei/{pageNum}',
            path: {
                'editionId': editionId,
                'pageNum': pageNum,
            },
        });
    }
}
