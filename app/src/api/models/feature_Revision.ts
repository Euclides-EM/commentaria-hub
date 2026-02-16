/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { common_Reference } from './common_Reference';
import type { feature_ExecutionStrategy } from './feature_ExecutionStrategy';
import type { feature_Type } from './feature_Type';
export type feature_Revision = {
    readonly created_at?: string;
    dataset_id?: string;
    description?: string;
    execution_strategy?: feature_ExecutionStrategy;
    feature_id?: string;
    /**
     * Features is relevant only if the parent feature is root, in which case it lists the features that are part of this revision.
     */
    readonly features?: Array<common_Reference>;
    readonly id?: string;
    name?: string;
    note?: string;
    /**
     * Prompt is relevant only if the execution strategy is ExecutionStrategy.
     */
    prompt?: string;
    /**
     * Regex is relevant only if the execution strategy is ExecutionStrategyRegex.
     */
    regex?: string;
    type?: feature_Type;
    readonly updated_at?: string;
};

