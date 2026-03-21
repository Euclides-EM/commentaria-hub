/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_SearchWithin } from './annotation_SearchWithin';
import type { common_ALTOPart } from './common_ALTOPart';
export type annotation_Search = {
    annotation_id?: string;
    categories?: Array<string>;
    dataset_id?: string;
    fallback_to_origin?: boolean;
    max_results?: number;
    regex?: string;
    readonly results?: Array<common_ALTOPart>;
    search_within?: Array<annotation_SearchWithin>;
};

