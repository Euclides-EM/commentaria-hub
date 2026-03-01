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
      .sort((a, b) => a.label.localeCompare(b.label))
  }, [datasets])

  const annotationOptions = useMemo(() => {
    return filteredAnnotations
      .filter((a) => !!a.id)
      .map((a) => ({
        value: a.id as string,
        label: a.name || (a.id as string),
      }))
      .sort((a, b) => a.label.localeCompare(b.label))
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

      {state.datasetId && (
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

          {stages && (
            <MultiSelectDropdown
              allItems={stages}
              selectedItems={selectedStages}
              setSelectedItems={setSelectedStages}
              itemsLabel="stages"
              getItemLabel={(stage) => getStageDisplayName(stage)}
            />
          )}
        </div>
      )}
    </div>
  )
}
