/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class GpuFarmService {
    /**
     * Download manual GPU farm run bundle
     * Downloads the bundle of a GPU farm run submitted with manual_run (annotation detection rules and model training).
     * @returns binary Manual run bundle ZIP
     * @throws ApiError
     */
    public static getGpuFarmManualBundles({
        runId,
    }: {
        /**
         * GPU farm run ID
         */
        runId: string,
    }): CancelablePromise<Blob> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/gpu_farm/manual_bundles/{runId}',
            path: {
                'runId': runId,
            },
        });
    }
}
