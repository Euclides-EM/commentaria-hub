/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_DefScope } from './feature_DefScope';
export type feature_Revision = {
    ai_model?: string;
    ai_provider?: 'openai' | 'ollama';
    categorizer?: string;
    readonly created_at?: string;
    description?: string;
    feature_id?: string;
    readonly id?: string;
    name?: string;
    prompt?: string;
    scope?: feature_DefScope;
    readonly updated_at?: string;
};
