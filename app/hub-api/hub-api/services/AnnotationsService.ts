/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotation_Annotation } from '../models/annotation_Annotation';
import type { annotation_DuplicateRequest } from '../models/annotation_DuplicateRequest';
import type { annotation_ExpectedBlocks } from '../models/annotation_ExpectedBlocks';
import type { annotation_Index } from '../models/annotation_Index';
import type { annotation_Search } from '../models/annotation_Search';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AnnotationsService {
    /**
     * List Annotations
     * Get a list of annotations for a specific dataset.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static getDatasetsAnnotations({
        dataSetId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
    }): CancelablePromise<Array<annotation_Annotation>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations',
            path: {
                'dataSetId': dataSetId,
            },
        });
    }
    /**
     * Create Annotation
     * Create a new annotation for a specific dataset.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static postDatasetsAnnotations({
        dataSetId,
        annotation,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation to create
         */
        annotation: annotation_Annotation,
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations',
            path: {
                'dataSetId': dataSetId,
            },
            body: annotation,
        });
    }
    /**
     * Duplicate Annotation
     * Duplicate an existing annotation for a specific dataset.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static postDatasetsAnnotationsDuplicate({
        dataSetId,
        annotationDuplicateRequest,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation duplication details
         */
        annotationDuplicateRequest: annotation_DuplicateRequest,
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations/duplicate',
            path: {
                'dataSetId': dataSetId,
            },
            body: annotationDuplicateRequest,
        });
    }
    /**
     * Upload from URL
     * Upload annotations from a ZIP file located at a URL.
     * @returns annotation_Annotation Created
     * @throws ApiError
     */
    public static postDatasetsAnnotationsFromurl({
        dataSetId,
        format,
        url,
        name,
        description,
        segmented,
        ocred,
        groundTruth,
        originAnnotationId,
        ocrModelId,
        segmentModelId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation format
         */
        format: 'ALTO' | 'YOLO',
        /**
         * URL of the ZIP file to download
         */
        url: string,
        /**
         * Name of the annotation
         */
        name?: string,
        /**
         * Description of the annotation
         */
        description?: string,
        /**
         * Whether the annotations are segmented
         */
        segmented?: boolean,
        /**
         * Whether the annotations are OCRed
         */
        ocred?: boolean,
        /**
         * Whether the annotations are ground truth
         */
        groundTruth?: boolean,
        /**
         * Origin annotation ID to copy applied rules from
         */
        originAnnotationId?: string,
        /**
         * Model ID that was used for OCR processing, only relevant if annotations are OCRed
         */
        ocrModelId?: string,
        /**
         * Model ID that was used for segmentation, only relevant if annotations are segmented
         */
        segmentModelId?: string,
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations/fromurl',
            path: {
                'dataSetId': dataSetId,
            },
            query: {
                'format': format,
                'url': url,
                'name': name,
                'description': description,
                'segmented': segmented,
                'ocred': ocred,
                'ground_truth': groundTruth,
                'origin_annotation_id': originAnnotationId,
                'ocr_model_id': ocrModelId,
                'segment_model_id': segmentModelId,
            },
        });
    }
    /**
     * Upload ZIP File
     * Upload a ZIP file containing annotations.
     * @returns annotation_Annotation Created
     * @throws ApiError
     */
    public static postDatasetsAnnotationsFromzip({
        dataSetId,
        file,
        format,
        name,
        description,
        segmented,
        ocred,
        groundTruth,
        originAnnotationId,
        ocrModelId,
        segmentModelId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * ZIP file to upload
         */
        file: Blob,
        /**
         * Annotation format
         */
        format: 'ALTO' | 'YOLO',
        /**
         * Name of the annotation
         */
        name?: string,
        /**
         * Description of the annotation
         */
        description?: string,
        /**
         * Whether the annotations are segmented
         */
        segmented?: boolean,
        /**
         * Whether the annotations are OCRed
         */
        ocred?: boolean,
        /**
         * Whether the annotations are ground truth
         */
        groundTruth?: boolean,
        /**
         * Origin annotation ID to copy applied rules from
         */
        originAnnotationId?: string,
        /**
         * Model ID that was used for OCR processing, only relevant if annotations are OCRed
         */
        ocrModelId?: string,
        /**
         * Model ID that was used for segmentation, only relevant if annotations are segmented
         */
        segmentModelId?: string,
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations/fromzip',
            path: {
                'dataSetId': dataSetId,
            },
            query: {
                'format': format,
                'name': name,
                'description': description,
                'segmented': segmented,
                'ocred': ocred,
                'ground_truth': groundTruth,
                'origin_annotation_id': originAnnotationId,
                'ocr_model_id': ocrModelId,
                'segment_model_id': segmentModelId,
            },
            formData: {
                'file': file,
            },
        });
    }
    /**
     * Get Annotation
     * Get a specific annotation for a specific dataset.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static getDatasetsAnnotations1({
        dataSetId,
        id,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
        });
    }
    /**
     * Update Annotation
     * Update a specific annotation for a specific dataset.
     * @returns annotation_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotations({
        dataSetId,
        id,
        annotation,
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
         * Annotation to update
         */
        annotation: annotation_Annotation,
    }): CancelablePromise<annotation_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            body: annotation,
        });
    }
    /**
     * Delete Annotation
     * Delete a specific annotation for a specific dataset.
     * @returns string OK
     * @throws ApiError
     */
    public static deleteDatasetsAnnotations({
        dataSetId,
        id,
        deep,
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
         * Whether to perform a deep delete, which removes all associated files
         */
        deep?: boolean,
    }): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/datasets/{dataSetId}/annotations/{id}',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'deep': deep,
            },
        });
    }
    /**
     * Get Available Categories
     * Get the available categories for a specific annotation in a specific dataset.
     * @returns string OK
     * @throws ApiError
     */
    public static getDatasetsAnnotationsCategories({
        dataSetId,
        id,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
        /**
         * Annotation ID
         */
        id: string,
    }): CancelablePromise<Array<string>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/categories',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
        });
    }
    /**
     * Get Annotation Index
     * Get the index of a specific annotation for a specific dataset.
     * @returns annotation_Index OK
     * @throws ApiError
     */
    public static getDatasetsAnnotationsIndex({
        dataSetId,
        id,
        categories,
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
         * Categories for the index
         */
        categories?: string,
    }): CancelablePromise<annotation_Index> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/index',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'categories': categories,
            },
        });
    }
    /**
     * Create an annotation review based on expected blocks
     * Create an annotation review by providing expected blocks for comparison
     * @returns annotation_ExpectedBlocks OK
     * @throws ApiError
     */
    public static postDatasetsAnnotationsReview({
        dataSetId,
        id,
        review,
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
         * Expected blocks for review
         */
        review: annotation_ExpectedBlocks,
    }): CancelablePromise<annotation_ExpectedBlocks> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/datasets/{dataSetId}/annotations/{id}/review',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            body: review,
        });
    }
    /**
     * Search within an annotation's OCR data
     * Search for text patterns within specified categories of an annotation's OCR data
     * @returns annotation_Search OK
     * @throws ApiError
     */
    public static getDatasetsAnnotationsSearch({
        dataSetId,
        id,
        regex,
        category,
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
         * Regular expression pattern to search for
         */
        regex: string,
        /**
         * Categories to search within (can be specified multiple times)
         */
        category?: Array<string>,
    }): CancelablePromise<annotation_Search> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/search',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            query: {
                'category': category,
                'regex': regex,
            },
        });
    }
    /**
     * Get Annotation TEI
     * Get the TEI representation of a specific annotation for a specific dataset and page.
     * @returns string TEI XML content
     * @throws ApiError
     */
    public static getDatasetsAnnotationsTei({
        dataSetId,
        id,
        pageNumOrKey,
        feature,
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
         * Page Number or Key
         */
        pageNumOrKey: string,
        /**
         * Features to include in TEI data (can be specified multiple times)
         */
        feature?: Array<string>,
    }): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/tei/{pageNumOrKey}',
            path: {
                'dataSetId': dataSetId,
                'id': id,
                'pageNumOrKey': pageNumOrKey,
            },
            query: {
                'feature': feature,
            },
        });
    }
}
