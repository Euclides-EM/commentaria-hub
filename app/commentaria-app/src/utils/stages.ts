import type { annotationrule_PipelineStage } from '../api'

const stageDisplayNames: Record<annotationrule_PipelineStage, string> = {
  raw: 'Raw',
  zone_segmentation: 'Zone Segmentation',
  text_line_segmentation: 'Text Line Segmentation',
  ocr: 'OCR',
}

export const getStageDisplayName = (
  stage: annotationrule_PipelineStage,
): string => {
  return stageDisplayNames[stage]
}
