import type {
  annotationrule_ExecutionMode,
  annotationrule_AddMargin,
  annotationrule_LinesDetect,
  annotationrule_LimitCategoryZones,
  annotationrule_PipelineStage,
  annotationrule_RecategorizeByAlignment,
  annotationrule_ReassignTextLinesByTolerance,
  annotationrule_RenameCategories,
  annotationrule_RemoveCategories,
  annotationrule_RemoveOverlap,
  annotationrule_ResolveOverlapWithPriority,
  annotationrule_ModelDetect,
  annotationrule_SlicePages,
  annotationrule_Stretch,
  annotationrule_TextBlockCorrections,
  annotation_Annotation,
  annotationrule_Type,
} from '@hub-api'

type BaseAnnotationRule = {
  type: annotationrule_Type
  applicable_stages?: annotationrule_PipelineStage[]
}

export type AnnotationRule = BaseAnnotationRule &
  (
    | annotationrule_ModelDetect
    | annotationrule_SlicePages
    | annotationrule_Stretch
    | annotationrule_AddMargin
    | annotationrule_LinesDetect
    | annotationrule_RemoveCategories
    | annotationrule_RenameCategories
    | annotationrule_RemoveOverlap
    | annotationrule_ResolveOverlapWithPriority
    | annotationrule_RecategorizeByAlignment
    | annotationrule_LimitCategoryZones
    | annotationrule_ReassignTextLinesByTolerance
    | annotationrule_TextBlockCorrections
  )

export type RuleRequestPayload = AnnotationRule & {
  execution_mode?: annotationrule_ExecutionMode
  name?: string
  description?: string
}

export type RuleOption = {
  value: annotationrule_Type
  label: string
}

export const isAnnotationRule = (value: unknown): value is AnnotationRule => {
  return (
    typeof value === 'object' &&
    value !== null &&
    'type' in value &&
    typeof value.type === 'string'
  )
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
  } else if (rule.type === 'model_detect') {
    const segmentRule = rule as annotationrule_ModelDetect
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
  annotation: annotation_Annotation,
): boolean => {
  const serializedRule = JSON.stringify(suggestedRule)
  return (
    !!annotation?.applied_rules &&
    annotation.applied_rules.some(
      (rule) => JSON.stringify(rule) === serializedRule,
    )
  )
}
