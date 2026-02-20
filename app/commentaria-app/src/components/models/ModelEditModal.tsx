import { useMemo, useState, type FormEvent } from 'react'
import type { model_Model } from '@hub-api'
import type { model_AnnotationReference } from '@hub-api'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { ListAdder } from '../core/ListAdder'
import Select from 'react-select'
import { selectStyles } from '../../styles/selectStyles.ts'
import { useDatasetsQuery } from '../../queries/datasets'
import { useAnnotationsQuery } from '../../queries/annotations'

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

function refsToRows(
  refs: model_AnnotationReference[] | undefined,
  startId: number,
): { rows: BaseAnnotationRow[]; nextId: number } {
  if (!refs?.length) return { rows: [], nextId: startId }
  const rows = refs.map((ref, i) => ({
    id: startId + i,
    datasetId: ref.dataset_id ?? null,
    annotationId: ref.id ?? null,
  }))
  return { rows, nextId: startId + rows.length }
}

interface ModelEditModalProps {
  model: model_Model | null
  allModels: model_Model[]
  onClose: () => void
  onSubmit: (updates: {
    name: string
    description?: string
    type: string
    algorithm_family?: string
    base_model_id?: string
    base_annotations?: model_AnnotationReference[]
  }) => void
  isSaving?: boolean
  errorMessage?: string | null
}

export function ModelEditModal({
  model,
  allModels,
  onClose,
  onSubmit,
  isSaving = false,
  errorMessage = null,
}: ModelEditModalProps) {
  if (!model) {
    return null
  }

  return (
    <ModelEditModalContent
      key={model.id ?? `${model.name ?? ''}:${model.type ?? ''}`}
      model={model}
      allModels={allModels}
      onClose={onClose}
      onSubmit={onSubmit}
      isSaving={isSaving}
      errorMessage={errorMessage}
    />
  )
}

type ModelEditModalContentProps = Omit<ModelEditModalProps, 'model'> & {
  model: model_Model
}

function ModelEditModalContent({
  model,
  allModels,
  onClose,
  onSubmit,
  isSaving = false,
  errorMessage = null,
}: ModelEditModalContentProps) {
  const { rows: initialBaseRows, nextId: initialRowId } = useMemo(
    () => refsToRows(model.base_annotations, 0),
    [model.base_annotations],
  )
  const [name, setName] = useState(() => model.name || '')
  const [description, setDescription] = useState(() => model.description || '')
  const [baseModelId, setBaseModelId] = useState<string | null>(
    () => model.base_model_id || null,
  )
  const [baseAnnotationRows, setBaseAnnotationRows] = useState<
    BaseAnnotationRow[]
  >(() => initialBaseRows)
  const [rowCounter, setRowCounter] = useState(() => initialRowId)
  const [error, setError] = useState<string | null>(null)
  const { data: datasets } = useDatasetsQuery()

  const baseModelOptions = useMemo(() => {
    return allModels
      .filter((item) => item.id && item.id !== model.id)
      .map((item) => ({
        value: item.id as string,
        label: item.name || (item.id as string),
      }))
  }, [allModels, model.id])

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
    const base_annotations: model_AnnotationReference[] = baseAnnotationRows
      .filter(
        (
          row,
        ): row is BaseAnnotationRow & {
          datasetId: string
          annotationId: string
        } => !!row.datasetId && !!row.annotationId,
      )
      .map((row) => ({ dataset_id: row.datasetId, id: row.annotationId }))
    onSubmit({
      name: name.trim(),
      description: description.trim() || undefined,
      type: model.type || 'segment',
      algorithm_family: model.algorithm_family || undefined,
      base_model_id: baseModelId || undefined,
      base_annotations,
    })
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={onClose}
    >
      <form
        className="bg-white rounded-lg max-w-xl w-full max-h-[85vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Edit model</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Name
            </label>
            <input
              type="text"
              autoComplete="on"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
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
              styles={selectStyles<{ value: string; label: string }>({
                controlWidth: 260,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          </div>

          <ListAdder
            label="Base annotations (optional)"
            items={baseAnnotationRows}
            onAdd={handleAddBaseAnnotation}
            isDisabled={isSaving}
            emptyLabel="No base annotations selected."
            renderItem={(row) => (
              <BaseAnnotationPicker
                row={row}
                datasets={datasets ?? []}
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
          <Button
            onClick={onClose}
            type="button"
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isSaving}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isSaving}
          >
            {isSaving ? 'Saving...' : 'Save changes'}
          </Button>
        </div>
      </form>
    </div>
  )
}

type BaseAnnotationPickerProps = {
  row: BaseAnnotationRow
  datasets: Array<{ id?: string; name?: string }>
  isSaving: boolean
  onChange: (updates: Partial<BaseAnnotationRow>) => void
  onRemove: () => void
}

function BaseAnnotationPicker({
  row,
  datasets,
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
        .filter((dataset) => dataset.id)
        .map((dataset) => ({
          value: dataset.id as string,
          label: dataset.name || (dataset.id as string),
        })),
    [datasets],
  )

  const annotationOptions = useMemo(() => {
    if (!annotations) return []
    return annotations
      .filter((annotation) => annotation.id)
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
          onChange({ datasetId: option?.value || null, annotationId: null })
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
