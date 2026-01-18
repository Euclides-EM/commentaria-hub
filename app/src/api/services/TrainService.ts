/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Training } from '../models/model_Training';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class TrainService {
    /**
     * Train Model
     * Train a new model based on the provided configuration.
     * @returns model_Training OK
     * @throws ApiError
     */
    public static postTrain({
        model,
    }: {
        /**
         * Training Configuration
         */
        model: model_Training,
    }): CancelablePromise<model_Training> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/train',
            body: model,
        });
    }
}
