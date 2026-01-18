/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_AnnotationPart } from './model_AnnotationPart';
export type model_AnnotationSearch = {
    annotation_id?: string;
    categories?: Array<string>;
    dataset_id?: string;
    max_results?: number;
    regex?: string;
    readonly results?: Array<model_AnnotationPart>;
};

