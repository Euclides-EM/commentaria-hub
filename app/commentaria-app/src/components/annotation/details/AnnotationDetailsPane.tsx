import { useEffect, useMemo, useState } from 'react'
import { useAppState } from '../../../context/useAppState.ts'
import {
  AnnotationsService,
  ApiError,
  type annotation_Annotation,
} from '@hub-api'
import Select from 'react-select'
import { Timestamp } from '../../core/Timestamp'
import { RuleDisplay } from '../../rules/RuleDisplay.tsx'
import { type AnnotationRule } from '../../../utils/rules.ts'
import { countPages } from '../../../utils/pages.ts'
import { useAuthStore } from '../../../store/authStore.ts'
import {
  useAnnotationCategories,
  useAnnotationsQuery,
} from '../../../queries/annotations.ts'
import {
  useDatasetImageKeysQuery,
  useDatasetsQuery,
} from '../../../queries/datasets.ts'
import { DeleteAnnotationModal } from '../../modal/DeleteAnnotationModal.tsx'
import { Button } from '../../core/Button.tsx'
import { getStageDisplayName } from '../../../utils/stages.ts'
import { ExportAnnotationModal } from './ExportAnnotationModal.tsx'
import { ErrorMessage } from '../../core/ErrorMessage'
import { selectStyles } from '../../../styles/selectStyles'
import { CreateAnnotationModal } from '../CreateAnnotationModal.tsx'
import { useRunningIntegrationJobsQuery } from '../../../queries/integrations.ts'
import { EditionDetailsTable } from '../../core/EditionDetailsTable.tsx'

interface AnnotationDetailsContentProps {
  annotation: annotation_Annotation
  imageKeysCount: number
  imageKeysLoading: boolean
  isExporting: boolean
  isEditing: boolean
  editedName: string
  editedDescription: string
  editedOriginAnnotationId: string | null
  editedGroundTruth: boolean
  editedHidden: boolean
  onNameChange: (name: string) => void
  onDescriptionChange: (description: string) => void
  onOriginAnnotationChange: (originAnnotationId: string | null) => void
  onGroundTruthChange: (groundTruth: boolean) => void
  onHiddenChange: (hidden: boolean) => void
  error?: string | null
}

const formatBoolean = (value: boolean | null | undefined) =>
  value ? <span className="text-teal-700 font-semibold">{'\u2713'}</span> : '-'

