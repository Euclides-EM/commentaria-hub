/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { model_Annotation } from '../models/model_Annotation';
import type { model_AnnotationDuplicateRequest } from '../models/model_AnnotationDuplicateRequest';
import type { model_AnnotationExpectedBlocks } from '../models/model_AnnotationExpectedBlocks';
import type { model_AnnotationIndex } from '../models/model_AnnotationIndex';
import type { model_AnnotationSearch } from '../models/model_AnnotationSearch';
import type { model_AnnotationUploadEscriptorium } from '../models/model_AnnotationUploadEscriptorium';
import type { model_AnnotationUploadRoboflow } from '../models/model_AnnotationUploadRoboflow';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AnnotationsService {
    /**
     * List Annotations
     * Get a list of annotations for a specific dataset.
     * @returns model_Annotation OK
     * @throws ApiError
     */
    public static getDatasetsAnnotations({
        dataSetId,
    }: {
        /**
         * Dataset ID
         */
        dataSetId: string,
    }): CancelablePromise<Array<model_Annotation>> {
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
     * @returns model_Annotation OK
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
        annotation: model_Annotation,
    }): CancelablePromise<model_Annotation> {
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
     * @returns model_Annotation OK
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
        annotationDuplicateRequest: model_AnnotationDuplicateRequest,
    }): CancelablePromise<model_Annotation> {
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
     * @returns model_Annotation Created
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
        segmented?: 'true' | 'false',
        /**
         * Whether the annotations are OCRed
         */
        ocred?: 'true' | 'false',
        /**
         * Whether the annotations are ground truth
         */
        groundTruth?: 'true' | 'false',
        /**
         * Origin annotation ID to copy applied rules from
         */
        originAnnotationId?: string,
    }): CancelablePromise<model_Annotation> {
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
            },
        });
    }
    /**
     * Upload ZIP File
     * Upload a ZIP file containing annotations.
     * @returns model_Annotation Created
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
        segmented?: 'true' | 'false',
        /**
         * Whether the annotations are OCRed
         */
        ocred?: 'true' | 'false',
        /**
         * Whether the annotations are ground truth
         */
        groundTruth?: 'true' | 'false',
        /**
         * Origin annotation ID to copy applied rules from
         */
        originAnnotationId?: string,
    }): CancelablePromise<model_Annotation> {
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
            },
            formData: {
                'file': file,
            },
        });
    }
    /**
     * Get Annotation
     * Get a specific annotation for a specific dataset.
     * @returns model_Annotation OK
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
    }): CancelablePromise<model_Annotation> {
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
     * @returns model_Annotation OK
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
        annotation: model_Annotation,
    }): CancelablePromise<model_Annotation> {
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
        deep?: 'true' | 'false',
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
     * Download annotation assets
     * Generate and download assets for a specific annotation within a dataset
     * @returns binary ZIP file containing the annotation assets
     * @throws ApiError
     */
    public static getDatasetsAnnotationsDownloadAssets({
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
    }): CancelablePromise<Blob> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/download_assets',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
        });
    }
    /**
     * Get Annotation Index
     * Get the index of a specific annotation for a specific dataset.
     * @returns model_AnnotationIndex OK
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
        categories: string,
    }): CancelablePromise<model_AnnotationIndex> {
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
     * @returns model_AnnotationExpectedBlocks OK
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
        review: model_AnnotationExpectedBlocks,
    }): CancelablePromise<model_AnnotationExpectedBlocks> {
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
     * @returns model_AnnotationSearch OK
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
    }): CancelablePromise<model_AnnotationSearch> {
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
        pageNum,
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
         * Page Number
         */
        pageNum: string,
    }): CancelablePromise<string> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/datasets/{dataSetId}/annotations/{id}/tei/{pageNum}',
            path: {
                'dataSetId': dataSetId,
                'id': id,
                'pageNum': pageNum,
            },
        });
    }
    /**
     * Upload Annotation to Escriptorium
     * Upload an annotation to Escriptorium for a specific dataset.
     * @returns model_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsUploadEscriptorium({
        dataSetId,
        id,
        annotationEscriptoriumUpload,
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
         * Annotation Escriptorium upload details
         */
        annotationEscriptoriumUpload: model_AnnotationUploadEscriptorium,
    }): CancelablePromise<model_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/upload/escriptorium',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            body: annotationEscriptoriumUpload,
        });
    }
    /**
     * Upload Annotation to Roboflow
     * Upload an annotation to Roboflow for a specific dataset.
     * @returns model_Annotation OK
     * @throws ApiError
     */
    public static putDatasetsAnnotationsUploadRoboflow({
        dataSetId,
        id,
        annotationRoboflowUpload,
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
         * Annotation Roboflow upload details
         */
        annotationRoboflowUpload: model_AnnotationUploadRoboflow,
    }): CancelablePromise<model_Annotation> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/datasets/{dataSetId}/annotations/{id}/upload/roboflow',
            path: {
                'dataSetId': dataSetId,
                'id': id,
            },
            body: annotationRoboflowUpload,
        });
    }
}
