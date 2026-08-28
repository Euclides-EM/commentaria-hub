import { type FormEvent, useEffect, useMemo, useState } from 'react'
import {
  DatasetsService,
  FacsimilesService,
  ApiError,
  ShelfmarksService,
} from '@hub-api'
import { useQuery } from '@tanstack/react-query'
import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { Button } from '../core/Button.tsx'
import Select from 'react-select'
import { wrappedOptionSelectStyles } from '../../styles/selectStyles.ts'
import { ErrorMessage } from '../core/ErrorMessage'
import { useAppState } from '../../context/useAppState'
import {
  normalizeEditionId,
  useAllEditionsQuery,
} from '../../queries/editions.ts'
import {
  formatCopyright,
  getFacsimileCopyright,
} from '../../utils/copyright.ts'

const DEFAULT_DPI = 300

type FacsimileOption = {
  value: string
  label: string
  facsimileId: string
  editionId: string
  facsimileName?: string
  copyright?: string
}

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
  const [facsimileMode, setFacsimileMode] = useState<'existing' | 'url'>(
    'existing',
  )
  const { setState, refetch } = useAppState()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [dpi, setDpi] = useState(DEFAULT_DPI)
  const [selectedFacsimile, setSelectedFacsimile] =
    useState<FacsimileOption | null>(null)
  const [facsimileUrl, setFacsimileUrl] = useState('')
  const [pages, setPages] = useState('')
  const [deskewed, setDeskewed] = useState(true)
  const [denoised, setDenoised] = useState(true)
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
  const { data: editions, isLoading: editionsLoading } = useAllEditionsQuery(
    undefined,
    isOpen,
  )
  const { data: shelfmarks, isLoading: shelfmarksLoading } = useQuery({
    queryKey: ['shelfmarks'],
    queryFn: () => ShelfmarksService.getShelfmarks({}),
    enabled: isOpen,
  })

  const facsimileOptions = useMemo(() => {
    const shelfmarksByEdition = new Map<string, NonNullable<typeof shelfmarks>>()
    for (const shelfmark of shelfmarks ?? []) {
      const editionId = normalizeEditionId(shelfmark.edition_id ?? '')
      if (!editionId) {
        continue
      }
      shelfmarksByEdition.set(editionId, [
        ...(shelfmarksByEdition.get(editionId) ?? []),
        shelfmark,
      ])
    }

    const editionById = new Map(
      (editions ?? []).map((edition) => [
        normalizeEditionId(edition.key ?? ''),
        edition,
      ]),
    )

    const copyrightForFacsimile = (facsimile: NonNullable<typeof facsimiles>[number]) => {
      const editionId = normalizeEditionId(facsimile.edition_id ?? '')
      const edition = editionById.get(editionId)
      return getFacsimileCopyright(
        edition
          ? {
              ...edition,
              shelfmarks: shelfmarksByEdition.get(editionId) ?? [],
            }
          : undefined,
        facsimile,
      )
    }

    const existingOptions: FacsimileOption[] = (facsimiles ?? [])
      .filter((f) => f.id != null && f.download_available === true)
      .map((f) => {
        const editionId = f.edition_id ?? ''
        return {
          value: `existing:${f.id}`,
          label: editionId,
          facsimileId: f.id!,
          editionId,
          facsimileName: f.name?.trim() || '',
          copyright: copyrightForFacsimile(f),
        }
      })

    const facsimileCounts = new Map<string, number>()
    for (const option of existingOptions) {
      facsimileCounts.set(
        option.editionId,
        (facsimileCounts.get(option.editionId) ?? 0) + 1,
      )
    }

    const facsimileNumbers = new Map<string, number>()
    return existingOptions
      .sort((a, b) => a.editionId.localeCompare(b.editionId))
      .map((option) => {
        const number = (facsimileNumbers.get(option.editionId) ?? 0) + 1
        facsimileNumbers.set(option.editionId, number)
        const showNumber = (facsimileCounts.get(option.editionId) ?? 0) > 1
        const facsimile =
          option.facsimileName || (showNumber ? `facsimile ${number}` : '')
        const facsimileLabel = facsimile ? ` · ${facsimile}` : ''
        return {
          ...option,
          label: `${option.editionId}${facsimileLabel} · ${formatCopyright(
            option.copyright,
          )}`,
        }
      })
  }, [editions, facsimiles, shelfmarks])

  const selectedFacsimileLabel = selectedFacsimile?.label ?? null

  const handleCopyFacsimile = () => {
    if (!selectedFacsimileLabel) return
    void navigator.clipboard.writeText(selectedFacsimileLabel).then(() => {
      setFacsimileCopied(true)
      setTimeout(() => setFacsimileCopied(false), 2000)
    })
  }

  useEffect(() => {
    if (isOpen) {
      setFacsimileMode('existing')
      setName('')
      setDescription('')
      setDpi(DEFAULT_DPI)
      setSelectedFacsimile(null)
      setFacsimileUrl('')
      setFacsimileCopied(false)
      setPages('')
      setDeskewed(true)
      setDenoised(true)
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
    try {
      setError(null)
      setLoading(true)
      let datasetFacsimileId: string
      if (facsimileMode === 'existing') {
        if (!selectedFacsimile) {
          setError('Please select a facsimile.')
          return
        }
        datasetFacsimileId = selectedFacsimile.facsimileId
      } else {
        const url = facsimileUrl.trim()
        if (!url) {
          setError('Please provide a facsimile URL.')
          return
        }
        let parsedUrl: URL
        try {
          parsedUrl = new URL(url)
        } catch {
          setError('Please provide a valid facsimile URL.')
          return
        }
        const createdFacsimile = await FacsimilesService.postFacsimilies({
          facsimile: {
            scan_url: parsedUrl.toString(),
          },
        })
        if (!createdFacsimile.id) {
          setError('Failed to create facsimile from URL.')
          return
        }
        datasetFacsimileId = createdFacsimile.id
      }
      const dataset = await DatasetsService.postDatasets({
        dataset: {
          name: name.trim(),
          description: description.trim() || undefined,
          dpi: dpi || undefined,
          facsimile_id: datasetFacsimileId,
          pages: pages.trim() || undefined,
          deskewed,
          denoised,
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
            <div className="flex items-center gap-4">
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="radio"
                  name="facsimile-mode"
                  value="existing"
                  checked={facsimileMode === 'existing'}
                  onChange={() => setFacsimileMode('existing')}
                  disabled={loading}
                />
                Select existing
              </label>
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="radio"
                  name="facsimile-mode"
                  value="url"
                  checked={facsimileMode === 'url'}
                  onChange={() => setFacsimileMode('url')}
                  disabled={loading}
                />
                Import from URL
              </label>
            </div>
            {facsimileMode === 'existing' ? (
              <div className="flex items-center gap-2">
                <div className="flex-1 min-w-0">
                  <Select
                    value={selectedFacsimile}
                    onChange={setSelectedFacsimile}
                    options={facsimileOptions}
                    placeholder="Select facsimile..."
                    isLoading={
                      facsimilesLoading || editionsLoading || shelfmarksLoading
                    }
                    isDisabled={loading}
                    styles={wrappedOptionSelectStyles<FacsimileOption>({
                      controlWidth: '100%',
                      menuWidth: 'min(52rem, calc(100vw - 2rem))',
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
            ) : (
              <input
                type="url"
                value={facsimileUrl}
                onChange={(e) => setFacsimileUrl(e.target.value)}
                placeholder="https://example.org/facsimile"
                className="w-full p-2 border border-gray-300 rounded-md"
                disabled={loading}
              />
            )}
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

          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
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
                checked={denoised}
                onChange={(e) => setDenoised(e.target.checked)}
                className="h-4 w-4"
                disabled={loading}
              />
              Automatic denoise
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
          </div>

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
