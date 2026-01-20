import { useMemo } from 'react'
import Select from 'react-select'
import { useDatasetsQuery } from '../queries/datasets'
import { useAnnotationsQuery } from '../queries/annotations'
import { selectStyles } from '../styles/selectStyles'
import { useAppState } from '../context/AppStateContext'
import {
  AnnotationFilterDropdown,
  type AnnotationFilter,
} from './AnnotationFilterDropdown'

export function BreadcrumbNav() {
  const { state, setState, toggleFilter } = useAppState()
  const annotationFilters = state.annotationFilters as AnnotationFilter[]

  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(state.dataset, annotationFilters)

  const datasetOptions = useMemo(() => {
    if (!datasets) return []
    return datasets.map((d) => ({
      value: d.id,
      label: d.name || d.id,
    }))
  }, [datasets])

  const annotationOptions = useMemo(() => {
    if (!annotations) return []
    return annotations.map((a) => ({
      value: a.id,
      label: a.name || a.id,
    }))
  }, [annotations])

  const selectedDataset =
    datasetOptions.find((d) => d.value === state.dataset) || null
  const selectedAnnotation =
    annotationOptions.find((a) => a.value === state.annotation) || null

  const handleDatasetChange = (value: string) => {
    setState({ dataset: value, annotation: '' })
  }

  const handleAnnotationChange = (value: string) => {
    setState({ annotation: value })
  }

  return (
    <div className="flex items-center text-sm">
      <span className="text-gray-400">/</span>

      <div className="ml-2" style={{ minWidth: '200px' }}>
        <Select
          value={selectedDataset}
          onChange={(option) => handleDatasetChange(option?.value || '')}
          options={datasetOptions}
          placeholder="Select dataset..."
          isLoading={datasetsLoading}
          styles={selectStyles}
          isClearable
        />
      </div>

      {state.dataset && (
        <>
          <span className="text-gray-400 mx-2">/</span>

          <AnnotationFilterDropdown
            filters={annotationFilters}
            onToggleFilter={toggleFilter}
          />

          <span className="text-gray-400 mx-2">/</span>

          <div style={{ minWidth: '200px' }}>
            <Select
              value={selectedAnnotation}
              onChange={(option) => handleAnnotationChange(option?.value || '')}
              options={annotationOptions}
              placeholder="Select annotation..."
              isLoading={annotationsLoading}
              styles={selectStyles}
              isClearable
              isDisabled={!state.dataset}
            />
          </div>
        </>
      )}
    </div>
  )
}
