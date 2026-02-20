/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
export type annotationrule_ReassignTextLinesByTolerance = {
    applicable_stages?: Array<annotationrule_PipelineStage>;
    from_category?: string;
    /**
     * MinOverlap is the minimum overlap ratio (0.0 to 1.0) required to reassign a text line.
     * For example, a value of 0.8 means that at least 80% of the text line's width must overlap with the target category's bounding box to be reassigned.
     */
    min_overlap?: number;
    precision_px?: number;
    to_category?: string;
    type?: annotationrule_Type;
};

