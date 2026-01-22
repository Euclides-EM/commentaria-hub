import { useMemo } from 'react'
import Select from 'react-select'
import { useDatasetsQuery } from '../queries/datasets'
import { useAnnotationsQuery } from '../queries/annotations'
import { selectStyles } from '../styles/selectStyles'
import { useAppState } from '../context/AppStateContext'
import {
  type AnnotationFilter,
  AnnotationFilterDropdown,
} from './AnnotationFilterDropdown'

const Separator = () => <span className="bg-gray-600 w-[1px] h-fill mx-2" />

export function BreadcrumbNav() {
  const { state, setState, toggleFilter } = useAppState()
  const annotationFilters = state.annotationFilters as AnnotationFilter[]

  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(state.datasetId, annotationFilters)

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
    if (!annotations) return []
    return annotations
      .filter((a) => !!a.id)
      .map((a) => ({
        value: a.id as string,
        label: a.name || (a.id as string),
      }))
  }, [annotations])

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

          <AnnotationFilterDropdown
            filters={annotationFilters}
            onToggleFilter={toggleFilter}
          />
        </>
      )}
    </div>
  )
}
