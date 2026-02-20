/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_City } from '../models/model_City';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class GeoDataService {
    /**
     * List cities
     * Returns cities metadata with city name, longitude, and latitude.
     * @returns model_City OK
     * @throws ApiError
     */
    public static getCities(): CancelablePromise<Array<model_City>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/cities',
        });
    }
}
