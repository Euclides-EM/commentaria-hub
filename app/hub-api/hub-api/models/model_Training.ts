/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_Reference } from './annotation_Reference';
import type { model_Model } from './model_Model';
import type { model_TrainingStatus } from './model_TrainingStatus';
export type model_Training = {
    annotation_sets?: Array<annotation_Reference>;
    readonly created_at?: string;
    description?: string;
    readonly id?: string;
    name?: string;
    origin_model?: model_Model;
    status?: model_TrainingStatus;
    readonly updated_at?: string;
};

