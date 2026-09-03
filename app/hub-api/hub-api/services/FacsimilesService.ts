/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { job_Job } from '../models/job_Job';
import type { model_DiagramCrops } from '../models/model_DiagramCrops';
import type { model_Facsimile } from '../models/model_Facsimile';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class FacsimilesService {
    /**
     * List Facsimiles (bulk get)
     * Get facsimiles, optionally filtered by edition ID.
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static getFacsimilies({
        editionId,
    }: {
        /**
         * Filter by edition ID
         */
        editionId?: Array<string>,
    }): CancelablePromise<Array<model_Facsimile>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/facsimilies',
            query: {
                'edition_id': editionId,
            },
        });
    }
    /**
     * Create Facsimile
     * Create a new facsimile
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static postFacsimilies({
        facsimile,
    }: {
        /**
         * Facsimile to create
         */
        facsimile: model_Facsimile,
    }): CancelablePromise<model_Facsimile> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/facsimilies',
            body: facsimile,
        });
    }
    /**
     * Import facsimiles and diagram crops from Google Drive inbox
     * Copies PDFs from the configured Google Drive folder into FACSIMILES_PDF_DIR, installs diagram crop archives into FACSIMILES_DIAGRAMS_PATH, updates metadata, then deletes successfully imported files from Drive.
     * @returns job_Job OK
     * @throws ApiError
     */
    public static postFacsimiliesImportFromDrive({
        async,
    }: {
        /**
         * Create a background import job instead of waiting for completion
         */
        async?: boolean,
    }): CancelablePromise<job_Job> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/facsimilies/import-from-drive',
            query: {
                'async': async,
            },
        });
    }
    /**
     * Download facsimile mapping CSVs
     * Downloads a ZIP containing facsimiles.csv and shelfmarks.csv for facsimile-to-shelfmark mapping.
     * @returns binary Facsimile mapping ZIP
     * @throws ApiError
     */
    public static getFacsimiliesMappingCsv(): CancelablePromise<Blob> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/facsimilies/mapping-csv',
        });
    }
    /**
     * Upload facsimile mapping CSV
     * Uploads an edited facsimiles.csv file and updates facsimile shelfmark mappings.
     * @returns string OK
     * @throws ApiError
     */
    public static postFacsimiliesMappingCsv({
        file,
    }: {
        /**
         * facsimiles.csv
         */
        file: Blob,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/facsimilies/mapping-csv',
            formData: {
                'file': file,
            },
        });
    }
    /**
     * Get Facsimile by ID
     * Get a single facsimile by its ID.
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static getFacsimilies1({
        id,
    }: {
        /**
         * Facsimile ID
         */
        id: string,
    }): CancelablePromise<model_Facsimile> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/facsimilies/{id}',
            path: {
                'id': id,
            },
            errors: {
                404: `Facsimile not found`,
            },
        });
    }
    /**
     * Update Facsimile
     * Update an existing facsimile identified by ID.
     * @returns model_Facsimile OK
     * @throws ApiError
     */
    public static putFacsimilies({
        id,
        facsimile,
    }: {
        /**
         * Facsimile ID
         */
        id: string,
        /**
         * Facsimile data to update
         */
        facsimile: model_Facsimile,
    }): CancelablePromise<model_Facsimile> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/facsimilies/{id}',
            path: {
                'id': id,
            },
            body: facsimile,
        });
    }
    /**
     * Get Facsimile Diagrams
     * Get diagram image URLs for a specific facsimile.
     * @returns model_DiagramCrops OK
     * @throws ApiError
     */
    public static getFacsimiliesDiagrams({
        id,
    }: {
        /**
         * Facsimile ID
         */
        id: string,
    }): CancelablePromise<model_DiagramCrops> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/facsimilies/{id}/diagrams',
            path: {
                'id': id,
            },
        });
    }
    /**
     * Download facsimile PDF
     * Downloads a local facsimile PDF by facsimile ID.
     * @returns binary Facsimile PDF
     * @throws ApiError
     */
    public static getFacsimiliesPdf({
        id,
    }: {
        /**
         * Facsimile ID
         */
        id: string,
    }): CancelablePromise<Blob> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/facsimilies/{id}/pdf',
            path: {
                'id': id,
            },
        });
    }
}
