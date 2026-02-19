/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { common_AuthValidateResponse } from '../models/common_AuthValidateResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AuthenticationService {
    /**
     * Validate authentication token
     * Validates the provided Bearer token and returns user information
     * @returns common_AuthValidateResponse OK
     * @throws ApiError
     */
    public static postAuthValidate(): CancelablePromise<common_AuthValidateResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/auth/validate',
            errors: {
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
}
