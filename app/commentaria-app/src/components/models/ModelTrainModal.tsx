import { type FormEvent, useMemo, useState } from 'react'
import { useQueries } from '@tanstack/react-query'
import Select from 'react-select'
import {
  AnnotationsService,
  type annotation_Annotation,
  type annotation_Reference,
  type common_OCRModelType,
  type model_Model,
  type model_ModelTraining,
} from '@hub-api'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { ListAdder } from '../core/ListAdder'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { selectStyles } from '../../styles/selectStyles'
import { useAnnotationGroupsQuery } from '../../queries/annotationGroups'
import { useAnnotationsQuery } from '../../queries/annotations'
import { useDatasetsQuery } from '../../queries/datasets'

interface ModelTrainModalProps {
  isOpen: boolean
  models: model_Model[]
  onClose: () => void
  onSubmit: (training: model_ModelTraining) => void
  isSaving?: boolean
  errorMessage?: string | null
}

type BaseAnnotationRow = {
  id: number
  datasetId: string | null
  annotationId: string | null
}

const emptyRow = (id: number): BaseAnnotationRow => ({
  id,
  datasetId: null,
  annotationId: null,
})

type AnnotationGroupOption = {
  value: string
  label: string
}

const trainableModelTypes: Array<{
  value: common_OCRModelType
  label: string
}> = [{ value: 'text', label: 'Text recognition' }]

const annotationKey = (
  datasetId: string | undefined,
  annotationId: string | undefined,
) => (datasetId && annotationId ? `${datasetId}:${annotationId}` : '')

