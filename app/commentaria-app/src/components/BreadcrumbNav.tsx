import { useMemo } from 'react'
import Select from 'react-select'
import { useDatasetsQuery } from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import { selectStyles } from '../styles/selectStyles.ts'
import { useAppState } from '../context/useAppState.ts'
import { usePipelineStages } from '../queries/metadata.ts'
import useLocalStorageState from 'use-local-storage-state'
import type { annotationrule_PipelineStage } from '@hub-api'
import { MultiSelectDropdown } from './core/MultiSelectDropdown.tsx'

import { getStageDisplayName } from '../utils/stages.ts'
import { Button } from './core/Button.tsx'

const Separator = () => <span className="bg-gray-600 w-px h-fill mx-2" />
const HIDDEN_FILTER = '__hidden__' as const
type AnnotationFilterItem = annotationrule_PipelineStage | typeof HIDDEN_FILTER

export function BreadcrumbNav() {
  const { state, setState } = useAppState()

  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(state.datasetId)
  const { data: stages } = usePipelineStages()
  const [selectedStages, setSelectedStages] = useLocalStorageState<
    AnnotationFilterItem[] | null
  >('annotationFilterStages', {
    defaultValue: null,
  })

  const stageFilterItems = useMemo<AnnotationFilterItem[]>(
    () => [...(stages || []), HIDDEN_FILTER],
    [stages],
  )
  const effectiveSelectedStages = useMemo<AnnotationFilterItem[] | null>(() => {
    if (selectedStages != null) {
      return selectedStages
    }
    return stages || null
  }, [selectedStages, stages])
  const filteredAnnotations = useMemo(() => {
    if (!annotations) {
      return []
    }
    const includeHidden =
      effectiveSelectedStages?.includes(HIDDEN_FILTER) ?? false
    return annotations.filter((a) => {
      if (a.hidden && !includeHidden) {
        return false
      }
      return (
        effectiveSelectedStages == null ||
        !a.pipeline_stage ||
        effectiveSelectedStages.includes(a.pipeline_stage)
      )
    })
  }, [annotations, effectiveSelectedStages])

  const datasetOptions = useMemo(() => {
    if (!datasets) return []
    return datasets
      .filter((d) => !!d.id)
      .map((d) => ({
        value: d.id as string,
        label: d.name || (d.id as string),
      }))
      .sort((a, b) => a.label.localeCompare(b.label))
  }, [datasets])

  const annotationOptions = useMemo(() => {
    const options = filteredAnnotations
      .filter((a) => !!a.id)
      .map((a) => ({
        value: a.id as string,
        label: a.name || (a.id as string),
      }))
    if (state.annotationId && annotations) {
      const selectedAnnotation = annotations.find(
        (a) => a.id === state.annotationId,
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
  }, [annotations, filteredAnnotations, state.annotationId])

  const selectedDataset =
    datasetOptions.find((d) => d.value === state.datasetId) || null
  const selectedAnnotation =
    annotationOptions.find((a) => a.value === state.annotationId) || null
  const showAnnotationSelect =
    !!state.datasetId && (annotationsLoading || (annotations?.length ?? 0) > 0)

  const handleDatasetChange = (value: string) => {
    setState({ datasetId: value, annotationId: '' })
  }

  const handleAnnotationChange = (value: string) => {
    setState({ annotationId: value })
  }

  const highlightDataset = !state.viewMode && !state.datasetId
  const highlightAnnotation =
    !state.viewMode && !!state.datasetId && !state.annotationId

  return (
    <div className="flex items-center text-sm gap-x-2 gap-y-4 flex-wrap">
      <Separator />
      <Button
        variant="primary"
        className={`h-12 w-20 px-2 ${state.viewMode === 'models' && '!bg-teal-100 hover:!bg-white'}`}
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
        className={`h-12 w-24 px-2 ${state.viewMode === 'annotations' && '!bg-teal-100 hover:!bg-white'}`}
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
        className={`h-12 w-20 px-2 ${state.viewMode === 'jobs' && '!bg-teal-100 hover:!bg-white'}`}
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

          <MultiSelectDropdown
            allItems={stageFilterItems}
            selectedItems={effectiveSelectedStages}
            setSelectedItems={setSelectedStages}
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

              if (
                allStageCount > 0 &&
                selectedStages.length === allStageCount
              ) {
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
        </div>
      )}
    </div>
  )
}
