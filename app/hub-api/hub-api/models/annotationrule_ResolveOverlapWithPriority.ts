/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
export type annotationrule_ResolveOverlapWithPriority = {
    applicable_stages?: Array<annotationrule_PipelineStage>;
    dominant_category?: string;
    /**
     * percentage 0–1
     */
    min_overlap?: number;
    suppressed_category?: string;
    type?: annotationrule_Type;
};

