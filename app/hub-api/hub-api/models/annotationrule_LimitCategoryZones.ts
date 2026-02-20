/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_KeepPosition } from './annotationrule_KeepPosition';
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
export type annotationrule_LimitCategoryZones = {
    applicable_stages?: Array<annotationrule_PipelineStage>;
    category?: string;
    keep_position?: annotationrule_KeepPosition;
    max_count?: number;
    type?: annotationrule_Type;
};

