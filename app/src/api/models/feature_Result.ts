/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_ResultSource } from './feature_ResultSource';
import type { feature_ResultValue } from './feature_ResultValue';
export type feature_Result = {
    annotation_id?: string;
    dataset_id?: string;
    feature?: string;
    key?: string;
    note?: string;
    source?: feature_ResultSource;
    values?: Array<feature_ResultValue>;
};

