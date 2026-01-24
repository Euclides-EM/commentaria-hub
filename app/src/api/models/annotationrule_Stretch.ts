/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_ContactSide } from './annotationrule_ContactSide';
import type { annotationrule_ContactType } from './annotationrule_ContactType';
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
export type annotationrule_Stretch = {
    applicable_stages?: Array<annotationrule_PipelineStage>;
    contact_side?: annotationrule_ContactSide;
    contact_type?: annotationrule_ContactType;
    stretch_category?: string;
    towards?: string;
    type?: annotationrule_Type;
};

