/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { common_FilesystemStatus } from './common_FilesystemStatus';
export type common_HealthStatus = {
    commit_sha?: string;
    db_ready?: boolean;
    filesystem?: common_FilesystemStatus;
    filesystem_error?: string;
};

