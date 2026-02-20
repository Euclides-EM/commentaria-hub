import { useEffect, useState } from 'react'
import { DatasetsService, ApiError } from '../../api'
import { AnnotationActions } from '../annotation/AnnotationActions.tsx'
import { Button } from '../core/Button.tsx'
import { ErrorMessage } from '../core/ErrorMessage.tsx'
import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { Timestamp } from '../core/Timestamp.tsx'
import { DeleteAnnotationModal } from '../modal/DeleteAnnotationModal.tsx'
import { useAppState } from '../../context/useAppState.ts'
import { useDatasetsQuery } from '../../queries/datasets.ts'
import { useAuthStore } from '../../store/authStore.ts'

type DatasetStatus = 'creating' | 'ready' | 'failed'

const DATASET_STATUS_LABELS: Record<DatasetStatus, string> = {
  creating: 'Creating',
  ready: 'Ready',
  failed: 'Failed',
}

export const DatasetDetails = () => {
  const { data: datasets } = useDatasetsQuery()
  const { state, setState, refetch } = useAppState()
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const currentDataset = datasets?.find((d) => d.id === state.datasetId)
  const isCreating = currentDataset?.status === 'creating'
  const datasetStatusLabel =
    currentDataset?.status && currentDataset.status in DATASET_STATUS_LABELS
      ? DATASET_STATUS_LABELS[currentDataset.status as DatasetStatus]
      : currentDataset?.status || 'N/A'

  const [isEditing, setIsEditing] = useState(false)
  const [editedName, setEditedName] = useState('')
  const [editedDescription, setEditedDescription] = useState('')
  const [editedDpi, setEditedDpi] = useState('')
  const [editedPages, setEditedPages] = useState('')
  const [editedDeskewed, setEditedDeskewed] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)

  useEffect(() => {
    setError(null)
    setIsEditing(false)
  }, [state.datasetId])

  const handleEditClick = () => {
    if (!currentDataset || isCreating) {
      return
    }
    setError(null)
    setEditedName(currentDataset.name || '')
    setEditedDescription(currentDataset.description || '')
    setEditedDpi(currentDataset.dpi != null ? String(currentDataset.dpi) : '')
    setEditedPages(currentDataset.pages || '')
    setEditedDeskewed(!!currentDataset.deskewed)
    setIsEditing(true)
  }

  const handleCancel = () => {
    setIsEditing(false)
    setError(null)
  }

  const handleSave = async () => {
    if (!currentDataset?.id || isCreating) {
      return
    }

    const parsedDpi =
      editedDpi.trim() === '' ? undefined : Number.parseInt(editedDpi, 10)
    if (parsedDpi != null && (Number.isNaN(parsedDpi) || parsedDpi <= 0)) {
      setError('DPI must be a positive integer.')
      return
    }

    try {
      setError(null)
      setIsSaving(true)
      await DatasetsService.putDatasets({
        dataSetId: currentDataset.id,
        dataset: {
          name: editedName.trim(),
          description: editedDescription.trim() || undefined,
          dpi: parsedDpi,
          pages: editedPages.trim() || undefined,
          deskewed: editedDeskewed,
        },
      })
      await refetch()
      setIsEditing(false)
    } catch (e) {
      console.error('Failed to update dataset:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    } finally {
      setIsSaving(false)
    }
  }

  const handleDeleteConfirm = async () => {
    if (!currentDataset?.id || isCreating) {
      return
    }

    try {
      setError(null)
      setIsDeleting(true)
      await DatasetsService.deleteDatasets({
        dataSetId: currentDataset.id,
      })
      setIsDeleteOpen(false)
      setState({ datasetId: '', annotationId: '' })
      await refetch()
    } catch (e) {
      console.error('Failed to delete dataset:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    } finally {
      setIsDeleting(false)
    }
  }

  if (!currentDataset) {
    return <AnnotationActions dataSetId={state.datasetId} />
  }

  return (
    <div className="h-full flex flex-col overflow-auto">
      <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col bg-white m-3 mb-0 w-full max-w-[80vw] self-center">
        <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
          <div>Dataset Details</div>
          {!isEditing && isAuthenticated && (
            <div className="flex items-center gap-2">
              <Button
                onClick={handleEditClick}
                className="px-2 py-1 text-xs"
                disabled={isCreating}
              >
                Edit
              </Button>
              <Button
                onClick={() => {
                  setError(null)
                  setIsDeleteOpen(true)
                }}
                variant="danger"
                className="px-2 py-1 text-xs"
                disabled={isCreating}
              >
                Delete
              </Button>
            </div>
          )}
        </div>

        <div className="p-3.5 overflow-auto leading-normal text-base box-border bg-gray-50">
          {isCreating && (
            <div className="mb-4">
              <LoadingSpinner size="sm" message="Dataset is being created..." />
              <p className="mt-2 text-sm text-gray-600">
                You can select another dataset from the sidebar. This one will
                be ready once creation finishes.
              </p>
            </div>
          )}
          <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 items-start">
            <div className="font-semibold text-xs opacity-80 pt-0.5">ID</div>
            <div className="text-sm leading-tight break-all font-mono">
              {currentDataset.id}
            </div>

            <div className="font-semibold text-xs opacity-80 pt-0.5">Name</div>
            {isEditing ? (
              <input
                type="text"
                autoComplete="on"
                value={editedName}
                onChange={(e) => setEditedName(e.target.value)}
                className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                disabled={isSaving}
              />
            ) : (
              <div className="text-sm leading-tight break-all">
                {currentDataset.name || 'N/A'}
              </div>
            )}

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Description
            </div>
            {isEditing ? (
              <textarea
                value={editedDescription}
                onChange={(e) => setEditedDescription(e.target.value)}
                className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                rows={3}
                disabled={isSaving}
              />
            ) : (
              <div className="text-sm leading-tight whitespace-pre-wrap break-words">
                {currentDataset.description?.replace(/\\n/g, '\n') || 'N/A'}
              </div>
            )}

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Facsimile
            </div>
            <div className="text-sm leading-tight break-all font-mono">
              {currentDataset.facsimile_id
                ? currentDataset.edition_id
                  ? `${currentDataset.edition_id} (${currentDataset.facsimile_id})`
                  : currentDataset.facsimile_id
                : 'N/A'}
            </div>

            <div className="font-semibold text-xs opacity-80 pt-0.5">DPI</div>
            {isEditing ? (
              <input
                type="number"
                min={1}
                value={editedDpi}
                onChange={(e) => setEditedDpi(e.target.value)}
                className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                disabled={isSaving}
              />
            ) : (
              <div className="text-sm leading-tight break-all">
                {currentDataset.dpi ?? 'N/A'}
              </div>
            )}

            <div className="font-semibold text-xs opacity-80 pt-0.5">Pages</div>
            {isEditing ? (
              <input
                type="text"
                autoComplete="off"
                value={editedPages}
                onChange={(e) => setEditedPages(e.target.value)}
                className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                disabled={isSaving}
              />
            ) : (
              <div className="text-sm leading-tight break-all">
                {currentDataset.pages || 'All'}
              </div>
            )}

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Deskewed
            </div>
            {isEditing ? (
              <label className="flex items-center gap-2 text-sm leading-tight">
                <input
                  type="checkbox"
                  checked={editedDeskewed}
                  onChange={(e) => setEditedDeskewed(e.target.checked)}
                  className="h-4 w-4"
                  disabled={isSaving}
                />
                {String(editedDeskewed)}
              </label>
            ) : (
              <div className="text-sm leading-tight break-all">
                {String(!!currentDataset.deskewed)}
              </div>
            )}

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Edition ID
            </div>
            <div className="text-sm leading-tight break-all font-mono">
              {currentDataset.edition_id || 'N/A'}
            </div>

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Status
            </div>
            <div className="text-sm leading-tight break-all">
              {datasetStatusLabel}
            </div>

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Creation error
            </div>
            <div className="text-sm leading-tight break-all">
              {currentDataset.creation_error || 'None'}
            </div>

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Created
            </div>
            <div className="text-sm leading-tight break-all">
              <Timestamp date={currentDataset.created_at} />
            </div>

            <div className="font-semibold text-xs opacity-80 pt-0.5">
              Updated
            </div>
            <div className="text-sm leading-tight break-all">
              <Timestamp date={currentDataset.updated_at} />
            </div>
          </div>

          <div className="mt-4">
            <ErrorMessage message={error} />
          </div>
          {isEditing && (
            <div className="flex justify-end gap-2 mt-4">
              <Button
                onClick={handleCancel}
                className="px-3 py-1.5 text-sm font-semibold"
                disabled={isSaving}
              >
                Cancel
              </Button>
              <Button
                onClick={handleSave}
                variant="primary"
                className="px-3 py-1.5 text-sm font-semibold"
                disabled={isSaving || isCreating}
              >
                {isSaving ? 'Saving...' : 'Save'}
              </Button>
            </div>
          )}
        </div>
      </section>

      <AnnotationActions dataSetId={state.datasetId} />

      <DeleteAnnotationModal
        isOpen={isDeleteOpen}
        annotationLabel={currentDataset.name || currentDataset.id || ''}
        title="Delete dataset"
        loadingMessage="Deleting dataset..."
        error={error}
        isDeleting={isDeleting}
        onCancel={() => setIsDeleteOpen(false)}
        onConfirm={handleDeleteConfirm}
      />
    </div>
  )
}
