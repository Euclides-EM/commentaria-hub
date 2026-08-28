/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_EditionShelfmark } from '../models/model_EditionShelfmark';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ShelfmarksService {
    /**
     * List shelfmarks
     * Lists shelfmarks, optionally filtered by edition ID.
     * @returns model_EditionShelfmark OK
     * @throws ApiError
     */
    public static getShelfmarks({
        editionId,
    }: {
        /**
         * Filter by edition ID
         */
        editionId?: Array<string>,
    } = {}): CancelablePromise<Array<model_EditionShelfmark>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/shelfmarks',
            query: {
                'edition_id': editionId,
            },
        });
    }
    /**
     * List edition shelfmarks
     * Lists shelfmarks for an edition.
     * @returns model_EditionShelfmark OK
     * @throws ApiError
     */
    public static getEditionsShelfmarks({
        editionId,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
    }): CancelablePromise<Array<model_EditionShelfmark>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/editions/{editionId}/shelfmarks',
            path: {
                'editionId': editionId,
            },
        });
    }
    /**
     * Upsert edition shelfmark
     * Creates or updates a shelfmark for an edition.
     * @returns model_EditionShelfmark OK
     * @throws ApiError
     */
    public static postEditionsShelfmarks({
        editionId,
        shelfmark,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Shelfmark
         */
        shelfmark: model_EditionShelfmark,
    }): CancelablePromise<model_EditionShelfmark> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/editions/{editionId}/shelfmarks',
            path: {
                'editionId': editionId,
            },
            body: shelfmark,
        });
    }
    /**
     * Update edition shelfmark
     * Updates a shelfmark for an edition.
     * @returns model_EditionShelfmark OK
     * @throws ApiError
     */
    public static putEditionsShelfmarks({
        editionId,
        shelfmarkId,
        shelfmark,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Shelfmark ID
         */
        shelfmarkId: string,
        /**
         * Shelfmark
         */
        shelfmark: model_EditionShelfmark,
    }): CancelablePromise<model_EditionShelfmark> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/editions/{editionId}/shelfmarks/{shelfmarkId}',
            path: {
                'editionId': editionId,
                'shelfmarkId': shelfmarkId,
            },
            body: shelfmark,
        });
    }
    /**
     * Delete edition shelfmark
     * Deletes a shelfmark from an edition.
     * @returns any OK
     * @throws ApiError
     */
    public static deleteEditionsShelfmarks({
        editionId,
        shelfmarkId,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Shelfmark ID
         */
        shelfmarkId: string,
    }): CancelablePromise<any> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/editions/{editionId}/shelfmarks/{shelfmarkId}',
            path: {
                'editionId': editionId,
                'shelfmarkId': shelfmarkId,
            },
        });
    }
}
