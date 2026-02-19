/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_ExecutionStrategy } from './feature_ExecutionStrategy';
export type feature_Revision = {
    readonly created_at?: string;
    dataset_id?: string;
    description?: string;
    execution_strategy?: feature_ExecutionStrategy;
    feature_id?: string;
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
    readonly updated_at?: string;
};

