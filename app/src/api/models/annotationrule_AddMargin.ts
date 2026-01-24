/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_ContactSide } from './annotationrule_ContactSide';
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
export type annotationrule_AddMargin = {
    applicable_stages?: Array<annotationrule_PipelineStage>;
    category?: string;
    margin?: number;
    sides?: annotationrule_ContactSide;
    type?: annotationrule_Type;
};

