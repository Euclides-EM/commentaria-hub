/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_AnnotationExpectedBlocksSanityType } from './model_AnnotationExpectedBlocksSanityType';
import type { model_SuggestedDiff } from './model_SuggestedDiff';
export type model_AnnotationExpectedBlocks = {
    category?: string;
    expected_blocks?: Array<Array<string>>;
    failed_checks?: Array<model_AnnotationExpectedBlocksSanityType>;
    sanity_checks?: Array<model_AnnotationExpectedBlocksSanityType>;
    suggested_diffs?: Array<model_SuggestedDiff>;
};

