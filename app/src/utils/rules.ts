import type {
  annotationrule_AddMargin,
  annotationrule_LinesDetect,
  annotationrule_ReassignTextLinesByTolerance,
  annotationrule_RemoveCategories,
  annotationrule_RemoveOverlap,
  annotationrule_Segment,
  annotationrule_SlicePages,
  annotationrule_Stretch,
  annotationrule_TextBlockCorrections,
  annotationrule_PipelineStage, // Added this import
  model_Annotation,
} from '../api'

export type AnnotationRule =
  | annotationrule_Segment
  | annotationrule_SlicePages
  | annotationrule_Stretch
  | annotationrule_AddMargin
  | annotationrule_LinesDetect
  | annotationrule_RemoveCategories
  | annotationrule_RemoveOverlap
  | annotationrule_ReassignTextLinesByTolerance
  | (annotationrule_TextBlockCorrections & { type: string })

const stageDisplayNames: Record<annotationrule_PipelineStage, string> = {
  raw: 'Raw',
  zone_segmentation: 'Zone Segmentation',
  text_line_segmentation: 'Text Line Segmentation',
  ocr: 'OCR',
}

export const getStageDisplayName = (stage: annotationrule_PipelineStage): string => {
  return stageDisplayNames[stage]
}

export const getRuleDisplayName = (rule: AnnotationRule): string => {
  if (!rule.type) return 'Unknown Rule'

  const typeDisplay = rule.type
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')

  const details: string[] = []

  if (rule.type === 'slice_pages') {
    const sliceRule = rule as annotationrule_SlicePages
    if (sliceRule.pages) {
      details.push(`Pages: ${sliceRule.pages}`)
    } else if (sliceRule.random_pages) {
      details.push(`Random: ${sliceRule.random_pages} pages`)
    }
  } else if (rule.type === 'segment') {
    const segmentRule = rule as annotationrule_Segment
    if (segmentRule.model) {
      details.push(`Model: ${segmentRule.model}`)
    }
  }

  return details.length > 0
    ? `${typeDisplay} (${details.join(', ')})`
    : typeDisplay
}

export const isRuleApplied = (
  suggestedRule: AnnotationRule,
  annotation: model_Annotation,
): boolean => {
  if (!annotation?.applied_rules || annotation.applied_rules.length === 0) {
    return false
  }

  return annotation.applied_rules.some((appliedRule: unknown) => {
    const typedAppliedRule = appliedRule as AnnotationRule
    if (suggestedRule.type === typedAppliedRule.type) {
      switch (suggestedRule.type) {
        case 'slice_pages':
          const sliceSuggested = suggestedRule as annotationrule_SlicePages
          const sliceApplied = typedAppliedRule as annotationrule_SlicePages
          return (
            sliceSuggested.pages === sliceApplied.pages ||
            sliceSuggested.random_pages === sliceApplied.random_pages
          )
        case 'segment':
          const segmentSuggested = suggestedRule as annotationrule_Segment
          const segmentApplied = typedAppliedRule as annotationrule_Segment
          return segmentSuggested.model === segmentApplied.model
        case 'stretch':
        case 'add_margin':
        case 'lines_detect':
        case 'remove_categories':
        case 'remove_overlap':
        case 'reassign_text_lines_by_tolerance':
        case 'text_blocks_corrections':
          const suggestedKeys = Object.keys(suggestedRule).filter(
            (k) => k !== 'type',
          )
          const appliedKeys = Object.keys(typedAppliedRule).filter(
            (k) => k !== 'type',
          )

          if (suggestedKeys.length === 0 && appliedKeys.length === 0) {
            return true
          }

          return suggestedKeys.every(
            (key) =>
              (suggestedRule as Record<string, unknown>)[key] ===
              (typedAppliedRule as Record<string, unknown>)[key],
          )
        default:
          return (
            JSON.stringify(suggestedRule) === JSON.stringify(typedAppliedRule)
          )
      }
    }
    return false
  })
}
