/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_MetadataDetails } from '../models/annotationrule_MetadataDetails';
import type { annotationrule_PipelineStage } from '../models/annotationrule_PipelineStage';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class MetadataService {
    /**
     * List Annotation Rules
     * Lists all available annotation rules with their default configurations.
     * @returns annotationrule_MetadataDetails OK
     * @throws ApiError
     */
    public static getAnnotationRules(): CancelablePromise<Array<Array<annotationrule_MetadataDetails>>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/annotation_rules',
        });
    }
    /**
     * List Pipeline Stages
     * Lists all defined pipeline stages for annotations.
     * @returns annotationrule_PipelineStage OK
     * @throws ApiError
     */
    public static getPipelineStages(): CancelablePromise<Array<Array<annotationrule_PipelineStage>>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/pipeline_stages',
        });
    }
}
