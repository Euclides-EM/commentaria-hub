/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_Part } from './annotation_Part';
export type annotation_Search = {
    annotation_id?: string;
    categories?: Array<string>;
    dataset_id?: string;
    max_results?: number;
    regex?: string;
    readonly results?: Array<annotation_Part>;
};

