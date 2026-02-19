/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { annotationrule_PipelineStage } from './annotationrule_PipelineStage';
import type { annotationrule_Type } from './annotationrule_Type';

export type annotationrule_RecategorizeByAlignmentRelativeTo = {
  category: string;
  alignment: 'horizontal' | 'vertical';
};

export type annotationrule_RecategorizeByAlignment = {
  applicable_stages?: Array<annotationrule_PipelineStage>;
  original_category?: string;
  target_category?: string;
  relative_to?: annotationrule_RecategorizeByAlignmentRelativeTo;
  tolerance_px?: number;
  type?: annotationrule_Type;
};
