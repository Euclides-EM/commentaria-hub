/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
import type { llm_Usage } from './llm_Usage';
export type annotationrule_LLMTranscriptionCorrector = {
    additional_annotations?: Array<string>;
    applicable_stages?: Array<annotationrule_PipelineStage>;
    include_edition_transcription?: boolean;
    model?: string;
    pages?: string;
    provider?: string;
    rounds?: number;
    type?: annotationrule_Type;
    readonly usage?: llm_Usage;
};

