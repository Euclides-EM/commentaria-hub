import { useMemo, useState, type FormEvent } from 'react'
import type { model_Model } from '../../api'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import Select from 'react-select'
import { selectStyles } from '../../styles/selectStyles.ts'

interface ModelEditModalProps {
  model: model_Model | null
  allModels: model_Model[]
  onClose: () => void
  onSubmit: (updates: {
    name: string
    description?: string
    type: string
    algorithm_family?: string
    base_model_id?: string
  }) => void
  isSaving?: boolean
  errorMessage?: string | null
}

export function ModelEditModal({
  model,
  allModels,
  onClose,
  onSubmit,
  isSaving = false,
  errorMessage = null,
}: ModelEditModalProps) {
  const [name, setName] = useState(() => model?.name || '')
  const [description, setDescription] = useState(() => model?.description || '')
  const [baseModelId, setBaseModelId] = useState<string | null>(
    () => model?.base_model_id || null,
  )
  const [error, setError] = useState<string | null>(null)

  const baseModelOptions = useMemo(() => {
    return allModels
      .filter((item) => item.id && item.id !== model?.id)
      .map((item) => ({
        value: item.id as string,
        label: item.name || (item.id as string),
      }))
  }, [allModels, model?.id])

  const handleSubmit = (event?: FormEvent<HTMLFormElement>) => {
    if (!model) {
      return
    }
    event?.preventDefault()
    if (!name.trim()) {
      setError('Please provide a model name.')
      return
    }
    onSubmit({
      name: name.trim(),
      description: description.trim() || undefined,
      type: model.type || 'segment',
      algorithm_family: model.algorithm_family || undefined,
      base_model_id: baseModelId || undefined,
    })
  }

  if (!model) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={onClose}
    >
      <form
        className="bg-white rounded-lg max-w-xl w-full max-h-[85vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Edit model</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Name
            </label>
            <input
              type="text"
              autoComplete="on"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              required
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
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Base model (optional)
            </label>
            <Select
              value={
                baseModelOptions.find(
                  (option) => option.value === baseModelId,
                ) || null
              }
              onChange={(option: { value: string; label: string } | null) =>
                setBaseModelId(option?.value || null)
              }
              options={baseModelOptions}
              placeholder="Select base model..."
              styles={selectStyles<{ value: string; label: string }>({
                controlWidth: 260,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          </div>

          {error && <ErrorMessage message={error} />}
          {!error && errorMessage && <ErrorMessage message={errorMessage} />}
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            onClick={onClose}
            type="button"
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isSaving}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isSaving}
          >
            {isSaving ? 'Saving...' : 'Save changes'}
          </Button>
        </div>
      </form>
    </div>
  )
}
