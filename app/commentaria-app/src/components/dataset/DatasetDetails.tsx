import { useEffect, useState } from 'react'
import { DatasetsService, ApiError, type model_Dataset } from '@hub-api'
import { AnnotationActions } from '../annotation/AnnotationActions.tsx'
import { DeleteAnnotationModal } from '../modal/DeleteAnnotationModal.tsx'
import { useAppState } from '../../context/useAppState.ts'
import { useDatasetsQuery } from '../../queries/datasets.ts'
import { useAuthStore } from '../../store/authStore.ts'
import useLocalStorageState from 'use-local-storage-state'
import { TabButton } from '../core/TabButton.tsx'
import { DatasetDetailsTab } from './DatasetDetailsTab.tsx'
import { DatasetFeaturesTab } from './DatasetFeaturesTab.tsx'

type DatasetStatus = 'creating' | 'ready' | 'failed'
type DatasetTab = 'details' | 'features'

const DATASET_STATUS_LABELS: Record<DatasetStatus, string> = {
  creating: 'Creating',
  ready: 'Ready',
  failed: 'Failed',
}

export const DatasetDetails = () => {
  const { data: datasets } = useDatasetsQuery()
  const { state, setState, refetch } = useAppState()
  const [activeTab, setActiveTab] = useLocalStorageState<DatasetTab>(
    'dataset-tab',
    { defaultValue: 'details' },
  )
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const currentDataset = datasets?.find((d) => d.id === state.datasetId) as
    | model_Dataset
    | undefined
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
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex w-full gap-4 p-3 border-b border-gray-200 bg-white items-center justify-center">
        <TabButton
          onSelected={() => setActiveTab('details')}
          title="Dataset Details"
          isActive={activeTab === 'details'}
        />
        <TabButton
          onSelected={() => setActiveTab('features')}
          title="Dataset Features"
          isActive={activeTab === 'features'}
        />
      </div>

      <div className="flex-1 overflow-auto">
        {activeTab === 'details' && (
          <DatasetDetailsTab
            dataset={currentDataset}
            datasetStatusLabel={datasetStatusLabel}
            isCreating={isCreating}
            isEditing={isEditing}
            isAuthenticated={isAuthenticated}
            isSaving={isSaving}
            editedName={editedName}
            editedDescription={editedDescription}
            editedDpi={editedDpi}
            editedPages={editedPages}
            editedDeskewed={editedDeskewed}
            error={error}
            onEditClick={handleEditClick}
            onDeleteClick={() => {
              setError(null)
              setIsDeleteOpen(true)
            }}
            onNameChange={setEditedName}
            onDescriptionChange={setEditedDescription}
            onDpiChange={setEditedDpi}
            onPagesChange={setEditedPages}
            onDeskewedChange={setEditedDeskewed}
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
    </div>
  )
}
