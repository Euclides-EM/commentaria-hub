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
  const [prompt, setPrompt] = useState('')
  const [categorizer, setCategorizer] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isOpen) {
      setPrompt(latestRevision?.prompt ?? '')
      setCategorizer(latestRevision?.categorizer ?? '')
      setError(null)
      setLoading(false)
    }
  }, [isOpen, latestRevision])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    if (!prompt.trim() && !categorizer.trim()) {
      setError('At least one of prompt or categorizer is required.')
      return
    }
    try {
      setError(null)
      setLoading(true)
      await FeatureRevisionsService.postDatasetsFeaturesRevisions({
        dataSetId: datasetId,
        featureId,
        revision: {
          prompt: prompt.trim() || undefined,
          categorizer: categorizer.trim() || undefined,
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
        className="bg-white rounded-lg max-w-xl w-full h-[90vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Create revision</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 flex flex-col gap-4 text-sm">
          <div className="flex-1 flex flex-col gap-2 min-h-0">
            <label className="block text-sm font-medium text-gray-700">
              Prompt
            </label>
            <textarea
              value={prompt}
              onChange={(e) => setPrompt(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md flex-1 min-h-[100px] resize-none"
              disabled={loading}
            />
          </div>

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
