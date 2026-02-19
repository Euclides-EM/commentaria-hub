/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_VCSStatus } from '../models/model_VCSStatus';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class VersionControlService {
    /**
     * Pull latest changes from the repository
     * Pulls the latest changes from the repository. If the current branch is not main, it will check out main before pulling. Requires GitHub token in the Authorization header.
     * @returns model_VCSStatus OK
     * @throws ApiError
     */
    public static postVersionControlPull(): CancelablePromise<model_VCSStatus> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/version_control/pull',
        });
    }
    /**
     * Push local changes to the repository
     * Pushes local changes to the repository. This will first pull the latest changes to ensure the local branch is up to date, then push any local commits. Requires GitHub token in the Authorization header.
     * @returns model_VCSStatus OK
     * @throws ApiError
     */
    public static postVersionControlPush(): CancelablePromise<model_VCSStatus> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/version_control/push',
        });
    }
}
