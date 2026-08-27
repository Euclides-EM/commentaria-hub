/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_Annotation } from '../models/annotation_Annotation';
import type { annotationrule_AddMargin } from '../models/annotationrule_AddMargin';
import type { annotationrule_ApplyRules } from '../models/annotationrule_ApplyRules';
import type { annotationrule_LimitCategoryZones } from '../models/annotationrule_LimitCategoryZones';
import type { annotationrule_LinesDetect } from '../models/annotationrule_LinesDetect';
import type { annotationrule_ModelDetect } from '../models/annotationrule_ModelDetect';
import type { annotationrule_ReassignTextLinesByTolerance } from '../models/annotationrule_ReassignTextLinesByTolerance';
import type { annotationrule_RecategorizeByAlignment } from '../models/annotationrule_RecategorizeByAlignment';
import type { annotationrule_RemoveCategories } from '../models/annotationrule_RemoveCategories';
import type { annotationrule_RemoveOverlap } from '../models/annotationrule_RemoveOverlap';
import type { annotationrule_RenameCategories } from '../models/annotationrule_RenameCategories';
import type { annotationrule_ResolveOverlapWithPriority } from '../models/annotationrule_ResolveOverlapWithPriority';
import type { annotationrule_SlicePages } from '../models/annotationrule_SlicePages';
import type { annotationrule_Stretch } from '../models/annotationrule_Stretch';
import type { annotationrule_TextBlockCorrections } from '../models/annotationrule_TextBlockCorrections';
import type { job_Job } from '../models/job_Job';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AnnotationsApplyRulesService {
    /**
     * Apply Rules to Annotation
     * Apply specific rules to an annotation.
     * @returns annotation_Annotation OK
     * @returns job_Job Accepted
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApply({
        dataSetId,
        id,
        annotationApplyRules,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Annotation apply rules
         */
        annotationApplyRules: annotationrule_ApplyRules,
    }): CancelablePromise<annotation_Annotation | job_Job> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            body: annotationApplyRules,
        });
    }
    /**
     * Add Margin Rule to Annotation
     * add margin to an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyAddMargin({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Annotation add margin rule
         */
        annotationRule: annotationrule_AddMargin,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/add_margin',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Detect Lines in Annotation
     * Detect lines in an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyDetectLines({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Annotation detect lines rule
         */
        annotationRule: annotationrule_LinesDetect,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/detect_lines',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Limit Category Zones in Annotation
     * Keep at most max_count zones of a category per page, keeping those closest to the specified page side (top/bottom/left/right).
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyLimitCategoryZones({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Limit category zones rule
         */
        annotationRule: annotationrule_LimitCategoryZones,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/limit_category_zones',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Apply ModelDetect Rule to Annotation
     * Apply a model detect rule to an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyModelDetect({
        dataSetId,
        id,
        annotationModelDetectRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Annotation model detect rule
         */
        annotationModelDetectRule: annotationrule_ModelDetect,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/model_detect',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationModelDetectRule,
        });
    }
    /**
     * Reassign Text Lines by Tolerance in Annotation
     * Reassign text lines by tolerance in an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyReassignTextLinesByTolerance({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Reassign text lines by tolerance rule
         */
        annotationRule: annotationrule_ReassignTextLinesByTolerance,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/reassign_text_lines_by_tolerance',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Recategorize by Alignment in Annotation
     * Recategorize zones that are horizontally or vertically aligned with zones of another category (within pixel tolerance).
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyRecategorizeByAlignment({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Recategorize by alignment rule
         */
        annotationRule: annotationrule_RecategorizeByAlignment,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/recategorize_by_alignment',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Remove Categories in Annotation
     * Remove categories in an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyRemoveCategories({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Remove categories rule
         */
        annotationRule: annotationrule_RemoveCategories,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/remove_categories',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Remove Overlap in Annotation
     * Remove overlapping annotations in an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyRemoveOverlap({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Remove overlap rule
         */
        annotationRule: annotationrule_RemoveOverlap,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/remove_overlap',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Rename Categories in Annotation
     * Rename categories in an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyRenameCategories({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Rename categories rule
         */
        annotationRule: annotationrule_RenameCategories,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/rename_categories',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Resolve Overlap with Priority in Annotation
     * Remove zones of SuppressedCategory that overlap DominantCategory by at least MinOverlap percent.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyResolveOverlapWithPriority({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Resolve overlap with priority rule
         */
        annotationRule: annotationrule_ResolveOverlapWithPriority,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/resolve_overlap_with_priority',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Apply Slice Pages Rule to Annotation
     * Apply a slice pages rule to an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplySlicePages({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Annotation slice pages rule
         */
        annotationRule: annotationrule_SlicePages,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/slice_pages',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Apply Stretch Rule to Annotation
     * Apply a stretch rule to an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyStretch({
        dataSetId,
        id,
        annotationRule,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Annotation stretch rule
         */
        annotationRule: annotationrule_Stretch,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/stretch',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationRule,
        });
    }
    /**
     * Apply Text Block Corrections to Annotation
     * Apply text block corrections to an annotation.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsApplyTextBlockCorrections({
        dataSetId,
        id,
        annotationTextBlockCorrections,
        action = 'overwrite',
        executionMode = 'sync',
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Text block corrections rule
         */
        annotationTextBlockCorrections: annotationrule_TextBlockCorrections,
        /**
         * Action to take when applying the rule
         */
        action?: 'overwrite' | 'create_new',
        /**
         * Execution mode for applying the rule
         */
        executionMode?: 'sync' | 'async',
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/apply/text_block_corrections',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'action': action,
                'execution_mode': executionMode,
            },
            body: annotationTextBlockCorrections,
        });
    }
    /**
     * Upload GPU farm annotation detection result
     * Uploads ALTO result ZIP produced by a GPU farm detection job.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static postDatasetsAnnotationsDetectionUpload({
        dataSetId,
        id,
        mode,
        file,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
        /**
         * Detection mode
         */
        mode: 'lines' | 'model_segment' | 'model_ocr',
        /**
         * ALTO result ZIP
         */
        file: Blob,
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations/{id}/detection_upload',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            formData: {
                'mode': mode,
                'file': file,
            },
        });
    }
}
