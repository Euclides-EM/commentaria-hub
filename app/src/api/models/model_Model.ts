/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { common_OCRModelType } from './common_OCRModelType';
import type { model_AnnotationReference } from './model_AnnotationReference';
import type { model_OCRModelAlgorithmFamily } from './model_OCRModelAlgorithmFamily';
import type { model_OCRModelLocation } from './model_OCRModelLocation';
export type model_Model = {
    algorithm_family?: model_OCRModelAlgorithmFamily;
    base_annotations?: Array<model_AnnotationReference>;
    base_model_id?: string;
    categories?: Array<string>;
    readonly created_at?: string;
    description?: string;
    readonly id?: string;
    /**
     * LocalPath is the path to the model file on the local filesystem. It is relevant only for local models.
     */
    readonly local_path?: string;
    location?: model_OCRModelLocation;
    name?: string;
    type?: common_OCRModelType;
    readonly updated_at?: string;
};

