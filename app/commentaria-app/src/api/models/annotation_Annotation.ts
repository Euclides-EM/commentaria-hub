/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
export type annotation_Annotation = {
    readonly applied_rules?: Array<any>;
    readonly created_at?: string;
    readonly dataset_id?: string;
    description?: string;
    ground_truth?: boolean;
    readonly id?: string;
    name?: string;
    readonly ocred?: boolean;
    readonly origin_annotation_id?: string;
    pages?: string;
    readonly pipeline_stage?: annotationrule_PipelineStage;
    readonly segmented?: boolean;
    readonly updated_at?: string;
};

