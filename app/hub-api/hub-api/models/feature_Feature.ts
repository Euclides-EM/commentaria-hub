/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Revision } from './feature_Revision';
import type { feature_DefScope } from './feature_DefScope';
export type feature_Feature = {
    color?: string;
    readonly created_at?: string;
    description?: string;
    readonly id?: string;
    is_default?: boolean;
    is_list?: boolean;
    /**
     * LatestRevision is the most recent revision of this feature. It is read-only and only included if expand=latest_revision is specified in the request.
     */
    readonly latest_revision?: feature_Revision;
    name?: string;
    properties?: Array<string>;
    scope?: feature_DefScope;
    /**
     * Revisions is the list of all revisions of this feature, ordered by created_at descending. It is read-only and only included if expand=revisions is specified in the request.
     */
    readonly revisions?: Array<feature_Revision>;
    readonly updated_at?: string;
};
