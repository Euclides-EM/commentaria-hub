import type { annotationrule_PipelineStage } from '../api'
import { getStageDisplayName } from '../utils/rules.ts'
import { MultiSelectDropdown } from './MultiSelectDropdown'

interface AnnotationFilterDropdownProps {
  allStages: annotationrule_PipelineStage[]
  selectedStages: annotationrule_PipelineStage[]
  onToggleStage: (filter: annotationrule_PipelineStage) => void
}

export function AnnotationFilterDropdown({
  allStages,
  selectedStages,
  onToggleStage,
}: AnnotationFilterDropdownProps) {
  const getLabel = () => {
    if (
      selectedStages.length === 0 ||
      selectedStages.length === allStages.length
    )
      return 'All stages'
    if (selectedStages.length === 1) {
      return getStageDisplayName(selectedStages[0])
    }
    return `${selectedStages.length} stages`
  }

  return (
    <MultiSelectDropdown
      allItems={allStages}
      selectedItems={selectedStages}
      onToggleItem={onToggleStage}
      getButtonLabel={() => getLabel()}
      getItemLabel={(stage) => getStageDisplayName(stage)}
    />
  )
}
