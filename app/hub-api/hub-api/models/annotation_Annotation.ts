/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_MergedReference } from './annotation_MergedReference';
import type { annotation_TranscriptionFallback } from './annotation_TranscriptionFallback';
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
export type annotation_Annotation = {
    readonly applied_rules?: Array<any>;
    readonly created_at?: string;
    readonly dataset_id?: string;
    description?: string;
    ground_truth?: boolean;
    hidden?: boolean;
    readonly id?: string;
    readonly lines_detected?: boolean;
    readonly merged_annotations?: Array<annotation_MergedReference>;
    name?: string;
    readonly ocred?: boolean;
    readonly origin_annotation_id?: string;
    pages?: string;
    readonly pipeline_stage?: annotationrule_PipelineStage;
    readonly segmented?: boolean;
    readonly transcription_fallback?: annotation_TranscriptionFallback;
    readonly updated_at?: string;
};
