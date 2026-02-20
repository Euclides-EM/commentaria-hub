import { type FormEvent, useEffect, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  type feature_ExecutionStrategy,
  type feature_Revision,
  FeatureRevisionsService,
} from '@hub-api'
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
  const [executionStrategy, setExecutionStrategy] =
    useState<feature_ExecutionStrategy>('prompt')
  const [prompt, setPrompt] = useState('')
  const [regex, setRegex] = useState('')
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isOpen) {
      setExecutionStrategy(latestRevision?.execution_strategy ?? 'prompt')
      setPrompt(latestRevision?.prompt ?? '')
      setRegex(latestRevision?.regex ?? '')
      setNote(latestRevision?.note ?? '')
      setError(null)
      setLoading(false)
    }
  }, [isOpen, latestRevision])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    if (executionStrategy === 'prompt' && !prompt.trim()) {
      setError('Prompt text is required for prompt-based revisions.')
      return
    }
    if (executionStrategy === 'regex' && !regex.trim()) {
      setError('Regex text is required for regex-based revisions.')
      return
    }
    try {
      setError(null)
      setLoading(true)
      await FeatureRevisionsService.postDatasetsFeaturesRevisions({
        dataSetId: datasetId,
        featureId,
        revision: {
          execution_strategy: executionStrategy,
          note: note.trim() || undefined,
          prompt: executionStrategy === 'prompt' ? prompt.trim() : undefined,
          regex: executionStrategy === 'regex' ? regex.trim() : undefined,
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
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Execution strategy
            </label>
            <select
              className="w-full p-2 border border-gray-300 rounded-md"
              value={executionStrategy}
              onChange={(e) =>
                setExecutionStrategy(
                  e.target.value as feature_ExecutionStrategy,
                )
              }
              disabled={loading}
            >
              <option value="prompt">Prompt</option>
              <option value="regex">Regex</option>
            </select>
          </div>

          {executionStrategy === 'prompt' && (
            <div className="flex-1 flex flex-col gap-2 min-h-0">
              <label className="block text-sm font-medium text-gray-700">
                Prompt
              </label>
              <textarea
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded-md flex-1 min-h-[100px] resize-none"
                required
                disabled={loading}
              />
            </div>
          )}

          {executionStrategy === 'regex' && (
            <div className="flex-1 flex flex-col gap-2 min-h-0">
              <label className="block text-sm font-medium text-gray-700">
                Regex
              </label>
              <textarea
                value={regex}
                onChange={(e) => setRegex(e.target.value)}
                className="w-full p-2 border border-gray-300 rounded-md flex-1 min-h-[100px] resize-none"
                required
                disabled={loading}
              />
            </div>
          )}

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Note (optional)
            </label>
            <textarea
              value={note}
              onChange={(e) => setNote(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              rows={3}
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
