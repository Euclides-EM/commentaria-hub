/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_FacsimileConnectionConfirmationStatus } from './model_FacsimileConnectionConfirmationStatus';
export type model_Facsimile = {
    facsimile_connection_confirmation_status?: model_FacsimileConnectionConfirmationStatus;
    readonly created_at?: string;
    description?: string;
    diagram_crops_available?: boolean;
    readonly download_available?: boolean;
    edition_id?: string;
    file_size_bytes?: number;
    readonly id?: string;
    imported_at?: string;
    main_text_pages?: string;
    name?: string;
    scan_url?: string;
    shelfmark_id?: string;
    readonly updated_at?: string;
};
