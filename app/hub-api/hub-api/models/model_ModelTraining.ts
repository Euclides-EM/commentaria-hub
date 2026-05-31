/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Model } from './model_Model';
export type model_ModelTraining = {
    backend?: string;
    readonly created_at?: string;
    description?: string;
    epochs?: number;
    gpu_farm_host?: string;
    readonly id?: string;
    model?: model_Model;
    name?: string;
    remote_run_dir?: string;
    status?: string;
    status_details?: Record<string, string>;
    readonly updated_at?: string;
};

