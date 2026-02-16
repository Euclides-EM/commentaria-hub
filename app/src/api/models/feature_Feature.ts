/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { feature_Revision } from './feature_Revision';
export type feature_Feature = {
    /**
     * Color is an optional UI color hint for this feature, e.g. "#FF0000" for red.
     */
    color?: string;
    readonly created_at?: string;
    dataset_id?: string;
    description?: string;
    readonly id?: string;
    is_default?: boolean;
    /**
     * IsRoot is immutable.
     */
    is_root?: boolean;
    /**
     * LatestRevision is the most recent revision of this feature. It is read-only and only included if expand=latest_revision is specified in the request.
     */
    readonly latest_revision?: feature_Revision;
    name?: string;
    /**
     * Revisions is the list of all revisions of this feature, ordered by created_at descending. It is read-only and only included if expand=revisions is specified in the request.
     */
    readonly revisions?: Array<feature_Revision>;
    readonly updated_at?: string;
};

