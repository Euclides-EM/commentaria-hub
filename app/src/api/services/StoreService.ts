/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class StoreService {
    /**
     * Cleanup Local Store
     * Cleans up the local store by removing temporary files and unused data.
     * @returns void
     * @throws ApiError
     */
    public static deleteStoreCleanupLocal({
        dryRun,
    }: {
        /**
         * If true, performs a dry run without deleting files
         */
        dryRun?: boolean,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/store/cleanup/local',
            query: {
                'dry_run': dryRun,
            },
        });
    }
}
