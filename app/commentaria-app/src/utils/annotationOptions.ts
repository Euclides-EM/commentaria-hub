import type {
  annotation_Annotation,
  annotationrule_PipelineStage,
} from '@hub-api'

export const HIDDEN_FILTER = '__hidden__' as const
export type AnnotationFilterItem =
  | annotationrule_PipelineStage
  | typeof HIDDEN_FILTER

export const buildAnnotationOptions = (
  annotations: annotation_Annotation[] | undefined,
  selectedStages: AnnotationFilterItem[] | null,
  selectedAnnotationId: string,
) => {
  if (!annotations) {
    return []
  }
  const includeHidden = selectedStages?.includes(HIDDEN_FILTER) ?? false
  const options = annotations
    .filter((a) => {
      if (a.hidden && !includeHidden) {
        return false
      }
      return (
        selectedStages == null ||
        !a.pipeline_stage ||
        selectedStages.includes(a.pipeline_stage)
      )
    })
    .filter((a) => !!a.id)
    .map((a) => ({
      value: a.id as string,
      label: a.name || (a.id as string),
    }))
  if (selectedAnnotationId) {
    const selectedAnnotation = annotations.find(
      (a) => a.id === selectedAnnotationId,
    )
    if (
      selectedAnnotation?.id &&
      !options.some((option) => option.value === selectedAnnotation.id)
    ) {
      options.push({
        value: selectedAnnotation.id,
        label: selectedAnnotation.name || selectedAnnotation.id,
      })
    }
  }
  return options.sort((a, b) => a.label.localeCompare(b.label))
}
