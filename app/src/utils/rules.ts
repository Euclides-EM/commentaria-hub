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
  model_Annotation,
} from '../api'

export type AnnotationRule = { type: string } & (
  | annotationrule_Segment
  | annotationrule_SlicePages
  | annotationrule_Stretch
  | annotationrule_AddMargin
  | annotationrule_LinesDetect
  | annotationrule_RemoveCategories
  | annotationrule_RemoveOverlap
  | annotationrule_ReassignTextLinesByTolerance
  | annotationrule_TextBlockCorrections
)

export type RuleRequestPayload = AnnotationRule & {
  name?: string
  description?: string
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
  let serializedRule = JSON.stringify(suggestedRule)
  return (
    !!annotation?.applied_rules &&
    annotation.applied_rules.some(
      (rule) => JSON.stringify(rule) === serializedRule,
    )
  )
}
