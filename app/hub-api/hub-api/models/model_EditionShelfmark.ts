/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_EditionShelfmarkStructuralMetadataAvailability } from './model_EditionShelfmarkStructuralMetadataAvailability';
import type { model_EditionShelfmarkTranscriptionAvailability } from './model_EditionShelfmarkTranscriptionAvailability';
export type model_EditionShelfmark = {
    annotations?: string;
    copyright?: string;
    edition_id?: string;
    frontispiece_img?: string;
    id?: string;
    note?: string;
    scan?: string;
    shelfmark?: string;
    structural_metadata_available?: model_EditionShelfmarkStructuralMetadataAvailability;
    title_page_img?: string;
    transcription_available?: model_EditionShelfmarkTranscriptionAvailability;
    volume?: number;
};
