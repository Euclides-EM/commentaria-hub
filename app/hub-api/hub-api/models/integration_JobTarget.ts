/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { integration_Platform } from './integration_Platform';
export type integration_JobTarget = {
    /**
     * For Roboflow and Commentaria
     */
    api_key?: string;
    base_path?: string;
    dataset_id?: string;
    document?: string;
    is_not_ground_truth?: boolean;
    password?: string;
    platform?: integration_Platform;
    project_id?: string;
    /**
     * For EScriptorium
     */
    username?: string;
    workspace_url?: string;
};

