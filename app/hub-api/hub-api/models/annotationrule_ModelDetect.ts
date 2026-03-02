/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
import type { common_OCRModelType } from './common_OCRModelType';
export type annotationrule_ModelDetect = {
    applicable_stages?: Array<annotationrule_PipelineStage>;
    model?: string;
    model_type?: common_OCRModelType;
    type?: annotationrule_Type;
};

