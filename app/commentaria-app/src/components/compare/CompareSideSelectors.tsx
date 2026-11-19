import { useMemo } from 'react'
import Select from 'react-select'
import { useAnnotationsQuery } from '../../queries/annotations.ts'
import { selectStyles } from '../../styles/selectStyles.ts'
import { useAppState } from '../../context/useAppState.ts'
import type { AppState } from '../../context/AppStateContext.ts'
import {
  COMPARE_SIDE_KEYS,
  type ComparePane,
  type CompareSide,
} from '../../context/compareSides.ts'
import { buildAnnotationOptions } from '../../utils/annotationOptions.ts'
import { Button } from '../core/Button.tsx'
import { useDatasetOptions } from '../useDatasetOptions.ts'
import { ComparePaneSelect } from './ComparePaneSelect.tsx'

interface CompareSideSelectorsProps {
  side: CompareSide
  selectedPanes: ComparePane[] | null
  setSelectedPanes: (panes: ComparePane[] | null) => void
  onResetDisplay: () => void
}

export function CompareSideSelectors({
  side,
  selectedPanes,
  setSelectedPanes,
  onResetDisplay,
}: CompareSideSelectorsProps) {
  const { state, setState } = useAppState()
  const keys = COMPARE_SIDE_KEYS[side]
  const datasetId = state[keys.datasetId]
  const annotationId = state[keys.annotationId]
  const { datasetOptions, datasetsLoading } = useDatasetOptions()
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(datasetId)

  const annotationOptions = useMemo(
    () => buildAnnotationOptions(annotations, null, annotationId),
    [annotations, annotationId],
  )

  const selectedDataset =
    datasetOptions.find((d) => d.value === datasetId) || null
  const selectedAnnotation =
    annotationOptions.find((a) => a.value === annotationId) || null

  const handleDatasetChange = (value: string) => {
    const updates: Partial<AppState> = {}
    updates[keys.datasetId] = value
    updates[keys.annotationId] = ''
    setState(updates)
  }

  const handleAnnotationChange = (value: string) => {
    const updates: Partial<AppState> = {}
    updates[keys.annotationId] = value
    setState(updates)
  }

  return (
    <div className="flex items-center text-sm gap-x-2 gap-y-2 flex-wrap">
      <div
        className={`text-shadow-gray-800 font-semibold px-1 rounded ${!datasetId ? 'label-glow text-teal-800' : ''}`}
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
      {datasetId && (
        <div className="flex items-center gap-2 flex-nowrap shrink-0">
          <div className="h-3 w-3 rotate-[-45deg] border-b border-r border-slate-600" />
          <div
            className={`text-shadow-gray-800 font-semibold px-1 rounded ${!annotationId ? 'label-glow text-teal-800' : ''}`}
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
            />
          </div>
        </div>
      )}
      <ComparePaneSelect
        selectedPanes={selectedPanes}
        setSelectedPanes={setSelectedPanes}
      />
      <Button
        onClick={onResetDisplay}
        className="h-8 px-2 text-xs shrink-0"
        title="Reset sizes and display settings of this side"
      >
        Reset display
      </Button>
    </div>
  )
}
