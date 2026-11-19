import { useMemo } from 'react'
import Select from 'react-select'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import { selectStyles } from '../styles/selectStyles.ts'
import { useAppState } from '../context/useAppState.ts'
import { usePipelineStages } from '../queries/metadata.ts'
import useLocalStorageState from 'use-local-storage-state'
import { MultiSelectDropdown } from './core/MultiSelectDropdown.tsx'

import { Button } from './core/Button.tsx'
import {
  DATASET_COMPLETENESS_ITEMS,
  getDatasetCompleteness,
  type DatasetCompleteness,
} from '../utils/datasets.ts'
import {
  buildAnnotationOptions,
  type AnnotationFilterItem,
} from '../utils/annotationOptions.ts'
import { StageFilterDropdown } from './StageFilterDropdown.tsx'
import { useDatasetOptions } from './useDatasetOptions.ts'

const Separator = () => <span className="self-stretch bg-gray-600 w-px mx-2" />

export function BreadcrumbNav() {
  const { state, setState } = useAppState()

  const {
    datasets,
    datasetsLoading,
    datasetOptions,
    selectedDatasetCompleteness,
    setSelectedDatasetCompleteness,
  } = useDatasetOptions()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(state.datasetId)
  const isCompareMode = state.viewMode === 'compare'
  const { data: stages } = usePipelineStages()
  const [selectedStages, setSelectedStages] = useLocalStorageState<
    AnnotationFilterItem[] | null
  >('annotationFilterStages', {
    defaultValue: null,
  })

  const effectiveSelectedStages = useMemo<AnnotationFilterItem[] | null>(() => {
    if (selectedStages != null) {
      return selectedStages
    }
    return stages || null
  }, [selectedStages, stages])

  const annotationOptions = useMemo(
    () =>
      buildAnnotationOptions(
        annotations,
        effectiveSelectedStages,
        state.annotationId,
      ),
    [annotations, effectiveSelectedStages, state.annotationId],
  )

  const selectedDataset =
    datasetOptions.find((d) => d.value === state.datasetId) || null
  const selectedAnnotation =
    annotationOptions.find((a) => a.value === state.annotationId) || null
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

  const highlightDataset = !state.viewMode && !state.datasetId
  const highlightAnnotation =
    !state.viewMode && !!state.datasetId && !state.annotationId

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
      <Button
        variant="primary"
        className={`h-8 w-16 px-2 text-xs px-2 ${isCompareMode && '!bg-teal-100 hover:!bg-white'}`}
        onClick={() =>
          setState(
            isCompareMode
              ? { viewMode: null }
              : {
                  viewMode: 'compare',
                  leftDatasetId: state.datasetId,
                  leftAnnotationId: state.annotationId,
                  leftPage: state.currentPageOrKey,
                },
          )
        }
      >
        Compare
      </Button>

      {!isCompareMode && (
        <>
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
        </>
      )}
    </div>
  )
}
