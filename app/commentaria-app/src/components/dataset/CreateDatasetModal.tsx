import { type FormEvent, useEffect, useMemo, useState } from 'react'
import { DatasetsService, FacsimilesService, ApiError } from '@hub-api'
import { useQuery } from '@tanstack/react-query'
import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { Button } from '../core/Button.tsx'
import Select from 'react-select'
import { selectStyles } from '../../styles/selectStyles.ts'
import { ErrorMessage } from '../core/ErrorMessage'
import { useAppState } from '../../context/useAppState'

const DEFAULT_DPI = 300

interface CreateDatasetModalProps {
  isOpen: boolean
  onClose: () => void
  onCreated?: (datasetId: string) => void
}

export function CreateDatasetModal({
  isOpen,
  onClose,
  onCreated,
}: CreateDatasetModalProps) {
  const { setState, refetch } = useAppState()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [dpi, setDpi] = useState(DEFAULT_DPI)
  const [facsimileId, setFacsimileId] = useState<string | null>(null)
  const [pages, setPages] = useState('')
  const [deskewed, setDeskewed] = useState(true)
  const [asyncCreate, setAsyncCreate] = useState(true)
  const [createDefaultAnnotation, setCreateDefaultAnnotation] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [facsimileCopied, setFacsimileCopied] = useState(false)

  const { data: facsimiles, isLoading: facsimilesLoading } = useQuery({
    queryKey: ['facsimiles'],
    queryFn: () => FacsimilesService.getFacsimilies({}),
    enabled: isOpen,
  })

  const facsimileOptions = useMemo(() => {
    if (!facsimiles) {
      return []
    }
    return facsimiles
      .filter((f) => f.id != null)
      .map((f) => ({
        value: `${f.id}/${f.edition_id}`,
        label: `${f.edition_id ?? ''}${f.id && ` (${f.id})`}`,
      }))
  }, [facsimiles])

  const selectedFacsimileLabel =
    facsimileOptions.find((o) => o.value === facsimileId)?.label ?? null

  const handleCopyFacsimile = () => {
    if (!selectedFacsimileLabel) return
    void navigator.clipboard.writeText(selectedFacsimileLabel).then(() => {
      setFacsimileCopied(true)
      setTimeout(() => setFacsimileCopied(false), 2000)
    })
  }

  useEffect(() => {
    if (isOpen) {
      setName('')
      setDescription('')
      setDpi(DEFAULT_DPI)
      setFacsimileId(null)
      setFacsimileCopied(false)
      setPages('')
      setDeskewed(true)
      setAsyncCreate(true)
      setCreateDefaultAnnotation(true)
      setError(null)
      setLoading(false)
    }
  }, [isOpen])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    if (!name.trim()) {
      setError('Please provide a name.')
      return
    }
    if (!facsimileId) {
      setError('Please select a facsimile.')
      return
    }
    try {
      setError(null)
      setLoading(true)
      const dataset = await DatasetsService.postDatasets({
        dataset: {
          name: name.trim(),
          description: description.trim() || undefined,
          dpi: dpi || undefined,
          facsimile_id: facsimileId,
          pages: pages.trim() || undefined,
          deskewed,
        },
        async: asyncCreate || undefined,
        createDefaultAnnotation,
      })
      setState({ datasetId: dataset.id!, annotationId: '' })
      refetch()
      onCreated?.(dataset.id!)
      onClose()
    } catch (e) {
      console.error('Failed to create dataset:', e)
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
          <h2 className="text-lg font-semibold">Create dataset</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Dataset name
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

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              DPI
            </label>
            <input
              type="number"
              min={1}
              value={dpi}
              onChange={(e) => setDpi(Number(e.target.value) || DEFAULT_DPI)}
              className="w-full p-2 border border-gray-300 rounded-md"
              disabled={loading}
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Facsimile
            </label>
            <div className="flex items-center gap-2">
              <div className="flex-1 min-w-0">
                <Select
                  value={
                    facsimileOptions.find(
                      (o) => o.value.split('/')[0] === facsimileId,
                    ) ?? null
                  }
                  onChange={(opt) =>
                    setFacsimileId(opt?.value.split('/')[0] || null)
                  }
                  options={facsimileOptions}
                  placeholder="Select facsimile..."
                  isLoading={facsimilesLoading}
                  isDisabled={loading}
                  styles={selectStyles<{ value: string; label: string }>({
                    controlWidth: '100%',
                  })}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                />
              </div>
              {selectedFacsimileLabel && (
                <Button
                  type="button"
                  className="px-2 py-1.5 text-xs shrink-0"
                  onClick={handleCopyFacsimile}
                  disabled={loading}
                >
                  {facsimileCopied ? 'Copied!' : 'Copy'}
                </Button>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Pages (optional)
            </label>
            <input
              type="text"
              value={pages}
              onChange={(e) => setPages(e.target.value)}
              placeholder="e.g. 1,3-5,10 — leave empty for all"
              className="w-full p-2 border border-gray-300 rounded-md"
              disabled={loading}
            />
          </div>

          <label className="flex items-center gap-2 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={deskewed}
              onChange={(e) => setDeskewed(e.target.checked)}
              className="h-4 w-4"
              disabled={loading}
            />
            Automatic deskew
          </label>

          <label className="flex items-center gap-2 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={createDefaultAnnotation}
              onChange={(e) => setCreateDefaultAnnotation(e.target.checked)}
              className="h-4 w-4"
              disabled={loading}
            />
            Create default annotation
          </label>

          <label className="flex items-center gap-2 text-sm text-gray-700">
            <input
              type="checkbox"
              checked={asyncCreate}
              onChange={(e) => setAsyncCreate(e.target.checked)}
              className="h-4 w-4"
              disabled={loading}
            />
            Create in background (async)
          </label>

          <ErrorMessage message={error} />
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          {loading ? (
            <LoadingSpinner size="sm" message="Creating dataset..." />
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
                Create
              </Button>
            </>
          )}
        </div>
      </form>
    </div>
  )
}
