/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_EditionTranscription } from '../models/model_EditionTranscription';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class TranscriptionsService {
    /**
     * List Transcriptions
     * Retrieve a list of transcriptions, optionally filtered by edition or facsimile.
     * @returns model_EditionTranscription List of transcriptions
     * @throws ApiError
     */
    public static getEditionsTranscriptions({
        editionId,
    }: {
        /**
         * Filter by edition ID
         */
        editionId?: Array<string>,
    }): CancelablePromise<Array<model_EditionTranscription>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/editions/transcriptions',
            query: {
                'edition_id': editionId,
            },
        });
    }
    /**
     * Update Preferred Transcription
     * Update the preferred transcription for a specific edition.
     * @returns model_EditionTranscription Updated preferred transcription
     * @throws ApiError
     */
    public static putEditionsTranscriptions({
        editionId,
        body,
    }: {
        /**
         * Edition ID
         */
        editionId: string,
        /**
         * Preferred transcription details
         */
        body: model_EditionTranscription,
    }): CancelablePromise<model_EditionTranscription> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/editions/{editionId}/transcriptions',
            path: {
                'edition_id': editionId,
            },
            body: body,
        });
    }
}
