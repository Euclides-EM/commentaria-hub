/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_Group } from '../models/annotation_Group';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AnnotationGroupsService {
    /**
     * List Annotation Groups
     * Get a list of all annotation groups
     * @returns annotation_Group OK
     * @throws ApiError
     */
    public static getAnnotationGroups(): CancelablePromise<Array<annotation_Group>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/annotation_groups',
        });
    }
    /**
     * Create Annotation Group
     * Create a new annotation group with the provided details
     * @returns annotation_Group OK
     * @throws ApiError
     */
    public static postAnnotationGroups({
        group,
    }: {
        /**
         * Annotation Group data
         */
        group: annotation_Group,
    }): CancelablePromise<annotation_Group> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/annotation_groups',
            body: group,
        });
    }
    /**
     * Get Annotation Group
     * Get details of a specific annotation group by ID
     * @returns annotation_Group OK
     * @throws ApiError
     */
    public static getAnnotationGroups1({
        groupId,
    }: {
        /**
         * Annotation Group ID
         */
        groupId: string,
    }): CancelablePromise<annotation_Group> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/annotation_groups/{groupId}',
            path: {
                'groupId': groupId,
            },
            errors: {
                404: `Not Found`,
            },
        });
    }
    /**
     * Update Annotation Group
     * Update details of an existing annotation group by ID
     * @returns annotation_Group OK
     * @throws ApiError
     */
    public static putAnnotationGroups({
        groupId,
        group,
    }: {
        /**
         * Annotation Group ID
         */
        groupId: string,
        /**
         * Updated Annotation Group data
         */
        group: annotation_Group,
    }): CancelablePromise<annotation_Group> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/annotation_groups/{groupId}',
            path: {
                'groupId': groupId,
            },
            body: group,
            errors: {
                404: `Not Found`,
            },
        });
    }
    /**
     * Delete Annotation Group
     * Delete an annotation group by ID
     * @returns void
     * @throws ApiError
     */
    public static deleteAnnotationGroups({
        groupId,
    }: {
        /**
         * Annotation Group ID
         */
        groupId: string,
    }): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/annotation_groups/{groupId}',
            path: {
                'groupId': groupId,
            },
            errors: {
                404: `Not Found`,
            },
        });
    }
}
