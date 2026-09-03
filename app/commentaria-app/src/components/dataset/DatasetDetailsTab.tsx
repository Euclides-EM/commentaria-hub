import { type model_Dataset } from '@hub-api'
import { AnnotationActions } from '../annotation/AnnotationActions.tsx'
import { Button } from '../core/Button.tsx'
import { EditionDetailsTable } from '../core/EditionDetailsTable.tsx'
import { ShelfmarkDetailsTable } from '../core/ShelfmarkDetailsTable.tsx'
import { ErrorMessage } from '../core/ErrorMessage.tsx'
import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { Timestamp } from '../core/Timestamp.tsx'
import { formatBoolean } from '../../utils/formatBoolean.tsx'

interface DatasetDetailsTabProps {
  dataset: model_Dataset
  editionId: string | null
  datasetStatusLabel: string
  isCreating: boolean
  isEditing: boolean
  isAuthenticated: boolean
  isSaving: boolean
  editedName: string
  editedDescription: string
  editedDpi: string
  editedPages: string
  editedDeskewed: boolean
  editedDenoised: boolean
  error: string | null
  isUploadingImages: boolean
  onEditClick: () => void
  onDeleteClick: () => void
  onUploadImagesClick: () => void
  onNameChange: (value: string) => void
  onDescriptionChange: (value: string) => void
  onDpiChange: (value: string) => void
  onPagesChange: (value: string) => void
  onDeskewedChange: (value: boolean) => void
  onDenoisedChange: (value: boolean) => void
  onCancel: () => void
  onSave: () => void
}

