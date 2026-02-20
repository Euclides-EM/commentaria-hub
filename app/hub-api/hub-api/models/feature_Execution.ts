/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_ExecutionApplyItem } from './feature_ExecutionApplyItem';
import type { feature_ExecutionPolicy } from './feature_ExecutionPolicy';
import type { feature_ExecutionStatus } from './feature_ExecutionStatus';
export type feature_Execution = {
    annotation_id?: string;
    apply?: Array<feature_ExecutionApplyItem>;
    readonly created_at?: string;
    dataset_id?: string;
    description?: string;
    readonly id?: string;
    /**
     * Keys is optional, if not provided, the execution will run on all keys of the dataset.
     */
    keys?: Array<string>;
    name?: string;
    policy?: feature_ExecutionPolicy;
    status?: feature_ExecutionStatus;
    readonly updated_at?: string;
};