export function ModelTrainModal({
  isOpen,
  models,
  onClose,
  onSubmit,
  isSaving = false,
  errorMessage = null,
}: ModelTrainModalProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [modelType, setModelType] = useState<common_OCRModelType>('text')
  const [epochs, setEpochs] = useState('')
  const [baseModelId, setBaseModelId] = useState<string | null>(null)
  const [selectedBaseGroupIds, setSelectedBaseGroupIds] = useState<string[]>([])
  const [baseAnnotationRows, setBaseAnnotationRows] = useState<
    BaseAnnotationRow[]
  >([])
  const [rowCounter, setRowCounter] = useState(0)
  const [error, setError] = useState<string | null>(null)
  const { data: datasets } = useDatasetsQuery()
  const { data: annotationGroups } = useAnnotationGroupsQuery()
  const datasetIds = useMemo(
    () =>
      (datasets ?? []).map((dataset) => dataset.id).filter(Boolean) as string[],
    [datasets],
  )
  const annotationQueries = useQueries({
    queries: datasetIds.map((datasetId) => ({
      queryKey: ['annotations', datasetId] as const,
      queryFn: () =>
        AnnotationsService.getDatasetsAnnotations({ dataSetId: datasetId }),
    })),
  })
  const eligibleAnnotationKeys = useMemo(() => {
    const keys = new Set<string>()
    annotationQueries.forEach((query) => {
      const annotations = query.data as annotation_Annotation[] | undefined
      annotations?.forEach((annotation) => {
        if (!annotation.ocred) return
        const key = annotationKey(annotation.dataset_id, annotation.id)
        if (key) keys.add(key)
      })
    })
    return keys
  }, [annotationQueries])
  const eligibleDatasetIds = useMemo(() => {
    const ids = new Set<string>()
    eligibleAnnotationKeys.forEach((key) => {
      const [datasetId] = key.split(':', 1)
      if (datasetId) ids.add(datasetId)
    })
    return ids
  }, [eligibleAnnotationKeys])

  const baseModelOptions = useMemo(() => {
    return models
      .filter((model) => model.id && model.type === modelType)
      .map((model) => ({
        value: model.id as string,
        label: model.name || (model.id as string),
      }))
  }, [modelType, models])

  const annotationGroupOptions = useMemo<AnnotationGroupOption[]>(() => {
    return (annotationGroups ?? [])
      .filter((group) => group.id)
      .flatMap((group) => {
        const count =
          group.annotations?.filter(
            (annotation) =>
              annotation.dataset_id &&
              annotation.id &&
              eligibleAnnotationKeys.has(
                annotationKey(annotation.dataset_id, annotation.id),
              ),
          ).length ?? 0
        if (!count) return []

        return [
          {
            value: group.id as string,
            label: `${group.name || (group.id as string)} (${count})`,
          },
        ]
      })
  }, [annotationGroups, eligibleAnnotationKeys])

  const handleAddBaseAnnotation = () => {
    setBaseAnnotationRows((current) => [...current, emptyRow(rowCounter + 1)])
    setRowCounter((current) => current + 1)
  }

  const handleRemoveBaseAnnotation = (id: number) => {
    setBaseAnnotationRows((current) => current.filter((row) => row.id !== id))
  }

  const handleUpdateBaseAnnotation = (
    id: number,
    updates: Partial<BaseAnnotationRow>,
  ) => {
    setBaseAnnotationRows((current) =>
      current.map((row) => (row.id === id ? { ...row, ...updates } : row)),
    )
  }

  const handleSubmit = (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    if (!name.trim()) {
      setError('Please provide a model name.')
      return
    }
    const hasInvalidBaseAnnotation = baseAnnotationRows.some(
      (row) => !row.datasetId || !row.annotationId,
    )
    if (hasInvalidBaseAnnotation) {
      setError('Please select both dataset and annotation for each base row.')
      return
    }
    const parsedEpochs = epochs.trim() ? Number(epochs) : 0
    if (
      !Number.isInteger(parsedEpochs) ||
      (epochs.trim() && parsedEpochs < 1)
    ) {
      setError('Epochs must be a positive whole number.')
      return
    }

    const explicitAnnotations: annotation_Reference[] = baseAnnotationRows
      .filter(
        (
          row,
        ): row is BaseAnnotationRow & {
          datasetId: string
          annotationId: string
        } =>
          !!row.datasetId &&
          !!row.annotationId &&
          eligibleAnnotationKeys.has(
            annotationKey(row.datasetId, row.annotationId),
          ),
      )
      .map((row) => ({ dataset_id: row.datasetId, id: row.annotationId }))

    const annotationsFromGroups: annotation_Reference[] =
      selectedBaseGroupIds.flatMap((groupId) => {
        return (
          annotationGroups
            ?.find((group) => group.id === groupId)
            ?.annotations?.flatMap((annotation) =>
              annotation.dataset_id &&
              annotation.id &&
              eligibleAnnotationKeys.has(
                annotationKey(annotation.dataset_id, annotation.id),
              )
                ? [{ dataset_id: annotation.dataset_id, id: annotation.id }]
                : [],
            ) ?? []
        )
      })

    const seen = new Set<string>()
    const baseAnnotations = [
      ...explicitAnnotations,
      ...annotationsFromGroups,
    ].filter((annotation) => {
      const key = `${annotation.dataset_id}:${annotation.id}`
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })

    if (baseAnnotations.length === 0) {
      setError('Please select at least one training annotation or group.')
      return
    }

    setError(null)
    onSubmit({
      model: {
        name: name.trim(),
        description: description.trim() || undefined,
        type: modelType,
        base_model_id: baseModelId || undefined,
        base_annotations: baseAnnotations,
      },
      epochs: parsedEpochs || undefined,
    })
  }

  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={onClose}
    >
      <form
        className="bg-white rounded-lg max-w-3xl w-full max-h-[85vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Train model</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Model type
            </label>
            <Select
              value={
                trainableModelTypes.find(
                  (option) => option.value === modelType,
                ) || null
              }
              onChange={(
                option: { value: common_OCRModelType; label: string } | null,
              ) => {
                setModelType(option?.value || 'text')
                setBaseModelId(null)
              }}
              options={trainableModelTypes}
              isDisabled={isSaving}
              styles={selectStyles<{
                value: common_OCRModelType
                label: string
              }>({
                controlWidth: 260,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Model name
            </label>
            <input
              type="text"
              autoComplete="on"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              disabled={isSaving}
              required
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Description (optional)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              rows={3}
              disabled={isSaving}
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Epochs (optional)
            </label>
            <input
              type="number"
              min={1}
              step={1}
              value={epochs}
              onChange={(e) => setEpochs(e.target.value)}
              className="w-40 p-2 border border-gray-300 rounded-md"
              disabled={isSaving}
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Base model (optional)
            </label>
            <Select
              value={
                baseModelOptions.find(
                  (option) => option.value === baseModelId,
                ) || null
              }
              onChange={(option: { value: string; label: string } | null) =>
                setBaseModelId(option?.value || null)
              }
              options={baseModelOptions}
              placeholder="Select base model..."
              isDisabled={isSaving}
              styles={selectStyles<{ value: string; label: string }>({
                controlWidth: 260,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Training annotation groups
            </label>
            <Select<AnnotationGroupOption, true>
              value={annotationGroupOptions.filter((option) =>
                selectedBaseGroupIds.includes(option.value),
              )}
              onChange={(options) =>
                setSelectedBaseGroupIds(
                  (options ?? []).map((option) => option.value),
                )
              }
              options={annotationGroupOptions}
              placeholder="Select groups..."
              isDisabled={isSaving}
              styles={selectStyles<AnnotationGroupOption, true>({
                controlWidth: 320,
                isMulti: true,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isMulti
            />
          </div>

          <ListAdder
            label="Training annotations"
            items={baseAnnotationRows}
            onAdd={handleAddBaseAnnotation}
            isDisabled={isSaving}
            emptyLabel="No individual training annotations selected."
            renderItem={(row) => (
              <BaseAnnotationPicker
                row={row}
                datasets={datasets ?? []}
                eligibleDatasetIds={eligibleDatasetIds}
                isSaving={isSaving}
                onChange={(updates) =>
                  handleUpdateBaseAnnotation(row.id, updates)
                }
                onRemove={() => handleRemoveBaseAnnotation(row.id)}
              />
            )}
          />

          {error && <ErrorMessage message={error} />}
          {!error && errorMessage && <ErrorMessage message={errorMessage} />}
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          {isSaving ? (
            <LoadingSpinner size="sm" message="Submitting training..." />
          ) : (
            <>
              <Button
                type="button"
                onClick={onClose}
                className="px-3 py-1.5 text-sm"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                className="px-3 py-1.5 text-sm"
              >
                Train
              </Button>
            </>
          )}
        </div>
      </form>
    </div>
  )
}

type BaseAnnotationPickerProps = {
  row: BaseAnnotationRow
  datasets: Array<{ id?: string; name?: string }>
  eligibleDatasetIds: Set<string>
  isSaving: boolean
  onChange: (updates: Partial<BaseAnnotationRow>) => void
  onRemove: () => void
}

function BaseAnnotationPicker({
  row,
  datasets,
  eligibleDatasetIds,
  isSaving,
  onChange,
  onRemove,
}: BaseAnnotationPickerProps) {
  const datasetId = row.datasetId || ''
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(datasetId)

  const datasetOptions = useMemo(
    () =>
      datasets
        .filter((dataset) => dataset.id && eligibleDatasetIds.has(dataset.id))
        .map((dataset) => ({
          value: dataset.id as string,
          label: dataset.name || (dataset.id as string),
        })),
    [datasets, eligibleDatasetIds],
  )

  const annotationOptions = useMemo(() => {
    if (!annotations) return []
    return annotations
      .filter((annotation) => annotation.id && annotation.ocred)
      .map((annotation) => ({
        value: annotation.id as string,
        label: annotation.name || (annotation.id as string),
      }))
  }, [annotations])

  return (
    <div className="flex flex-col sm:flex-row sm:items-center gap-2">
      <Select
        value={
          datasetOptions.find((option) => option.value === row.datasetId) ||
          null
        }
        onChange={(option: { value: string; label: string } | null) =>
          onChange({
            datasetId: option?.value || null,
            annotationId: null,
          })
        }
        options={datasetOptions}
        placeholder="Select dataset..."
        isDisabled={isSaving}
        styles={selectStyles<{ value: string; label: string }>({
          controlWidth: 220,
        })}
        menuPortalTarget={document.body}
        menuPosition="fixed"
      />
      <Select
        value={
          annotationOptions.find(
            (option) => option.value === row.annotationId,
          ) || null
        }
        onChange={(option: { value: string; label: string } | null) =>
          onChange({ annotationId: option?.value || null })
        }
        options={annotationOptions}
        placeholder="Select annotation..."
        isLoading={annotationsLoading}
        isDisabled={isSaving || !row.datasetId}
        styles={selectStyles<{ value: string; label: string }>({
          controlWidth: 260,
        })}
        menuPortalTarget={document.body}
        menuPosition="fixed"
      />
      <Button
        type="button"
        variant="danger"
        className="px-2 py-1 text-xs"
        onClick={onRemove}
        disabled={isSaving}
      >
        Remove
      </Button>
    </div>
  )
}
