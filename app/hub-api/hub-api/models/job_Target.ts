/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { job_Platform } from './job_Platform';
export type job_Target = {
    /**
     * For AnnotationRuleApply
     */
    annotation_id?: string;
    /**
     * For Roboflow and Commentaria
     */
    api_key?: string;
    /**
     * For backups
     */
    backup_id?: string;
    base_path?: string;
    /**
     * For Commentaria and AnnotationRuleApply
     */
    dataset_id?: string;
    document?: string;
    is_not_ground_truth?: boolean;
    password?: string;
    platform?: job_Platform;
    project_id?: string;
    sync_to_drive?: boolean;
    /**
     * For EScriptorium
     */
    username?: string;
    /**
     * For Roboflow
     */
    workspace_url?: string;
};

