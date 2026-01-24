import { useMemo } from 'react'
import Select from 'react-select'
import { useDatasetsQuery } from '../queries/datasets'
import { useAnnotationsQuery } from '../queries/annotations'
import { selectStyles } from '../styles/selectStyles'
import { useAppState } from '../context/AppStateContext'
import { AnnotationFilterDropdown } from './AnnotationFilterDropdown'
import { usePipelineStages } from '../queries/metadata.ts'
import useLocalStorageState from 'use-local-storage-state'
import type { annotationrule_PipelineStage } from '../api'

const Separator = () => <span className="bg-gray-600 w-[1px] h-fill mx-2" />

export function BreadcrumbNav() {
  const { state, setState } = useAppState()

  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(state.datasetId)
  const { data: stages } = usePipelineStages()
  const [selectedStages, setSelectedStages] = useLocalStorageState<
    annotationrule_PipelineStage[]
  >('annotationFilterStages', {
    defaultValue: [],
  })

  const filteredAnnotations = useMemo(() => {
    if (!annotations) {
      return []
    }
    return annotations.filter(
      (a) =>
        selectedStages.length === 0 ||
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

  return (
    <div className="flex items-center text-sm gap-2 flex-wrap">
      <Separator />

      <div className="text-shadow-gray-800 font-semibold">Dataset</div>
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

          <div className="text-shadow-gray-800 font-semibold">Annotation</div>
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
            <AnnotationFilterDropdown
              allStages={stages}
              selectedStages={selectedStages}
              onToggleStage={(stage) => {
                if (
                  selectedStages.length === 0 ||
                  selectedStages.includes(stage)
                ) {
                  setSelectedStages(
                    (selectedStages.length === 0
                      ? stages
                      : selectedStages
                    ).filter((s) => s !== stage),
                  )
                } else {
                  const nextStages = [...selectedStages, stage]
                  setSelectedStages(
                    nextStages.length === stages.length ? [] : nextStages,
                  )
                }
              }}
            />
          )}
        </>
      )}
    </div>
  )
}
