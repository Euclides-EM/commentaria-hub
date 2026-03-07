/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_Reference } from './annotation_Reference';
import type { integration_JobStatus } from './integration_JobStatus';
import type { integration_JobTarget } from './integration_JobTarget';
import type { integration_Task } from './integration_Task';
export type integration_Job = {
    annotation?: annotation_Reference;
    readonly created_at?: string;
    description?: string;
    readonly details?: string;
    readonly finished_at?: string;
    readonly id?: string;
    name?: string;
    readonly status?: integration_JobStatus;
    target?: integration_JobTarget;
    task?: integration_Task;
    readonly updated_at?: string;
};

