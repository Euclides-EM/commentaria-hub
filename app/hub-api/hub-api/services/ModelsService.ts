/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Model } from '../models/model_Model';
import type { model_ModelTraining } from '../models/model_ModelTraining';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class ModelsService {
    /**
     * List Models
     * Get a list of available models.
     * @returns model_Model OK
     * @throws ApiError
     */
    public static getModels({
        expand,
    }: {
        /**
         * Include related entities
         */
        expand?: Array<string>,
    }): CancelablePromise<Array<model_Model>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/models',
            query: {
                'expand': expand,
            },
        });
    }
    /**
     * Train a Model
     * Submits model training to the GPU farm.
     * @returns model_ModelTraining Created
     * @throws ApiError
     */
    public static postModelsTrain({
        training,
    }: {
        /**
         * Model training details
         */
        training: model_ModelTraining,
    }): CancelablePromise<model_ModelTraining> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/models_train',
            body: training,
            mediaType: 'application/json',
        });
    }
    /**
     * Upload a Model
     * Upload a new model to the system.
     * @returns model_Model OK
     * @throws ApiError
     */
    public static postModelsUpload({
        file,
        name,
        description,
        baseAnnotations,
        baseModelId,
    }: {
        /**
         * Model file to upload
         */
        file: Blob,
        /**
         * Name of the model
         */
        name?: string,
        /**
         * Description of the model
         */
        description?: string,
        /**
         * Comma-separated list of base annotation IDs in the format <dataset_id>:<annotation_id>
         */
        baseAnnotations?: string,
        /**
         * ID of the base model this model is derived from
         */
        baseModelId?: string,
    }): CancelablePromise<model_Model> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/models_upload',
            formData: {
                'file': file,
                'name': name,
                'description': description,
                'base_annotations': baseAnnotations,
                'base_model_id': baseModelId,
            },
        });
    }
    /**
     * Update a Model
     * Update an existing model.
     * @returns model_Model OK
     * @throws ApiError
     */
    public static putModels({
        id,
        model,
    }: {
        /**
         * Model ID
         */
        id: string,
        /**
         * Updated model
         */
        model: model_Model,
    }): CancelablePromise<model_Model> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/models/{id}',
            path: {
                'id': id,
            },
            body: model,
        });
    }
    /**
     * Delete a Model
     * Delete a model by its ID.
     * @returns string OK
     * @throws ApiError
     */
    public static deleteModels({
        id,
        deep,
    }: {
        /**
         * Model ID
         */
        id: string,
        /**
         * If true, also delete the model file from filesystem
         */
        deep?: boolean,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/models/{id}',
            path: {
                'id': id,
            },
            query: {
                'deep': deep,
            },
        });
    }
}
