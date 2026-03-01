import { type FormEvent, useEffect, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { type feature_Revision, FeatureRevisionsService } from '@hub-api'
import { Button } from '../core/Button.tsx'
import { ErrorMessage } from '../core/ErrorMessage.tsx'

interface CreateRevisionModalProps {
  isOpen: boolean
  onClose: () => void
  datasetId: string
  featureId: string
  latestRevision?: feature_Revision
}

export function CreateRevisionModal({
  isOpen,
  onClose,
  datasetId,
  featureId,
  latestRevision,
}: CreateRevisionModalProps) {
  const queryClient = useQueryClient()
  const [type, setType] = useState<'prompt' | 'categorizer'>('prompt')
  const [prompt, setPrompt] = useState('')
  const [categorizer, setCategorizer] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isOpen) {
      const latestType =
        latestRevision?.prompt?.trim()
          ? 'prompt'
          : latestRevision?.categorizer?.trim()
            ? 'categorizer'
            : 'prompt'
      setType(latestType)
      setPrompt(latestRevision?.prompt ?? '')
      setCategorizer(latestRevision?.categorizer ?? '')
      setError(null)
      setLoading(false)
    }
  }, [isOpen, latestRevision])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    const trimmedPrompt = prompt.trim()
    const trimmedCategorizer = categorizer.trim()
    const selectedValue =
      type === 'prompt' ? trimmedPrompt : trimmedCategorizer

    if (!selectedValue) {
      setError(
        type === 'prompt' ? 'Prompt is required.' : 'Categorizer is required.',
      )
      return
    }
    try {
      setError(null)
      setLoading(true)
      await FeatureRevisionsService.postDatasetsFeaturesRevisions({
        dataSetId: datasetId,
        featureId,
        revision: {
          prompt: type === 'prompt' ? trimmedPrompt : undefined,
          categorizer: type === 'categorizer' ? trimmedCategorizer : undefined,
        },
      })
      await queryClient.invalidateQueries({
        queryKey: ['features', 'definitions', datasetId],
      })
      onClose()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create revision.')
    } finally {
      setLoading(false)
    }
  }

  if (!isOpen) return null

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={loading ? undefined : onClose}
    >
      <form
        className="bg-white rounded-lg max-w-xl w-full max-h-[72vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Create revision</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 flex flex-col gap-4 text-sm">
          <div className="flex flex-col gap-2">
            <span className="block text-sm font-medium text-gray-700">Type</span>
            <div className="flex items-center gap-6">
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="radio"
                  name="revisionType"
                  value="prompt"
                  checked={type === 'prompt'}
                  onChange={() => setType('prompt')}
                  disabled={loading}
                />
                Prompt
              </label>
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="radio"
                  name="revisionType"
                  value="categorizer"
                  checked={type === 'categorizer'}
                  onChange={() => setType('categorizer')}
                  disabled={loading}
                />
                Categorizer
              </label>
            </div>
          </div>

          {type === 'prompt' ? (
            <div className="flex flex-col gap-2">
              <label className="block text-sm font-medium text-gray-700">
                Prompt
              </label>
              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded-md min-h-[100px] resize-y"
                disabled={loading}
              />
            </div>
          ) : (
            <div className="flex-1 flex flex-col gap-2 min-h-0">
              <label className="block text-sm font-medium text-gray-700">
                Categorizer
              </label>
              <textarea
                value={categorizer}
                onChange={(e) => setCategorizer(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded-md flex-1 min-h-[100px] resize-none"
                disabled={loading}
              />
            </div>
          )}

          <ErrorMessage message={error} />
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm"
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            className="px-3 py-1.5 text-sm"
            disabled={loading}
          >
            {loading ? 'Creating...' : 'Create'}
          </Button>
        </div>
      </form>
    </div>
  )
}
