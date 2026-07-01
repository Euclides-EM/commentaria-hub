/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';
export type annotationrule_LinesDetect = {
    applicable_stages?: Array<annotationrule_PipelineStage>;
    /**
     * IgnoreCategories specifies which categories to ignore when running line detection. For example, "GraphicZone", "DigitizationArtefactZone", ...
     * Example: ["CatchWord", "DigitizationArtefactZone", "DropCapitalZone", "GraphicZone-Decoration", "GraphicZone-Diagram", "NumberingZone", "QuireMarksZone", "RunningTitleZone"]
     */
    ignore_categories?: Array<string>;
    /**
     * IncludeCategories specifies which categories to run line detection on. For example, "MainZone".
     * Example: ["MainZone"]
     */
    include_categories?: Array<string>;
    type?: annotationrule_Type;
    use_gpu_farm?: boolean;
};
