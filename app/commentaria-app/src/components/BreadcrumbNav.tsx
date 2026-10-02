import { useMemo } from 'react'
import Select from 'react-select'
import { useDatasetsQuery } from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import { selectStyles } from '../styles/selectStyles.ts'
import { useAppState } from '../context/useAppState.ts'
import { usePipelineStages } from '../queries/metadata.ts'
import useLocalStorageState from 'use-local-storage-state'
import type {
  annotation_Annotation,
  annotationrule_PipelineStage,
} from '@hub-api'
import { MultiSelectDropdown } from './core/MultiSelectDropdown.tsx'

import { getStageDisplayName } from '../utils/stages.ts'
import { Button } from './core/Button.tsx'
import {
  DATASET_COMPLETENESS_ITEMS,
  getDatasetCompleteness,
  type DatasetCompleteness,
} from '../utils/datasets.ts'

const Separator = () => <span className="self-stretch bg-gray-600 w-px mx-2" />
const HIDDEN_FILTER = '__hidden__' as const
type AnnotationFilterItem = annotationrule_PipelineStage | typeof HIDDEN_FILTER

const buildAnnotationOptions = (
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

interface StageFilterDropdownProps {
  stages: annotationrule_PipelineStage[] | undefined
  selectedItems: AnnotationFilterItem[] | null
  setSelectedItems: (items: AnnotationFilterItem[] | null) => void
}

const StageFilterDropdown = ({
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

export function BreadcrumbNav() {
  const { state, setState } = useAppState()

  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(state.datasetId)
  const isCompareMode =
    !state.viewMode && !!state.annotationId && state.annotationTab === 'compare'
  const { data: otherAnnotations, isLoading: otherAnnotationsLoading } =
    useAnnotationsQuery(isCompareMode ? state.otherDatasetId : '')
  const { data: stages } = usePipelineStages()
  const [selectedStages, setSelectedStages] = useLocalStorageState<
    AnnotationFilterItem[] | null
  >('annotationFilterStages', {
    defaultValue: null,
  })
  const [selectedOtherStages, setSelectedOtherStages] = useLocalStorageState<
    AnnotationFilterItem[] | null
  >('otherAnnotationFilterStages', {
    defaultValue: null,
  })
  const [selectedDatasetCompleteness, setSelectedDatasetCompleteness] =
    useLocalStorageState<DatasetCompleteness[] | null>(
      'datasetCompletenessFilter',
      { defaultValue: null },
    )

  const effectiveSelectedStages = useMemo<AnnotationFilterItem[] | null>(() => {
    if (selectedStages != null) {
      return selectedStages
    }
    return stages || null
  }, [selectedStages, stages])
  const effectiveSelectedOtherStages = useMemo<
    AnnotationFilterItem[] | null
  >(() => {
    if (selectedOtherStages != null) {
      return selectedOtherStages
    }
    return stages || null
  }, [selectedOtherStages, stages])
  const datasetOptions = useMemo(() => {
    if (!datasets) return []
    const selectedCompleteness =
      selectedDatasetCompleteness ?? DATASET_COMPLETENESS_ITEMS
    return datasets
      .filter((d) => {
        if (!d.id) return false
        const completeness = getDatasetCompleteness(d.pages)
        return selectedCompleteness.includes(completeness)
      })
      .map((d) => ({
        value: d.id as string,
        label: d.name || (d.id as string),
      }))
      .sort((a, b) => a.label.localeCompare(b.label))
  }, [datasets, selectedDatasetCompleteness])

  const annotationOptions = useMemo(
    () =>
      buildAnnotationOptions(
        annotations,
        effectiveSelectedStages,
        state.annotationId,
      ),
    [annotations, effectiveSelectedStages, state.annotationId],
  )

  const otherAnnotationOptions = useMemo(
    () =>
      buildAnnotationOptions(
        otherAnnotations,
        effectiveSelectedOtherStages,
        state.otherAnnotationId,
      ),
    [otherAnnotations, effectiveSelectedOtherStages, state.otherAnnotationId],
  )

  const selectedDataset =
    datasetOptions.find((d) => d.value === state.datasetId) || null
  const selectedAnnotation =
    annotationOptions.find((a) => a.value === state.annotationId) || null
  const selectedOtherDataset =
    datasetOptions.find((d) => d.value === state.otherDatasetId) || null
  const selectedOtherAnnotation =
    otherAnnotationOptions.find((a) => a.value === state.otherAnnotationId) ||
    null
  const showAnnotationSelect =
    !!state.datasetId && (annotationsLoading || (annotations?.length ?? 0) > 0)

  const handleDatasetChange = (value: string) => {
    setState({ datasetId: value, annotationId: '' })
  }

  const handleDatasetCompletenessChange = (
    items: DatasetCompleteness[] | null,
  ) => {
    setSelectedDatasetCompleteness(items)

    if (!state.datasetId || !datasets) return
    const selectedDataset = datasets.find(
      (dataset) => dataset.id === state.datasetId,
    )
    if (!selectedDataset) return

    const selectedCompleteness = getDatasetCompleteness(selectedDataset.pages)
    const effectiveItems = items ?? DATASET_COMPLETENESS_ITEMS
    if (!effectiveItems.includes(selectedCompleteness)) {
      handleDatasetChange('')
    }
  }

  const handleAnnotationChange = (value: string) => {
    setState({ annotationId: value })
  }

  const handleOtherDatasetChange = (value: string) => {
    setState({ otherDatasetId: value, otherAnnotationId: '' })
  }

  const handleOtherAnnotationChange = (value: string) => {
    setState({ otherAnnotationId: value })
  }

  const highlightDataset = !state.viewMode && !state.datasetId
  const highlightAnnotation =
    !state.viewMode && !!state.datasetId && !state.annotationId
  const highlightOtherDataset = isCompareMode && !state.otherDatasetId
  const highlightOtherAnnotation =
    isCompareMode && !!state.otherDatasetId && !state.otherAnnotationId

  return (
    <div className="flex items-center text-sm gap-x-2 gap-y-4 flex-wrap">
      <Separator />
      <Button
        variant="primary"
        className={`h-8 w-16 px-2 text-xs ${state.viewMode === 'models' && '!bg-teal-100 hover:!bg-white'}`}
        onClick={() =>
          setState({
            viewMode: state.viewMode === 'models' ? null : 'models',
          })
        }
      >
        Models
      </Button>
      <Button
        variant="primary"
        className={`h-8 w-20 px-2 text-xs px-2 ${state.viewMode === 'annotations' && '!bg-teal-100 hover:!bg-white'}`}
        onClick={() =>
          setState({
            viewMode: state.viewMode === 'annotations' ? null : 'annotations',
          })
        }
      >
        Annotations
      </Button>
      <Button
        variant="primary"
        className={`h-8 w-16 px-2 text-xs px-2 ${state.viewMode === 'features' && '!bg-teal-100 hover:!bg-white'}`}
        onClick={() =>
          setState({
            viewMode: state.viewMode === 'features' ? null : 'features',
          })
        }
      >
        Features
      </Button>
      <Button
        variant="primary"
        className={`h-8 w-16 px-2 text-xs px-2 ${state.viewMode === 'jobs' && '!bg-teal-100 hover:!bg-white'}`}
        onClick={() =>
          setState({
            viewMode: state.viewMode === 'jobs' ? null : 'jobs',
          })
        }
      >
        Jobs
      </Button>

      <Separator />

      <div
        className={`text-shadow-gray-800 font-semibold px-1 rounded ${highlightDataset ? 'label-glow text-teal-800' : ''}`}
      >
        Dataset
      </div>
      <div style={{ minWidth: '200px' }}>
        <Select
          value={selectedDataset}
          onChange={(option: { value: string; label: string } | null) =>
            handleDatasetChange(option?.value || '')
          }
          options={datasetOptions}
          placeholder="Select dataset..."
          isLoading={datasetsLoading}
          styles={selectStyles<{ value: string; label: string }>()}
          menuPortalTarget={document.body}
          menuPosition="fixed"
          isClearable
        />
      </div>

      <MultiSelectDropdown<DatasetCompleteness>
        allItems={DATASET_COMPLETENESS_ITEMS}
        selectedItems={selectedDatasetCompleteness}
        setSelectedItems={handleDatasetCompletenessChange}
        itemsLabel="coverage types"
        getItemLabel={(item) => (item === 'full' ? 'Full' : 'Partial')}
        getPickerLabel={({ selectedItems }) => {
          if (
            selectedItems == null ||
            selectedItems.length === DATASET_COMPLETENESS_ITEMS.length
          ) {
            return 'Full & partial'
          }
          if (selectedItems.length === 0) return 'None'
          return selectedItems[0] === 'full' ? 'Full only' : 'Partial only'
        }}
      />

      {showAnnotationSelect && (
        <div className="flex items-center gap-2 flex-nowrap shrink-0">
          <div className="h-3 w-3 rotate-[-45deg] border-b border-r border-slate-600" />

          <div
            className={`text-shadow-gray-800 font-semibold px-1 rounded ${highlightAnnotation ? 'label-glow text-teal-800' : ''}`}
          >
            Annotation
          </div>
          <div style={{ minWidth: '200px' }}>
            <Select
              value={selectedAnnotation}
              onChange={(option: { value: string; label: string } | null) =>
                handleAnnotationChange(option?.value || '')
              }
              options={annotationOptions}
              placeholder="Select annotation..."
              isLoading={annotationsLoading}
              styles={selectStyles<{ value: string; label: string }>()}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
              isDisabled={!state.datasetId}
            />
          </div>

          <StageFilterDropdown
            stages={stages}
            selectedItems={effectiveSelectedStages}
            setSelectedItems={setSelectedStages}
          />
        </div>
      )}

      {isCompareMode && (
        <>
          <Separator />
          <div
            className={`text-shadow-gray-800 font-semibold px-1 rounded ${highlightOtherDataset ? 'label-glow text-teal-800' : ''}`}
          >
            Compare to dataset
          </div>
          <div style={{ minWidth: '200px' }}>
            <Select
              value={selectedOtherDataset}
              onChange={(option: { value: string; label: string } | null) =>
                handleOtherDatasetChange(option?.value || '')
              }
              options={datasetOptions}
              placeholder="Select dataset..."
              isLoading={datasetsLoading}
              styles={selectStyles<{ value: string; label: string }>()}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          </div>
          {state.otherDatasetId && (
            <div className="flex items-center gap-2 flex-nowrap shrink-0">
              <div className="h-3 w-3 rotate-[-45deg] border-b border-r border-slate-600" />
              <div
                className={`text-shadow-gray-800 font-semibold px-1 rounded ${highlightOtherAnnotation ? 'label-glow text-teal-800' : ''}`}
              >
                Annotation
              </div>
              <div style={{ minWidth: '200px' }}>
                <Select
                  value={selectedOtherAnnotation}
                  onChange={(option: { value: string; label: string } | null) =>
                    handleOtherAnnotationChange(option?.value || '')
                  }
                  options={otherAnnotationOptions}
                  placeholder="Select annotation..."
                  isLoading={otherAnnotationsLoading}
                  styles={selectStyles<{ value: string; label: string }>()}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                  isClearable
                />
              </div>
              <StageFilterDropdown
                stages={stages}
                selectedItems={effectiveSelectedOtherStages}
                setSelectedItems={setSelectedOtherStages}
              />
            </div>
          )}
        </>
      )}
    </div>
  )
}
