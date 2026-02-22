/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_ImageMetadata } from '../models/model_ImageMetadata';
import type { model_ImageUpload } from '../models/model_ImageUpload';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DatasetImagesService {
    /**
     * Get Dataset Images
     * Get a list of images associated with a dataset.
     * @returns model_ImageMetadata OK
     * @throws ApiError
     */
    public static getDatasetsImages({
        dataSetId,
        uniqueOnly,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * If true, return only unique images (one per page number or key), otherwise return all images including duplicates for different keys
         */
        uniqueOnly?: boolean,
    }): CancelablePromise<Array<model_ImageMetadata>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/images',
            path: {
                'dataSetId': dataSetId,
            },
            query: {
                'uniqueOnly': uniqueOnly,
            },
        });
    }
    /**
     * Delete Dataset Image
     * Delete a specific image associated with a dataset.
     * @returns string OK
     * @throws ApiError
     */
    public static deleteDatasetsImages({
        dataSetId,
        pageNumOrKey,
        filename,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Page number or image key to identify the image to delete
         */
        pageNumOrKey?: Array<string>,
        /**
         * Filename of the image to delete (optional, used if pageNumOrKey is not sufficient to identify the image)
         */
        filename?: Array<string>,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/datasets/{dataSetId}/images',
            path: {
                'dataSetId': dataSetId,
            },
            query: {
                'pageNumOrKey': pageNumOrKey,
                'filename': filename,
            },
        });
    }
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
     * Get Page Image
     * Get the image for a specific page in a dataset.
     * @returns binary PNG image content
     * @throws ApiError
     */
    public static getDatasetsImages1({
        dataSetId,
        pageNumOrKey,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Page Number
         */
        pageNumOrKey: string,
    }): CancelablePromise<Blob> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/images/{pageNumOrKey}',
            path: {
                'dataSetId': dataSetId,
                'pageNumOrKey': pageNumOrKey,
            },
        });
    }
}
