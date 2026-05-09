/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class BackupsService {
    /**
     * List backups
     * Returns a list of available backups with their metadata.
     * @returns string OK
     * @throws ApiError
     */
    public static getBackups(): CancelablePromise<Array<string>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/backups',
        });
    }
    /**
     * Create backup
     * Creates a new backup of the current system state.
     * @returns string Backup ID
     * @throws ApiError
     */
    public static postBackups({
        syncToDrive,
    }: {
        /**
         * If true, the backup will be copied to Google Drive using rclone after creation
         */
        syncToDrive?: boolean,
    }): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/backups',
            query: {
                'sync_to_drive': syncToDrive,
            },
        });
    }
    /**
     * Create backup from zip
     * Creates a new backup by uploading a zip file containing the backup data.
     * @returns string Backup ID
     * @throws ApiError
     */
    public static postBackupsFromzip({
        file,
    }: {
        /**
         * Zip file containing the backup data
         */
        file: Blob,
    }): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/backups/fromzip',
            formData: {
                'file': file,
            },
        });
    }
    /**
     * Download backup
     * Downloads the specified backup file.
     * @returns binary Backup file
     * @throws ApiError
     */
    public static getBackups1({
        backupId,
    }: {
        /**
         * ID of the backup to download
         */
        backupId: string,
    }): CancelablePromise<Blob> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/backups/{backupId}',
            path: {
                'backupId': backupId,
            },
        });
    }
    /**
     * Restore latest backup
     * Restores the system state from the latest available backup.
     * @returns string OK
     * @throws ApiError
     */
    public static putBackupsRestore({
        backupId,
    }: {
        /**
         * ID of the backup to restore, if the backupId is 'latest', the latest backup will be restored
         */
        backupId: string,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/backups/{backupId}/restore',
            path: {
                'backupId': backupId,
            },
        });
    }
    /**
     * Sync backup to Google Drive
     * Copies an existing backup to Google Drive using rclone.
     * @returns string OK
     * @throws ApiError
     */
    public static putBackupsSync({
        backupId,
    }: {
        /**
         * ID of the backup to sync
         */
        backupId: string,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/backups/{backupId}/sync',
            path: {
                'backupId': backupId,
            },
        });
    }
}