export function DatasetDetailsTab({
  dataset,
  editionId,
  datasetStatusLabel,
  isCreating,
  isEditing,
  isAuthenticated,
  isSaving,
  editedName,
  editedDescription,
  editedDpi,
  editedPages,
  editedDeskewed,
  editedDenoised,
  error,
  isUploadingImages,
  onEditClick,
  onDeleteClick,
  onUploadImagesClick,
  onNameChange,
  onDescriptionChange,
  onDpiChange,
  onPagesChange,
  onDeskewedChange,
  onDenoisedChange,
  onCancel,
  onSave,
}: DatasetDetailsTabProps) {
  return (
    <>
      <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col bg-white m-3 mb-0 w-[calc(100%-1.5rem)] max-w-[80vw] mx-auto">
        <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
          <div>Dataset Details</div>
          {isAuthenticated && (
            <div className="flex items-center gap-2">
              {isEditing ? (
                <>
                  <Button
                    onClick={onCancel}
                    className="px-2 py-1 text-xs"
                    disabled={isSaving}
                  >
                    Cancel
                  </Button>
                  <Button
                    onClick={onSave}
                    variant="primary"
                    className="px-2 py-1 text-xs"
                    disabled={isSaving || isCreating}
                  >
                    {isSaving ? 'Saving...' : 'Save'}
                  </Button>
                </>
              ) : (
                <>
                  <Button
                    onClick={onUploadImagesClick}
                    className="px-2 py-1 text-xs"
                    disabled={isCreating || isUploadingImages}
                  >
                    {isUploadingImages ? 'Uploading...' : 'Upload images'}
                  </Button>
                  <Button
                    onClick={onEditClick}
                    className="px-2 py-1 text-xs"
                    disabled={isCreating}
                  >
                    Edit
                  </Button>
                  <Button
                    onClick={onDeleteClick}
                    variant="danger"
                    className="px-2 py-1 text-xs"
                    disabled={isCreating}
                  >
                    Delete
                  </Button>
                </>
              )}
            </div>
          )}
        </div>

        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          <div className="mt-2.5 border border-gray-200 rounded-lg bg-gray-50 p-3.5 overflow-auto leading-normal text-base box-border">
            {isCreating && (
              <div className="mb-4">
                <LoadingSpinner
                  size="sm"
                  message="Dataset is being created..."
                />
                <p className="mt-2 text-sm text-gray-600">
                  You can select another dataset from the navbar. This one will
                  be ready once creation finishes.
                </p>
              </div>
            )}
            <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 items-start">
              <div className="font-semibold text-xs opacity-80 pt-0.5">ID</div>
              <div className="text-sm leading-tight break-all font-mono">
                {dataset.id}
              </div>

              <div className="font-semibold text-xs opacity-80 pt-0.5">
                Name
              </div>
              {isEditing ? (
                <input
                  type="text"
                  autoComplete="on"
                  value={editedName}
                  onChange={(e) => onNameChange(e.target.value)}
                  className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                  disabled={isSaving}
                />
              ) : (
                <div className="text-sm leading-tight break-all">
                  {dataset.name || 'N/A'}
                </div>
              )}

              <div className="font-semibold text-xs opacity-80 pt-0.5">
                Description
              </div>
              {isEditing ? (
                <textarea
                  value={editedDescription}
                  onChange={(e) => onDescriptionChange(e.target.value)}
                  className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                  rows={3}
                  disabled={isSaving}
                />
              ) : (
                <div className="text-sm leading-tight whitespace-pre-wrap break-words">
                  {dataset.description?.replace(/\\n/g, '\n') || 'N/A'}
                </div>
              )}

              <div className="font-semibold text-xs opacity-80 pt-0.5">
                Facsimile
              </div>
              <div className="text-sm leading-tight break-all font-mono">
                {dataset.facsimile_id
                  ? dataset.edition_id
                    ? `${dataset.edition_id} (${dataset.facsimile_id})`
                    : dataset.facsimile_id
                  : 'N/A'}
              </div>

              <div className="font-semibold text-xs opacity-80 pt-0.5">DPI</div>
              {isEditing ? (
                <input
                  type="number"
                  min={1}
                  value={editedDpi}
                  onChange={(e) => onDpiChange(e.target.value)}
                  className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                  disabled={isSaving}
                />
              ) : (
                <div className="text-sm leading-tight break-all">
                  {dataset.dpi ?? 'N/A'}
                </div>
              )}

              <div className="font-semibold text-xs opacity-80 pt-0.5">
                Pages
              </div>
              {isEditing ? (
                <input
                  type="text"
                  autoComplete="off"
                  value={editedPages}
                  onChange={(e) => onPagesChange(e.target.value)}
                  className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
                  disabled={isSaving}
                />
              ) : (
                <div className="text-sm leading-tight break-all">
                  {dataset.pages || 'All'}
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
                    onChange={(e) => onDeskewedChange(e.target.checked)}
                    className="h-4 w-4"
                    disabled={isSaving}
                  />
                  {formatBoolean(editedDeskewed)}
                </label>
              ) : (
                <div className="text-sm leading-tight break-all">
                  {formatBoolean(dataset.deskewed)}
                </div>
              )}

              <div className="font-semibold text-xs opacity-80 pt-0.5">
                Denoised
              </div>
              {isEditing ? (
                <label className="flex items-center gap-2 text-sm leading-tight">
                  <input
                    type="checkbox"
                    checked={editedDenoised}
                    onChange={(e) => onDenoisedChange(e.target.checked)}
                    className="h-4 w-4"
                    disabled={isSaving}
                  />
                  {formatBoolean(editedDenoised)}
                </label>
              ) : (
                <div className="text-sm leading-tight break-all">
                  {formatBoolean(dataset.denoised)}
                </div>
              )}

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
                {dataset.creation_error || 'None'}
              </div>

              <div className="font-semibold text-xs opacity-80 pt-0.5">
                Created
              </div>
              <div className="text-sm leading-tight break-all">
                <Timestamp date={dataset.created_at} />
              </div>

              <div className="font-semibold text-xs opacity-80 pt-0.5">
                Updated
              </div>
              <div className="text-sm leading-tight break-all">
                <Timestamp date={dataset.updated_at} />
              </div>

              {editionId && (
                <>
                  <div className="font-semibold text-xs opacity-80 pt-0.5">
                    Edition ID
                  </div>
                  <div className="text-sm leading-tight break-all font-mono">
                    {editionId}
                  </div>

                  <div className="font-semibold text-xs opacity-80 pt-0.5">
                    Edition
                  </div>
                  <div className="text-sm leading-tight break-all">
                    <EditionDetailsTable editionId={editionId} />
                  </div>

                  <div className="font-semibold text-xs opacity-80 pt-0.5">
                    Shelfmark
                  </div>
                  <div className="text-sm leading-tight break-all">
                    <ShelfmarkDetailsTable
                      editionId={editionId}
                      facsimileId={dataset.facsimile_id}
                    />
                  </div>
                </>
              )}
            </div>
            <div className="mt-4">
              <ErrorMessage message={error} />
            </div>
          </div>
        </div>
      </section>

      <AnnotationActions dataSetId={dataset.id || ''} />
    </>
  )
}
