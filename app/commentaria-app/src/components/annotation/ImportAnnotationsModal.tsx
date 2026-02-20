import { type FormEvent, useEffect, useMemo, useState } from 'react'
import { AnnotationsService, ApiError } from '../../api'
import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { Button } from '../core/Button.tsx'
import { FileUpload } from '../core/FileUpload'
import Select from 'react-select'
import { selectStyles } from '../../styles/selectStyles.ts'
import { ErrorMessage } from '../core/ErrorMessage'
import { useAnnotationsQuery } from '../../queries/annotations'
import { useAppState } from '../../context/useAppState'

type ImportMode = 'url' | 'zip'

interface ImportAnnotationsModalProps {
  isOpen: boolean
  mode: ImportMode
  dataSetId: string
  onClose: () => void
  onImported?: (annotationId: string) => void
}

const formatOptions = [
  { value: 'ALTO', label: 'ALTO' },
  { value: 'YOLO', label: 'YOLO' },
]

export function ImportAnnotationsModal({
  isOpen,
  mode,
  dataSetId,
  onClose,
  onImported,
}: ImportAnnotationsModalProps) {
  const { setState, refetch } = useAppState()
  const [format, setFormat] = useState<'ALTO' | 'YOLO'>('ALTO')
  const [url, setUrl] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [segmented, setSegmented] = useState(false)
  const [ocred, setOcred] = useState(false)
  const [groundTruth, setGroundTruth] = useState(false)
  const [originAnnotationId, setOriginAnnotationId] = useState<string | null>(
    null,
  )
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
      setFormat('ALTO')
      setUrl('')
      setFile(null)
      setName('')
      setDescription('')
      setSegmented(false)
      setOcred(false)
      setGroundTruth(false)
      setOriginAnnotationId(null)
      setError(null)
      setLoading(false)
    }
  }, [isOpen])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    try {
      setError(null)
      if (mode === 'url' && !url.trim()) {
        setError('Please provide a URL.')
        return
      }
      if (mode === 'zip' && !file) {
        setError('Please choose a ZIP file.')
        return
      }

      setLoading(true)
      if (mode === 'url') {
        const annotation =
          await AnnotationsService.postDatasetsAnnotationsFromurl({
            dataSetId,
            format,
            url: url.trim(),
            name: name.trim() || undefined,
            description: description.trim() || undefined,
            segmented,
            ocred,
            groundTruth,
            originAnnotationId: originAnnotationId || undefined,
          })
        setState({ annotationId: annotation.id! })
        refetch()
        onImported?.(annotation.id!)
      } else {
        const zipFile = file
        if (!zipFile) {
          setError('Please choose a ZIP file.')
          return
        }
        const annotation =
          await AnnotationsService.postDatasetsAnnotationsFromzip({
            dataSetId,
            format,
            file: zipFile,
            name: name.trim() || undefined,
            description: description.trim() || undefined,
            segmented,
            ocred,
            groundTruth,
            originAnnotationId: originAnnotationId || undefined,
          })
        setState({ annotationId: annotation.id! })
        refetch()
        onImported?.(annotation.id!)
      }
      onClose()
    } catch (e) {
      console.error('Failed to import annotation:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    } finally {
      setLoading(false)
    }
  }

  if (!isOpen) {
    return null
  }

  const title = mode === 'url' ? 'Import from URL' : 'Import from ZIP'

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
          <h2 className="text-lg font-semibold">{title}</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Format
            </label>
            <Select
              options={formatOptions}
              value={formatOptions.find((option) => option.value === format)}
              onChange={(option) =>
                setFormat((option?.value as 'ALTO' | 'YOLO') || 'ALTO')
              }
              isDisabled={loading}
              className="text-sm"
              styles={selectStyles<{ value: string; label: string }>({
                controlWidth: 160,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
            />
          </div>

          {mode === 'url' ? (
            <div className="space-y-2">
              <label className="block text-sm font-medium text-gray-700">
                File URL
              </label>
              <input
                type="url"
                autoComplete="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded-md"
                placeholder="https://example.com/annotations.zip"
                disabled={loading}
                required
              />
            </div>
          ) : (
            <div className="space-y-2">
              <label className="block text-sm font-medium text-gray-700">
                ZIP file
              </label>
              <FileUpload
                file={file}
                onFileChange={setFile}
                accept=".zip"
                disabled={loading}
                required
              />
            </div>
          )}

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Annotation name
            </label>
            <input
              type="text"
              autoComplete="on"
              value={name}
              required
              onChange={(e) => setName(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              disabled={loading}
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Origin annotation (optional)
            </label>
            <Select
              value={
                originAnnotationOptions.find(
                  (option) => option.value === originAnnotationId,
                ) || null
              }
              onChange={(option: { value: string; label: string } | null) =>
                setOriginAnnotationId(option?.value || null)
              }
              options={originAnnotationOptions}
              placeholder="Select origin annotation..."
              isLoading={annotationsLoading}
              isDisabled={loading}
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
              Description (optional)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              rows={3}
              disabled={loading}
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={segmented}
                onChange={(e) => setSegmented(e.target.checked)}
                className="h-4 w-4"
                disabled={loading}
              />
              Segmented
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={ocred}
                onChange={(e) => setOcred(e.target.checked)}
                className="h-4 w-4"
                disabled={loading}
              />
              OCRed
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input
                type="checkbox"
                checked={groundTruth}
                onChange={(e) => setGroundTruth(e.target.checked)}
                className="h-4 w-4"
                disabled={loading}
              />
              Ground truth
            </label>
          </div>

          <ErrorMessage message={error} />
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          {loading ? (
            <LoadingSpinner size="sm" message="Importing annotations..." />
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
                Import
              </Button>
            </>
          )}
        </div>
      </form>
    </div>
  )
}
