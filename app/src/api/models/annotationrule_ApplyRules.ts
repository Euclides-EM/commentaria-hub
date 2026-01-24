/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_ApplyRulesAction } from './annotationrule_ApplyRulesAction';
export type annotationrule_ApplyRules = {
    action?: annotationrule_ApplyRulesAction;
    /**
     * Description is used only if the action is ApplyRulesActionCreateNew
     */
    description?: string;
    /**
     * Name is used only if the action is ApplyRulesActionCreateNew
     */
    name?: string;
    rules?: Array<any>;
};

