/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type model_ImageMetadata = {
    /**
     * Filename is the name of the image file, which may not be unique across the dataset. It is unique across the dataset.
     */
    filename?: string;
    /**
     * Key can be either a page number (as a string) or an image key, depending on how the image was uploaded and identified in the dataset.
     * The key is not necessarily unique across the entire dataset.
     */
    key?: string;
    modified_at?: string;
};

