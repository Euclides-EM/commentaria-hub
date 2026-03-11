import { type FormEvent, useEffect, useMemo, useState } from 'react'
import { AnnotationsService, ApiError } from '@hub-api'
import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { Button } from '../core/Button.tsx'
import Select from 'react-select'
import { selectStyles } from '../../styles/selectStyles.ts'
import { ErrorMessage } from '../core/ErrorMessage'
import { useAnnotationsQuery } from '../../queries/annotations'
import { useAppState } from '../../context/useAppState'

interface CreateAnnotationModalProps {
  isOpen: boolean
  dataSetId: string
  mode?: 'create' | 'duplicate'
  initialOriginAnnotationId?: string | null
  initialName?: string
  initialDescription?: string
  initialGroundTruth?: boolean
  onClose: () => void
  onCreated?: (annotationId: string) => void
}

export function CreateAnnotationModal({
  isOpen,
  dataSetId,
  mode = 'create',
  initialOriginAnnotationId = null,
  initialName = '',
  initialDescription = '',
  initialGroundTruth = false,
  onClose,
  onCreated,
}: CreateAnnotationModalProps) {
  const isDuplicateMode = mode === 'duplicate'
  const { setState, refetch } = useAppState()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [groundTruth, setGroundTruth] = useState(false)
  const [hidden, setHidden] = useState(false)
  const [copyFeatureResults, setCopyFeatureResults] = useState(true)
  const [originAnnotationId, setOriginAnnotationId] = useState<string | null>(
    null,
  )
  const [nameTouched, setNameTouched] = useState(false)
  const [descriptionTouched, setDescriptionTouched] = useState(false)
  const [groundTruthTouched, setGroundTruthTouched] = useState(false)
  const [hiddenTouched, setHiddenTouched] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const { data: annotations, isLoading: annotationsLoading } =
    useAnnotationsQuery(dataSetId)

  const originAnnotationOptions = useMemo(() => {
    if (!annotations) return []
    return annotations
      .filter((annotation) => !!annotation.id)
      .map((annotation) => ({
        value: annotation.id as string,
        label: annotation.name || (annotation.id as string),
      }))
  }, [annotations])

  useEffect(() => {
    if (isOpen) {
      setName(initialName)
      setDescription(initialDescription)
      setGroundTruth(initialGroundTruth)
      setHidden(false)
      setCopyFeatureResults(true)
      setOriginAnnotationId(initialOriginAnnotationId)
      setNameTouched(false)
      setDescriptionTouched(false)
      setGroundTruthTouched(false)
      setHiddenTouched(false)
      setError(null)
      setLoading(false)
    }
  }, [
    initialDescription,
    initialGroundTruth,
    initialName,
    initialOriginAnnotationId,
    isOpen,
  ])

  const originAnnotation = useMemo(
    () =>
      annotations?.find((annotation) => annotation.id === originAnnotationId) ||
      null,
    [annotations, originAnnotationId],
  )

  useEffect(() => {
    if (!originAnnotation) return
    if (!nameTouched && !name.trim()) {
      const baseName =
        originAnnotation.name || originAnnotation.id || 'Annotation'
      setName(`${baseName} (copy)`)
    }
    if (!descriptionTouched && !description.trim()) {
      setDescription(originAnnotation.description || '')
    }
    if (!groundTruthTouched) {
      setGroundTruth(Boolean(originAnnotation.ground_truth))
    }
    if (!hiddenTouched) {
      setHidden(Boolean(originAnnotation.hidden))
    }
  }, [
    originAnnotation,
    nameTouched,
    descriptionTouched,
    groundTruthTouched,
    hiddenTouched,
    name,
    description,
  ])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    if (!name.trim()) {
      setError('Please provide a name.')
      return
    }
    if (isDuplicateMode && !originAnnotationId) {
      setError('Please select a source annotation.')
      return
    }
    try {
      setError(null)
      setLoading(true)
      const annotation = isDuplicateMode
        ? await AnnotationsService.postDatasetsAnnotationsDuplicate({
            dataSetId,
            annotationDuplicateRequest: {
              name: name.trim(),
              description: description.trim() || undefined,
              source_annotation_id: originAnnotationId || undefined,
              copy_feature_results:
                originAnnotationId != null ? copyFeatureResults : false,
            },
          })
        : await AnnotationsService.postDatasetsAnnotations({
            dataSetId,
            annotation: {
              name: name.trim(),
              description: description.trim() || undefined,
              ground_truth: groundTruth,
              hidden,
              origin_annotation_id: originAnnotationId || undefined,
            },
            copyFeatureResults:
              originAnnotationId != null ? copyFeatureResults : false,
          })
      setState({ annotationId: annotation.id! })
      refetch()
      onCreated?.(annotation.id!)
      onClose()
    } catch (e) {
      console.error('Failed to create annotation:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    } finally {
      setLoading(false)
    }
  }

  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={loading ? undefined : onClose}
    >
      <form
        className="bg-white rounded-lg max-w-xl w-full max-h-[85vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">
            {isDuplicateMode ? 'Duplicate annotation' : 'Create annotation'}
          </h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Annotation name
            </label>
            <input
              type="text"
              autoComplete="on"
              value={name}
              required
              onChange={(e) => {
                setNameTouched(true)
                setName(e.target.value)
              }}
              className="w-full p-2 border border-gray-300 rounded-md"
              disabled={loading}
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              {isDuplicateMode
                ? 'Source annotation'
                : 'Origin annotation (optional)'}
            </label>
            <Select
              value={
                originAnnotationOptions.find(
                  (option) => option.value === originAnnotationId,
                ) || null
              }
              onChange={(option: { value: string; label: string } | null) => {
                setOriginAnnotationId(option?.value || null)
                if (!option) {
                  setCopyFeatureResults(true)
                }
              }}
              options={originAnnotationOptions}
              placeholder="Select origin annotation..."
              isLoading={annotationsLoading}
              isDisabled={loading}
              styles={selectStyles<{ value: string; label: string }>({
                controlWidth: 260,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable={!isDuplicateMode}
            />
          </div>

          {originAnnotationId && (
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={copyFeatureResults}
                onChange={(e) => setCopyFeatureResults(e.target.checked)}
                className="h-4 w-4"
                disabled={loading}
              />
              Copy feature results
            </label>
          )}

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Description (optional)
            </label>
            <textarea
              value={description}
              onChange={(e) => {
                setDescriptionTouched(true)
                setDescription(e.target.value)
              }}
              className="w-full p-2 border border-gray-300 rounded-md"
              rows={3}
              disabled={loading}
            />
          </div>

          {!isDuplicateMode && (
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={groundTruth}
                  onChange={(e) => {
                    setGroundTruthTouched(true)
                    setGroundTruth(e.target.checked)
                  }}
                  className="h-4 w-4"
                  disabled={loading}
                />
                Ground truth
              </label>
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={hidden}
                  onChange={(e) => {
                    setHiddenTouched(true)
                    setHidden(e.target.checked)
                  }}
                  className="h-4 w-4"
                  disabled={loading}
                />
                Hidden
              </label>
            </div>
          )}

          <ErrorMessage message={error} />
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          {loading ? (
            <LoadingSpinner
              size="sm"
              message={
                isDuplicateMode
                  ? 'Duplicating annotation...'
                  : 'Creating annotation...'
              }
            />
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
                {isDuplicateMode ? 'Duplicate' : 'Create'}
              </Button>
            </>
          )}
        </div>
      </form>
    </div>
  )
}
