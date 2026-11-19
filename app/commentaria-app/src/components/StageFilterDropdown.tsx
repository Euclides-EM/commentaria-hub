import { useMemo } from 'react'
import type { annotationrule_PipelineStage } from '@hub-api'
import { MultiSelectDropdown } from './core/MultiSelectDropdown.tsx'
import { getStageDisplayName } from '../utils/stages.ts'
import {
  HIDDEN_FILTER,
  type AnnotationFilterItem,
} from '../utils/annotationOptions.ts'

interface StageFilterDropdownProps {
  stages: annotationrule_PipelineStage[] | undefined
  selectedItems: AnnotationFilterItem[] | null
  setSelectedItems: (items: AnnotationFilterItem[] | null) => void
}

export const StageFilterDropdown = ({
  stages,
  selectedItems,
  setSelectedItems,
}: StageFilterDropdownProps) => {
  const stageFilterItems = useMemo<AnnotationFilterItem[]>(
    () => [...(stages || []), HIDDEN_FILTER],
    [stages],
  )

  return (
    <MultiSelectDropdown
      allItems={stageFilterItems}
      selectedItems={selectedItems}
      setSelectedItems={setSelectedItems}
      itemsLabel="stages"
      bulkActionItems={stages || []}
      bulkActionLabel="stages"
      showSeparatorBeforeItem={(item) => item === HIDDEN_FILTER}
      getItemLabel={(item) =>
        item === HIDDEN_FILTER ? 'Hidden' : getStageDisplayName(item)
      }
      getPickerLabel={({ selectedItems }) => {
        const selected = selectedItems ?? stageFilterItems
        const allStageCount = stages?.length ?? 0
        const selectedStages = (stages || []).filter((stage) =>
          selected.includes(stage),
        )
        const isHiddenSelected = selected.includes(HIDDEN_FILTER)

        if (allStageCount > 0 && selectedStages.length === allStageCount) {
          return 'All stages'
        }
        if (selectedStages.length === 0) {
          return isHiddenSelected ? 'Hidden' : 'None'
        }
        if (selectedStages.length === 1) {
          return getStageDisplayName(selectedStages[0])
        }
        return `${selectedStages.length} stages`
      }}
    />
  )
}
