import { type FormEvent, useEffect, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { type feature_Revision, FeatureRevisionsService } from '@hub-api'
import Select from 'react-select'
import { useFeaturePropertiesQuery } from '../../queries/datasets.ts'
import { selectStyles } from '../../styles/selectStyles.ts'
import { getFeaturePropertyDisplayName } from '../../utils/featureProperties.ts'
import { Button } from '../core/Button.tsx'
import { ErrorMessage } from '../core/ErrorMessage.tsx'

interface CreateRevisionModalProps {
  isOpen: boolean
  onClose: () => void
  datasetId: string
  featureId: string
  latestRevision?: feature_Revision
}

type CategorizerOption = {
  value: string
  label: string
}

type AIProviderOption = {
  value: NonNullable<feature_Revision['ai_provider']>
  label: string
}

const aiProviderOptions: AIProviderOption[] = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'ollama', label: 'Ollama' },
]

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
  const [aiProvider, setAIProvider] =
    useState<feature_Revision['ai_provider']>()
  const [aiModel, setAIModel] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const {
    data: categorizerOptions = [],
    isLoading: isLoadingCategorizerOptions,
  } = useFeaturePropertiesQuery(isOpen)
  const categorizerSelectOptions = useMemo<CategorizerOption[]>(
    () =>
      categorizerOptions.map((option) => ({
        value: option,
        label: getFeaturePropertyDisplayName(option),
      })),
    [categorizerOptions],
  )

  useEffect(() => {
    if (isOpen) {
      const latestType = latestRevision?.prompt?.trim()
        ? 'prompt'
        : latestRevision?.categorizer?.trim()
          ? 'categorizer'
          : 'prompt'
      setType(latestType)
      setPrompt(latestRevision?.prompt ?? '')
      setCategorizer(latestRevision?.categorizer ?? '')
      setAIProvider(latestRevision?.ai_provider)
      setAIModel(latestRevision?.ai_model ?? '')
      setError(null)
      setLoading(false)
    }
  }, [isOpen, latestRevision])

  useEffect(() => {
    if (!isOpen || type !== 'categorizer') return

    if (categorizer && categorizerOptions.includes(categorizer)) return

    setCategorizer('')
  }, [categorizer, categorizerOptions, isOpen, type])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    const trimmedPrompt = prompt.trim()
    const trimmedCategorizer = categorizer.trim()
    const trimmedAIModel = aiModel.trim()
    const selectedValue = type === 'prompt' ? trimmedPrompt : trimmedCategorizer

    if (type === 'prompt' && !aiProvider) {
      setError('AI provider is required.')
      return
    }

    if (type === 'prompt' && !trimmedAIModel) {
      setError('AI model is required.')
      return
    }

    if (!selectedValue) {
      setError(
        type === 'prompt' ? 'Prompt is required.' : 'Categorizer is required.',
      )
      return
    }

    if (
      type === 'categorizer' &&
      !categorizerOptions.includes(trimmedCategorizer)
    ) {
      setError('Categorizer must be selected from feature properties.')
      return
    }

    try {
      setError(null)
      setLoading(true)
      await FeatureRevisionsService.postDatasetsFeaturesRevisions({
        dataSetId: datasetId,
        featureId,
        revision: {
          ai_provider: type === 'prompt' ? aiProvider : undefined,
          ai_model: type === 'prompt' ? trimmedAIModel : undefined,
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
            <span className="block text-sm font-medium text-gray-700">
              Type
            </span>
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
              <Select<CategorizerOption, false>
                options={categorizerSelectOptions}
                value={
                  categorizerSelectOptions.find(
                    (option) => option.value === categorizer,
                  ) ?? null
                }
                onChange={(option) => setCategorizer(option?.value ?? '')}
                isDisabled={loading || isLoadingCategorizerOptions}
                placeholder={
                  isLoadingCategorizerOptions
                    ? 'Loading feature properties...'
                    : categorizerOptions.length === 0
                      ? 'No feature properties available'
                      : 'Select a categorizer'
                }
                noOptionsMessage={() => 'No feature properties available'}
                styles={selectStyles<CategorizerOption>({
                  controlWidth: '100%',
                })}
                menuPortalTarget={document.body}
                menuPosition="fixed"
              />
            </div>
          )}

          {type === 'prompt' && (
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="flex flex-col gap-2">
                <label className="block text-sm font-medium text-gray-700">
                  AI provider
                </label>
                <Select<AIProviderOption, false>
                  options={aiProviderOptions}
                  value={
                    aiProviderOptions.find(
                      (option) => option.value === aiProvider,
                    ) ?? null
                  }
                  onChange={(option) => setAIProvider(option?.value)}
                  isDisabled={loading}
                  placeholder="Select a provider"
                  styles={selectStyles<AIProviderOption>({
                    controlWidth: '100%',
                  })}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                />
              </div>

              <div className="flex flex-col gap-2">
                <label className="block text-sm font-medium text-gray-700">
                  AI model
                </label>
                <input
                  value={aiModel}
                  onChange={(e) => setAIModel(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-md"
                  disabled={loading}
                  placeholder="gpt-5-mini"
                />
              </div>
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
