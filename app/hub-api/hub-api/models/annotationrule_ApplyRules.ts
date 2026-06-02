/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_ApplyRulesAction } from './annotationrule_ApplyRulesAction';
import type { annotationrule_ExecutionMode } from './annotationrule_ExecutionMode';
export type annotationrule_ApplyRules = {
    action?: annotationrule_ApplyRulesAction;
    /**
     * CopyFeatureResults is used only if the action is ApplyRulesActionCreateNew. If true, the feature results of the original annotation will be copied to the new annotation.
     */
    copy_feature_results?: boolean;
    /**
     * Description is used only if the action is ApplyRulesActionCreateNew
     */
    description?: string;
    execution_mode?: annotationrule_ExecutionMode;
    /**
     * Name is used only if the action is ApplyRulesActionCreateNew
     */
    name?: string;
    rules?: Array<any>;
};

