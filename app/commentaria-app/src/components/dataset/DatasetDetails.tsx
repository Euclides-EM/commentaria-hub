import { type ChangeEvent, useEffect, useRef, useState } from 'react'
import { DatasetsService, ApiError, type model_Dataset } from '@hub-api'
import { AnnotationActions } from '../annotation/AnnotationActions.tsx'
import { DeleteAnnotationModal } from '../modal/DeleteAnnotationModal.tsx'
import { NoticeModal } from '../modal/NoticeModal.tsx'
import { ReplaceImageModal } from '../modal/ReplaceImageModal.tsx'
import { useAppState } from '../../context/useAppState.ts'
import {
  useDatasetsQuery,
  useReplaceDatasetImageMutation,
} from '../../queries/datasets.ts'
import { useAuthStore } from '../../store/authStore.ts'
import { TabButton } from '../core/TabButton.tsx'
import { DatasetDetailsTab } from './DatasetDetailsTab.tsx'
import { DatasetFeaturesTab } from './DatasetFeaturesTab.tsx'
import { TITLE_PAGES_DATASET_ID } from '../../utils/editions.ts'

type DatasetStatus = 'creating' | 'ready' | 'failed'

const DATASET_STATUS_LABELS: Record<DatasetStatus, string> = {
  creating: 'Creating',
  ready: 'Ready',
  failed: 'Failed',
}

export const DatasetDetails = () => {
  const { data: datasets } = useDatasetsQuery()
  const { state, setState, refetch } = useAppState()
  const activeTab = state.datasetTab
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const currentDataset = datasets?.find((d) => d.id === state.datasetId) as
    | model_Dataset
    | undefined
  const editionId = currentDataset?.edition_id || null
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
  const [editedDenoised, setEditedDenoised] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [isUploadImagesOpen, setIsUploadImagesOpen] = useState(false)
  const [uploadImagesError, setUploadImagesError] = useState<string | null>(
    null,
  )
  const [uploadSuccessCount, setUploadSuccessCount] = useState<number | null>(
    null,
  )
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const uploadImagesMutation = useReplaceDatasetImageMutation()

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
    setEditedDenoised(!!currentDataset.denoised)
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
          denoised: editedDenoised,
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

  const handleUploadImagesConfirm = () => {
    setUploadImagesError(null)
    setIsUploadImagesOpen(false)
    fileInputRef.current?.click()
  }

  const handleUploadImagesCancel = () => {
    if (uploadImagesMutation.isPending) {
      return
    }
    setUploadImagesError(null)
    setIsUploadImagesOpen(false)
  }

  const handleUploadImagesChange = async (
    event: ChangeEvent<HTMLInputElement>,
  ) => {
    const file = event.target.files?.[0]
    event.target.value = ''

    if (!file || !currentDataset?.id || isCreating) {
      return
    }

    try {
      setUploadImagesError(null)
      const result = await uploadImagesMutation.mutateAsync({
        datasetId: currentDataset.id,
        type: currentDataset.id === TITLE_PAGES_DATASET_ID ? 'tp' : 'facsimile',
        file,
      })
      setUploadSuccessCount(result.uploaded ?? 0)
    } catch (e) {
      setUploadImagesError(
        e instanceof ApiError
          ? typeof e.body === 'string'
            ? e.body
            : JSON.stringify(e.body)
          : String(e),
      )
      setIsUploadImagesOpen(true)
    }
  }

  if (!currentDataset) {
    return <AnnotationActions dataSetId={state.datasetId} />
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex w-full gap-4 p-3 border-b border-gray-200 bg-white items-center justify-center">
        <TabButton
          onSelected={() => setState({ datasetTab: 'details' })}
          title="Dataset Details"
          isActive={activeTab === 'details'}
        />
        <TabButton
          onSelected={() => setState({ datasetTab: 'features' })}
          title="Dataset Features"
          isActive={activeTab === 'features'}
        />
      </div>

      <div className="flex-1 overflow-auto">
        {activeTab === 'details' && (
          <DatasetDetailsTab
            dataset={currentDataset}
            editionId={editionId}
            datasetStatusLabel={datasetStatusLabel}
            isCreating={isCreating}
            isEditing={isEditing}
            isAuthenticated={isAuthenticated}
            isSaving={isSaving}
            isUploadingImages={uploadImagesMutation.isPending}
            editedName={editedName}
            editedDescription={editedDescription}
            editedDpi={editedDpi}
            editedPages={editedPages}
            editedDeskewed={editedDeskewed}
            editedDenoised={editedDenoised}
            error={error}
            onEditClick={handleEditClick}
            onDeleteClick={() => {
              setError(null)
              setIsDeleteOpen(true)
            }}
            onUploadImagesClick={() => {
              setUploadImagesError(null)
              setUploadSuccessCount(null)
              setIsUploadImagesOpen(true)
            }}
            onNameChange={setEditedName}
            onDescriptionChange={setEditedDescription}
            onDpiChange={setEditedDpi}
            onPagesChange={setEditedPages}
            onDeskewedChange={setEditedDeskewed}
            onDenoisedChange={setEditedDenoised}
            onCancel={handleCancel}
            onSave={handleSave}
          />
        )}

        {activeTab === 'features' && <DatasetFeaturesTab />}
      </div>

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
      <input
        ref={fileInputRef}
        type="file"
        accept=".zip,application/zip"
        className="sr-only"
        onChange={(event) => {
          void handleUploadImagesChange(event)
        }}
        disabled={uploadImagesMutation.isPending}
      />
      <ReplaceImageModal
        isOpen={isUploadImagesOpen}
        title="Upload images"
        body={`Upload a ZIP file to replace images for dataset ${
          currentDataset.name || currentDataset.id
        }? This will overwrite existing images for matching entries in this dataset.`}
        confirmLabel="Choose ZIP"
        loadingMessage="Uploading images..."
        isReplacing={uploadImagesMutation.isPending}
        error={uploadImagesError}
        onCancel={handleUploadImagesCancel}
        onConfirm={handleUploadImagesConfirm}
      />
      <NoticeModal
        isOpen={uploadSuccessCount != null}
        title="Upload complete"
        message={`Successfully uploaded ${uploadSuccessCount ?? 0} images for dataset ${currentDataset.name || currentDataset.id}`}
        onClose={() => setUploadSuccessCount(null)}
      />
    </div>
  )
}
