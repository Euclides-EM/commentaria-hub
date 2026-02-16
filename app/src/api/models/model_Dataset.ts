/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type model_Dataset = {
    readonly created_at?: string;
    /**
     * set when status is "failed"
     */
    creation_error?: string;
    description?: string;
    deskewed?: boolean;
    dpi?: number;
    edition_id?: string;
    facsimile_id?: string;
    readonly id?: string;
    name?: string;
    /**
     * "creating" | "ready" | "failed"
     */
    status?: string;
    readonly updated_at?: string;
};

