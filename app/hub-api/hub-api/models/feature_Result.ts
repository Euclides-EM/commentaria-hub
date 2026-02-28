/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_ResultSource } from './feature_ResultSource';
import type { feature_ResultValue } from './feature_ResultValue';
export type feature_Result = {
    annotation_id?: string;
    readonly created_at?: string;
    dataset_id?: string;
    description?: string;
    feature_id?: string;
    readonly id?: string;
    name?: string;
    page_key?: string;
    /**
     * Source indicates the origin of the value, such as which OCR process or manual correction it came from. This is important for traceability and debugging.
     */
    source?: feature_ResultSource;
    readonly updated_at?: string;
    /**
     * Values is a list of all the values that were extracted for this feature. There may be multiple values if the feature appears multiple times in the document.
     */
    values?: Array<feature_ResultValue>;
};

