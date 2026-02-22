/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { common_Reference } from './common_Reference';
import type { feature_Revision } from './feature_Revision';
import type { feature_Type } from './feature_Type';
export type feature_Feature = {
    /**
     * Color is an optional UI color hint for this feature, e.g. "#FF0000" for red.
     */
    color?: string;
    readonly created_at?: string;
    dataset_id?: string;
    description?: string;
    /**
     * Features is relevant only if this feature is root; it lists the child features that are part of this feature.
     */
    features?: Array<common_Reference>;
    readonly id?: string;
    is_default?: boolean;
    /**
     * IsList indicates whether this feature can have multiple values (e.g. a list of named entities) or just a single value (e.g. an annotation). It is immutable.
     */
    is_list?: boolean;
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
    /**
     * Type is immutable and determines the type of this feature, e.g. annotation or NER.
     */
    type?: feature_Type;
    readonly updated_at?: string;
};

