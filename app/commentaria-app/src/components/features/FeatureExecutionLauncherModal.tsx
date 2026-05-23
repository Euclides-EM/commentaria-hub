import { useEffect, useMemo, useState } from 'react'
import Select from 'react-select'
import { useAnnotationsQuery } from '../../queries/annotations'
import { ErrorMessage } from '../core/ErrorMessage'
import { Button } from '../core/Button'
import { selectStyles } from '../../styles/selectStyles'

type ScopeOption = 'editions' | 'dataset'

type EntityOption = {
  value: string
  label: string
}

type ExecutionTarget =
  | { scope: 'editions' }
  | { scope: 'dataset'; datasetId: string; annotationId: string }

interface FeatureExecutionLauncherModalProps {
  isOpen: boolean
  onClose: () => void
  onContinue: (target: ExecutionTarget) => void
  datasetOptions: EntityOption[]
  loadingDatasets: boolean
}

export function FeatureExecutionLauncherModal({
  isOpen,
  onClose,
  onContinue,
  datasetOptions,
  loadingDatasets,
}: FeatureExecutionLauncherModalProps) {
  const [scope, setScope] = useState<ScopeOption>('editions')
  const [selectedDataset, setSelectedDataset] = useState<EntityOption | null>(
    null,
  )
  const [selectedAnnotation, setSelectedAnnotation] =
    useState<EntityOption | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)

  const datasetId = selectedDataset?.value || ''
  const annotationsQuery = useAnnotationsQuery(datasetId)

  const annotationOptions = useMemo(
    () =>
      [...(annotationsQuery.data ?? [])]
        .filter((annotation) => !!annotation.id)
        .map((annotation) => ({
          value: annotation.id!,
          label: annotation.name || annotation.id!,
        }))
        .sort((left, right) =>
          left.label.localeCompare(right.label, undefined, {
            sensitivity: 'base',
          }),
        ),
    [annotationsQuery.data],
  )

  useEffect(() => {
    if (!isOpen) return
    setScope('editions')
    setSelectedDataset(null)
    setSelectedAnnotation(null)
    setValidationError(null)
  }, [isOpen])

  useEffect(() => {
    if (scope !== 'dataset') {
      setSelectedDataset(null)
      setSelectedAnnotation(null)
      setValidationError(null)
    }
  }, [scope])

  useEffect(() => {
    setSelectedAnnotation(null)
    setValidationError(null)
  }, [datasetId])

  useEffect(() => {
    if (selectedAnnotation) {
      setValidationError(null)
    }
  }, [selectedAnnotation])

  if (!isOpen) return null

  const handleContinue = () => {
    if (scope === 'editions') {
      onContinue({ scope: 'editions' })
      return
    }
    if (!selectedDataset?.value) {
      setValidationError('Select a dataset.')
      return
    }
    if (!selectedAnnotation?.value) {
      setValidationError('Select an annotation.')
      return
    }
    onContinue({
      scope: 'dataset',
      datasetId: selectedDataset.value,
      annotationId: selectedAnnotation.value,
    })
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] flex flex-col m-4"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Execute Features</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4">
          <div className="space-y-2">
            <div className="text-sm font-medium text-gray-700">Scope</div>
            <select
              value={scope}
              onChange={(event) => setScope(event.target.value as ScopeOption)}
              className="h-9 px-3 text-sm border border-gray-300 rounded-md bg-white"
            >
              <option value="editions">Editions</option>
              <option value="dataset">Dataset</option>
            </select>
          </div>

          {scope === 'dataset' && (
            <div className="grid gap-4">
              <div className="space-y-2">
                <div className="text-sm font-medium text-gray-700">Dataset</div>
                <Select
                  value={selectedDataset}
                  onChange={(option) => setSelectedDataset(option)}
                  options={datasetOptions}
                  placeholder="Select dataset..."
                  styles={selectStyles<EntityOption>({ controlWidth: 320 })}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                  isClearable
                  isLoading={loadingDatasets}
                />
              </div>

              <div className="space-y-2">
                <div className="text-sm font-medium text-gray-700">
                  Annotation
                </div>
                <Select
                  value={selectedAnnotation}
                  onChange={(option) => setSelectedAnnotation(option)}
                  options={annotationOptions}
                  placeholder={
                    selectedDataset ? 'Select annotation...' : 'Select dataset first'
                  }
                  styles={selectStyles<EntityOption>({ controlWidth: 320 })}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                  isClearable
                  isDisabled={!selectedDataset}
                  isLoading={annotationsQuery.isLoading}
                />
              </div>
            </div>
          )}

          <ErrorMessage
            message={
              validationError ||
              (annotationsQuery.error instanceof Error
                ? annotationsQuery.error.message
                : null)
            }
          />
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm"
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="primary"
            className="px-3 py-1.5 text-sm"
            onClick={handleContinue}
          >
            Continue
          </Button>
        </div>
      </div>
    </div>
  )
}
