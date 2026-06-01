/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_Reference } from './annotation_Reference';
import type { annotationrule_ApplyRules } from './annotationrule_ApplyRules';
import type { job_Status } from './job_Status';
import type { job_Target } from './job_Target';
import type { job_Task } from './job_Task';
import type { model_ModelTraining } from './model_ModelTraining';
export type job_Job = {
    annotation?: annotation_Reference;
    readonly created_at?: string;
    description?: string;
    readonly details?: string;
    effectiveAnnotation?: annotation_Reference;
    readonly finished_at?: string;
    readonly id?: string;
    model_training?: model_ModelTraining;
    name?: string;
    rules?: annotationrule_ApplyRules;
    readonly status?: job_Status;
    target?: job_Target;
    task?: job_Task;
    readonly updated_at?: string;
};