const AnnotationDetailsContent = ({
  annotation,
  imageKeysCount,
  imageKeysLoading,
  isExporting,
  isEditing,
  editedName,
  editedDescription,
  editedOriginAnnotationId,
  editedGroundTruth,
  editedHidden,
  onNameChange,
  onDescriptionChange,
  onOriginAnnotationChange,
  onGroundTruthChange,
  onHiddenChange,
  error,
}: AnnotationDetailsContentProps) => {
  const { setState } = useAppState()
  const { data: annotations } = useAnnotationsQuery(annotation.dataset_id!)
  const { data: datasets } = useDatasetsQuery()
  const { data: categories, isLoading: categoriesLoading } =
    useAnnotationCategories(annotation.dataset_id!, annotation.id!)
  const datasetForAnnotation =
    datasets?.find((d) => d.id === annotation.dataset_id) || null
  const editionId = datasetForAnnotation?.edition_id || null
  const hasEdition = !!editionId
  const appliedRules = (annotation.applied_rules || []) as AnnotationRule[]

  const originAnnotationOptions = useMemo(() => {
    return (
      annotations
        ?.filter((a) => a.id && a.id !== annotation.id)
        .map((a) => ({
          value: a.id as string,
          label: a.name || (a.id as string),
        })) || []
    )
  }, [annotation.id, annotations])

  const hasPages = annotation.pages != null && annotation.pages !== ''

  const originAnnotation = annotations?.find(
    (a) => a.id === annotation.origin_annotation_id,
  )
  return (
    <div className="mt-2.5 border border-gray-200 rounded-lg bg-gray-50 p-3.5 overflow-auto leading-normal text-base box-border">
      <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 items-start">
        <div className="font-semibold text-xs opacity-80 pt-0.5">ID</div>
        <div className="text-sm leading-tight break-all font-mono">
          {annotation.id}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Name</div>
        {isEditing ? (
          <input
            type="text"
            autoComplete="on"
            value={editedName}
            onChange={(e) => onNameChange(e.target.value)}
            className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full bg-white"
          />
        ) : (
          <div className="text-sm leading-tight break-all">
            {annotation.name}
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
          />
        ) : (
          <div className="text-sm leading-tight whitespace-pre-wrap break-words">
            {annotation.description?.replace(/\\n/g, '\n')}
          </div>
        )}
        <div className="font-semibold text-xs opacity-80 pt-0.5">Dataset</div>
        <div className="text-sm leading-tight break-all font-mono">
          {datasetForAnnotation?.name || 'N/A'}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          {hasPages ? 'Pages' : 'Keys'}
        </div>
        <div className="text-sm leading-tight break-all">
          {hasPages
            ? annotation.pages || ''
            : imageKeysLoading
              ? 'Loading…'
              : imageKeysCount}
          {hasPages && (
            <span className="text-gray-600 ml-1">
              (total: {countPages(annotation.pages || '')})
            </span>
          )}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Ground truth
        </div>
        {isEditing ? (
          <label className="flex items-center gap-2 text-sm leading-tight">
            <input
              type="checkbox"
              checked={editedGroundTruth}
              onChange={(e) => onGroundTruthChange(e.target.checked)}
              className="h-4 w-4"
            />
            {formatBoolean(editedGroundTruth)}
          </label>
        ) : (
          <div className="text-sm leading-tight break-all">
            {formatBoolean(annotation.ground_truth)}
          </div>
        )}
        <div className="font-semibold text-xs opacity-80 pt-0.5">Stage</div>
        <div className="text-sm leading-tight break-all">
          {annotation.pipeline_stage &&
            getStageDisplayName(annotation.pipeline_stage)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Segmented</div>
        <div className="text-sm leading-tight break-all">
          {formatBoolean(annotation.segmented)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Lines detected
        </div>
        <div className="text-sm leading-tight break-all">
          {formatBoolean(annotation.lines_detected)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">OCRed</div>
        <div className="text-sm leading-tight break-all">
          {formatBoolean(annotation.ocred)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Hidden</div>
        {isEditing ? (
          <label className="flex items-center gap-2 text-sm leading-tight">
            <input
              type="checkbox"
              checked={editedHidden}
              onChange={(e) => onHiddenChange(e.target.checked)}
              className="h-4 w-4"
            />
            {formatBoolean(editedHidden)}
          </label>
        ) : (
          <div className="text-sm leading-tight break-all">
            {formatBoolean(annotation.hidden)}
          </div>
        )}
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Categories
        </div>
        <div className="text-sm leading-tight break-all flex flex-wrap items-center gap-2">
          {categoriesLoading ? (
            <span className="text-gray-500">Loading…</span>
          ) : categories && categories.length > 0 ? (
            <>
              <div className="flex flex-wrap gap-1">
                {[...categories]
                  .sort((a, b) => a.localeCompare(b))
                  .map((cat) => (
                    <span
                      key={cat}
                      className="inline-flex items-center rounded bg-gray-200 px-1.5 py-0.5 text-xs font-medium text-gray-800"
                    >
                      {cat}
                    </span>
                  ))}
              </div>
              <Button
                type="button"
                onClick={() => {
                  const sorted = [...categories].sort((a, b) =>
                    a.localeCompare(b),
                  )
                  void navigator.clipboard.writeText(sorted.join(', '))
                }}
                className="px-2 py-1 text-xs"
              >
                Copy
              </Button>
            </>
          ) : (
            <span className="text-gray-500">None</span>
          )}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Origin annotation
        </div>
        {isEditing ? (
          <div className="text-sm leading-tight break-all">
            <Select
              value={
                originAnnotationOptions.find(
                  (option) => option.value === editedOriginAnnotationId,
                ) || null
              }
              onChange={(option: { value: string; label: string } | null) =>
                onOriginAnnotationChange(option?.value || null)
              }
              options={originAnnotationOptions}
              placeholder="Select origin annotation..."
              styles={selectStyles<{ value: string; label: string }>({
                controlWidth: 260,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          </div>
        ) : (
          <div className="text-sm leading-tight break-all">
            {originAnnotation && (
              <button
                className="underline text-gray-800 hover:text-gray-500"
                onClick={() => {
                  setState({ annotationId: originAnnotation.id })
                }}
              >
                {originAnnotation.name}
              </button>
            )}
          </div>
        )}
        <div className="font-semibold text-xs opacity-80 pt-0.5">Created</div>
        <div className="text-sm leading-tight break-all ">
          <Timestamp date={annotation.created_at} />
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Updated</div>
        <div className="text-sm leading-tight break-all ">
          <Timestamp date={annotation.updated_at} />
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Export</div>
        <div className="text-sm leading-tight break-all">
          {isExporting ? 'In progress' : 'Idle'}
        </div>
        {hasEdition && (
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
          </>
        )}
      </div>

      <div className="font-semibold text-xs opacity-80 py-2">Applied rules</div>
      <div className="text-sm leading-tight">
        {appliedRules.length > 0 ? (
          <div className="space-y-2">
            {appliedRules.map((rule, index) => (
              <RuleDisplay key={index} rule={rule} />
            ))}
          </div>
        ) : (
          <span className="text-gray-500">None</span>
        )}
      </div>

      <div className="mt-4">
        <ErrorMessage message={error} />
      </div>
    </div>
  )
}

export function AnnotationDetailsPane() {
  const { annotation, refetch, setState } = useAppState()
  const [isEditing, setIsEditing] = useState(false)
  const [editedName, setEditedName] = useState('')
  const [editedDescription, setEditedDescription] = useState('')
  const [editedOriginAnnotationId, setEditedOriginAnnotationId] = useState<
    string | null
  >(null)
  const [editedGroundTruth, setEditedGroundTruth] = useState(false)
  const [editedHidden, setEditedHidden] = useState(false)
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const [error, setError] = useState<string | null>(null)
  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [isExportOpen, setIsExportOpen] = useState(false)
  const [isDuplicateOpen, setIsDuplicateOpen] = useState(false)
  const { data: runningJobs } = useRunningIntegrationJobsQuery()
  const shouldLoadImageKeys = !!annotation && !annotation.pages
  const { data: imageKeys = [], isLoading: imageKeysLoading } =
    useDatasetImageKeysQuery(annotation?.dataset_id || '', shouldLoadImageKeys)
  const isExporting = !!runningJobs?.some(
    (job) =>
      job.annotation?.dataset_id === annotation?.dataset_id &&
      job.annotation?.id === annotation?.id,
  )

  useEffect(() => {
    setError(null)
  }, [annotation?.id])

  const handleEditClick = () => {
    setError(null)
    if (annotation) {
      setEditedName(annotation.name || '')
      setEditedDescription(annotation.description || '')
      setEditedOriginAnnotationId(annotation.origin_annotation_id || null)
      setEditedGroundTruth(!!annotation.ground_truth)
      setEditedHidden(!!annotation.hidden)
      setIsEditing(true)
    }
  }

  const handleSave = async () => {
    if (!annotation) {
      return
    }

    try {
      await AnnotationsService.putDatasetsAnnotations({
        dataSetId: annotation.dataset_id!,
        id: annotation.id!,
        annotation: {
          ...annotation,
          name: editedName,
          description: editedDescription,
          origin_annotation_id: editedOriginAnnotationId || undefined,
          ground_truth: editedGroundTruth,
          hidden: editedHidden,
        },
      })
      refetch()
      setIsEditing(false)
    } catch (e) {
      console.error('Failed to update annotation:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    }
  }

  const handleCancel = () => {
    setIsEditing(false)
    if (annotation) {
      setEditedName(annotation.name || '')
      setEditedDescription(annotation.description || '')
      setEditedOriginAnnotationId(annotation.origin_annotation_id || null)
      setEditedGroundTruth(!!annotation.ground_truth)
      setEditedHidden(!!annotation.hidden)
    }
  }

  const handleDeleteClick = () => {
    if (!annotation) {
      return
    }
    setError(null)
    setIsDeleteOpen(true)
  }

  const handleDeleteConfirm = async () => {
    if (!annotation) {
      return
    }
    try {
      setError(null)
      setIsDeleting(true)
      await AnnotationsService.deleteDatasetsAnnotations({
        dataSetId: annotation.dataset_id!,
        id: annotation.id!,
      })
      setIsDeleteOpen(false)
      setState({ annotationId: '' })
      refetch()
    } catch (e) {
      console.error('Failed to delete annotation:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white m-3 mb-0">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Annotation Details</div>
        {isAuthenticated && annotation && (
          <div className="flex items-center gap-2">
            {isEditing ? (
              <>
                <Button onClick={handleCancel} className="px-2 py-1 text-xs">
                  Cancel
                </Button>
                <Button
                  onClick={handleSave}
                  variant="primary"
                  className="px-2 py-1 text-xs"
                >
                  Save
                </Button>
              </>
            ) : (
              <>
                <Button onClick={handleEditClick} className="px-2 py-1 text-xs">
                  Edit
                </Button>
                <Button
                  onClick={() => setIsDuplicateOpen(true)}
                  className="px-2 py-1 text-xs"
                >
                  Duplicate
                </Button>
                <Button
                  onClick={() => setIsExportOpen(true)}
                  className="px-2 py-1 text-xs"
                >
                  {isExporting ? 'Exporting…' : 'Export'}
                </Button>
                <Button
                  onClick={handleDeleteClick}
                  variant="danger"
                  className="px-2 py-1 text-xs"
                >
                  Delete
                </Button>
              </>
            )}
          </div>
        )}
      </div>

      {annotation && (
        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          <AnnotationDetailsContent
            annotation={annotation}
            imageKeysCount={imageKeys.length}
            imageKeysLoading={imageKeysLoading}
            isExporting={isExporting}
            isEditing={isEditing}
            editedName={editedName}
            editedDescription={editedDescription}
            editedOriginAnnotationId={editedOriginAnnotationId}
            editedGroundTruth={editedGroundTruth}
            editedHidden={editedHidden}
            onNameChange={setEditedName}
            onDescriptionChange={setEditedDescription}
            onOriginAnnotationChange={setEditedOriginAnnotationId}
            onGroundTruthChange={setEditedGroundTruth}
            onHiddenChange={setEditedHidden}
            error={error}
          />
        </div>
      )}

      {annotation && (
        <DeleteAnnotationModal
          isOpen={isDeleteOpen}
          annotationLabel={annotation.name || annotation.id || ''}
          error={error}
          isDeleting={isDeleting}
          onCancel={() => setIsDeleteOpen(false)}
          onConfirm={handleDeleteConfirm}
        />
      )}
      {annotation && annotation.dataset_id && (
        <ExportAnnotationModal
          isOpen={isExportOpen}
          onClose={() => setIsExportOpen(false)}
        />
      )}
      {annotation && annotation.dataset_id && (
        <CreateAnnotationModal
          isOpen={isDuplicateOpen}
          mode="duplicate"
          dataSetId={annotation.dataset_id}
          initialOriginAnnotationId={annotation.id || null}
          initialName={`${annotation.name || annotation.id} (copy)`}
          initialDescription={annotation.description || ''}
          onClose={() => setIsDuplicateOpen(false)}
          onCreated={(annotationId) => {
            setState({ annotationId })
            refetch()
          }}
        />
      )}
    </section>
  )
}
