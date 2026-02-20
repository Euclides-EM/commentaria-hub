/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_ExpectedBlocksSanityType } from './annotation_ExpectedBlocksSanityType';
import type { annotation_SuggestedDiff } from './annotation_SuggestedDiff';
export type annotation_ExpectedBlocks = {
    category?: string;
    expected_blocks?: Array<Array<string>>;
    failed_checks?: Array<annotation_ExpectedBlocksSanityType>;
    sanity_checks?: Array<annotation_ExpectedBlocksSanityType>;
    suggested_diffs?: Array<annotation_SuggestedDiff>;
};

