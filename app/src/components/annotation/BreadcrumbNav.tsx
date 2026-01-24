import { useMemo } from 'react'
import Select from 'react-select'
import { useDatasetsQuery } from '../../queries/datasets'
import { useAnnotationsQuery } from '../../queries/annotations'
import { selectStyles } from '../../styles/selectStyles'
import { useAppState } from '../../context/useAppState'
import { usePipelineStages } from '../../queries/metadata.ts'
import useLocalStorageState from 'use-local-storage-state'
import type { annotationrule_PipelineStage } from '../../api'
import { MultiSelectDropdown } from '../core/MultiSelectDropdown.tsx'

import { getStageDisplayName } from '../../utils/stages.ts'

const Separator = () => <span className="bg-gray-600 w-[1px] h-fill mx-2" />

export function BreadcrumbNav() {
  const { state, setState } = useAppState()

  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(state.datasetId)
  const { data: stages } = usePipelineStages()
  const [selectedStages, setSelectedStages] = useLocalStorageState<
    annotationrule_PipelineStage[] | null
  >('annotationFilterStages', {
    defaultValue: null,
  })

  const filteredAnnotations = useMemo(() => {
    if (!annotations) {
      return []
    }
    return annotations.filter(
      (a) =>
        selectedStages == null ||
        !a.pipeline_stage ||
        selectedStages.includes(a.pipeline_stage),
    )
  }, [annotations, selectedStages])

  const datasetOptions = useMemo(() => {
    if (!datasets) return []
    return datasets
      .filter((d) => !!d.id)
      .map((d) => ({
        value: d.id as string,
        label: d.name || (d.id as string),
      }))
  }, [datasets])

  const annotationOptions = useMemo(() => {
    return filteredAnnotations
      .filter((a) => !!a.id)
      .map((a) => ({
        value: a.id as string,
        label: a.name || (a.id as string),
      }))
  }, [filteredAnnotations])

  const selectedDataset =
    datasetOptions.find((d) => d.value === state.datasetId) || null
  const selectedAnnotation =
    annotationOptions.find((a) => a.value === state.annotationId) || null

  const handleDatasetChange = (value: string) => {
    setState({ datasetId: value, annotationId: '' })
  }

  const handleAnnotationChange = (value: string) => {
    setState({ annotationId: value })
  }

  const highlightDataset = !state.datasetId
  const highlightAnnotation = !!state.datasetId && !state.annotationId

  return (
    <div className="flex items-center text-sm gap-2 flex-wrap">
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
          isClearable
        />
      </div>

      {state.datasetId && (
        <>
          <Separator />

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
              isClearable
              isDisabled={!state.datasetId}
            />
          </div>

          {stages && (
            <MultiSelectDropdown
              allItems={stages}
              selectedItems={selectedStages}
              setSelectedItems={setSelectedStages}
              itemsLabel="stages"
              getItemLabel={(stage) => getStageDisplayName(stage)}
            />
          )}
        </>
      )}
    </div>
  )
}
