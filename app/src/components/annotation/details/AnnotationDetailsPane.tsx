import { useEffect, useState } from 'react'
import { useAppState } from '../../../context/useAppState.ts'
import {
  AnnotationsService,
  ApiError,
  type model_Annotation,
} from '../../../api'
import TimeAgo from 'javascript-time-ago'
import en from 'javascript-time-ago/locale/en'
import { RuleDisplay } from '../../rules/RuleDisplay.tsx'
import { type AnnotationRule } from '../../../utils/rules.ts'
import { useAuthStore } from '../../../store/authStore.ts'
import { useAnnotationsQuery } from '../../../queries/annotations.ts'
import { useDatasetsQuery } from '../../../queries/datasets.ts'
import { DeleteAnnotationModal } from '../../modal/DeleteAnnotationModal.tsx'
import { Button } from '../../core/Button.tsx'
import { getStageDisplayName } from '../../../utils/stages.ts'
import { ExportAnnotationModal } from './ExportAnnotationModal.tsx'

TimeAgo.addDefaultLocale(en)
const timeAgo = new TimeAgo('en-US')

const Timestamp = ({ date }: { date: string | undefined }) => {
  if (!date) {
    return 'N/A'
  }
  const d = new Date(date)
  return (
    <div className="flex gap-2 items-center">
      <span>{timeAgo.format(d)}</span>
      <span className="text-xs text-gray-500">{d.toLocaleString()}</span>
    </div>
  )
}

interface AnnotationDetailsContentProps {
  annotation: model_Annotation
  isEditing: boolean
  editedName: string
  editedDescription: string
  onNameChange: (name: string) => void
  onDescriptionChange: (description: string) => void
  onSave: () => void
  onCancel: () => void
  error?: string | null
}

const AnnotationDetailsContent = ({
  annotation,
  isEditing,
  editedName,
  editedDescription,
  onNameChange,
  onDescriptionChange,
  onSave,
  onCancel,
  error,
}: AnnotationDetailsContentProps) => {
  const { setState } = useAppState()
  const { data: annotations } = useAnnotationsQuery(annotation.dataset_id!)
  const { data: datasets } = useDatasetsQuery()
  const appliedRules = (annotation.applied_rules || []) as AnnotationRule[]

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
            value={editedName}
            onChange={(e) => onNameChange(e.target.value)}
            className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full"
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
            className="text-sm leading-tight break-all border border-gray-300 rounded p-1 w-full"
            rows={3}
          />
        ) : (
          <div className="text-sm leading-tight break-all">
            {annotation.description}
          </div>
        )}
        <div className="font-semibold text-xs opacity-80 pt-0.5">Dataset</div>
        <div className="text-sm leading-tight break-all font-mono">
          {datasets?.find((d) => d.id === annotation.dataset_id)?.name}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Pages</div>
        <div className="text-sm leading-tight break-all">
          {annotation.pages}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Stage</div>
        <div className="text-sm leading-tight break-all">
          {annotation.pipeline_stage &&
            getStageDisplayName(annotation.pipeline_stage)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Segmented</div>
        <div className="text-sm leading-tight break-all">
          {String(!!annotation.segmented)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Ground truth
        </div>
        <div className="text-sm leading-tight break-all">
          {String(!!annotation.ground_truth)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">OCRed</div>
        <div className="text-sm leading-tight break-all">
          {String(!!annotation.ocred)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Origin annotation
        </div>
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
        <div className="font-semibold text-xs opacity-80 pt-0.5">Created</div>
        <div className="text-sm leading-tight break-all ">
          <Timestamp date={annotation.created_at} />
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Updated</div>
        <div className="text-sm leading-tight break-all ">
          <Timestamp date={annotation.updated_at} />
        </div>
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

      {error && <div className="mt-4 text-sm text-red-600">{error}</div>}
      {isEditing && (
        <div className="flex justify-end gap-2 mt-4">
          <Button
            onClick={onCancel}
            className="px-3 py-1.5 text-sm font-semibold"
          >
            Cancel
          </Button>
          <Button
            onClick={onSave}
            variant="primary"
            className="px-3 py-1.5 text-sm font-semibold"
          >
            Save
          </Button>
        </div>
      )}
    </div>
  )
}

export function AnnotationDetailsPane() {
  const { annotation, refetch, setState } = useAppState()
  const [isEditing, setIsEditing] = useState(false)
  const [editedName, setEditedName] = useState('')
  const [editedDescription, setEditedDescription] = useState('')
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const [error, setError] = useState<string | null>(null)
  const [isDeleteOpen, setIsDeleteOpen] = useState(false)
  const [isDeleting, setIsDeleting] = useState(false)
  const [isExportOpen, setIsExportOpen] = useState(false)

  useEffect(() => {
    setError(null)
  }, [annotation?.id])

  const handleEditClick = () => {
    setError(null)
    if (annotation) {
      setEditedName(annotation.name || '')
      setEditedDescription(annotation.description || '')
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
        {isAuthenticated && !isEditing && annotation && (
          <div className="flex items-center gap-2">
            <Button onClick={handleEditClick} className="px-2 py-1 text-xs">
              Edit
            </Button>
            {annotation.ocred && (
              <Button
                onClick={() => setIsExportOpen(true)}
                className="px-2 py-1 text-xs"
              >
                Export
              </Button>
            )}
            <Button
              onClick={handleDeleteClick}
              variant="danger"
              className="px-2 py-1 text-xs"
            >
              Delete
            </Button>
          </div>
        )}
      </div>

      {annotation && (
        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          <AnnotationDetailsContent
            annotation={annotation}
            isEditing={isEditing}
            editedName={editedName}
            editedDescription={editedDescription}
            onNameChange={setEditedName}
            onDescriptionChange={setEditedDescription}
            onSave={handleSave}
            onCancel={handleCancel}
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
    </section>
  )
}
