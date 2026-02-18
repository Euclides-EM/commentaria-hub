/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_ExpectedBlocks } from '../models/annotation_ExpectedBlocks';
import type { model_Dataset } from '../models/model_Dataset';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DatasetsService {
    /**
     * List Datasets
     * Get a list of datasets with optional filtering and sorting.
     * @returns model_Dataset OK
     * @throws ApiError
     */
    public static getDatasets({
        filter,
        sort,
    }: {
        /**
         * Filter conditions
         */
        filter?: string,
        /**
         * Sort conditions
         */
        sort?: string,
    }): CancelablePromise<Array<model_Dataset>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets',
            query: {
                'filter': filter,
                'sort': sort,
            },
        });
    }
    /**
     * Create Dataset
     * Create a new dataset. Use async=true to return immediately with status "creating"; poll GET /datasets/{id} for status "ready" or "failed".
     * @returns model_Dataset OK
     * @throws ApiError
     */
    public static postDatasets({
        dataset,
        enforceSingleDataset,
        async,
    }: {
        /**
         * Dataset to create
         */
        dataset: model_Dataset,
        /**
         * If true, dataset will only be created if no other dataset exists
         */
        enforceSingleDataset?: boolean,
        /**
         * If true, return immediately and create in background (status creating → ready or failed)
         */
        async?: boolean,
    }): CancelablePromise<model_Dataset> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets',
            query: {
                'enforce_single_dataset': enforceSingleDataset,
                'async': async,
            },
            body: dataset,
        });
    }
    /**
     * Update Dataset
     * Update an existing dataset.
     * @returns model_Dataset OK
     * @throws ApiError
     */
    public static putDatasets({
        dataSetId,
        dataset,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Updated dataset
         */
        dataset: model_Dataset,
    }): CancelablePromise<model_Dataset> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}',
            path: {
                'dataSetId': dataSetId,
            },
            body: dataset,
        });
    }
    /**
     * Delete Dataset
     * Delete a dataset by its ID.
     * @returns string OK
     * @throws ApiError
     */
    public static deleteDatasets({
        dataSetId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/datasets/{dataSetId}',
            path: {
                'dataSetId': dataSetId,
            },
        });
    }
    /**
     * Get Page Image
     * Get the image for a specific page in a dataset.
     * @returns binary PNG image content
     * @throws ApiError
     */
    public static getDatasetsImages({
        dataSetId,
        pageNum,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Page Number
         */
        pageNum: string,
    }): CancelablePromise<Blob> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/images/{pageNum}',
            path: {
                'dataSetId': dataSetId,
                'pageNum': pageNum,
            },
        });
    }
    /**
     * List Suggested Annotation Reviews for Dataset
     * Get a list of suggested annotation reviews for a specific dataset.
     * @returns annotation_ExpectedBlocks OK
     * @throws ApiError
     */
    public static getDatasetsSuggestedReviews({
        dataSetId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
    }): CancelablePromise<Array<Array<annotation_ExpectedBlocks>>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/suggested_reviews',
            path: {
                'dataSetId': dataSetId,
            },
        });
    }
    /**
     * List Suggested Annotation Rules for Dataset
     * Get a list of suggested annotation rules for a specific dataset.
     * @returns any OK
     * @throws ApiError
     */
    public static getDatasetsSuggestedRules({
        dataSetId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
    }): CancelablePromise<Array<Array<any>>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/suggested_rules',
            path: {
                'dataSetId': dataSetId,
            },
        });
    }
}
