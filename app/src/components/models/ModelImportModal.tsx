import { type FormEvent, useEffect, useMemo, useState } from 'react'
import Select from 'react-select'
import type { model_Model } from '../../api'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { FileUpload } from '../core/FileUpload'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { selectStyles } from '../../styles/selectStyles'

interface ModelImportModalProps {
  isOpen: boolean
  models: model_Model[]
  onClose: () => void
  onSubmit: (payload: {
    file: File
    name: string
    description?: string
    baseModelId?: string
  }) => void
  isSaving?: boolean
  errorMessage?: string | null
}

export function ModelImportModal({
  isOpen,
  models,
  onClose,
  onSubmit,
  isSaving = false,
  errorMessage = null,
}: ModelImportModalProps) {
  const [file, setFile] = useState<File | null>(null)
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [baseModelId, setBaseModelId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen) {
      setFile(null)
      setName('')
      setDescription('')
      setBaseModelId(null)
      setError(null)
    }
  }, [isOpen])

  const baseModelOptions = useMemo(() => {
    return models
      .filter((model) => model.id)
      .map((model) => ({
        value: model.id as string,
        label: model.name || (model.id as string),
      }))
  }, [models])

  const handleSubmit = (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    if (!file) {
      setError('Please choose a model file.')
      return
    }
    setError(null)
    if (!name.trim()) {
      setError('Please provide a model name.')
      return
    }
    onSubmit({
      file,
      name: name.trim(),
      description: description.trim() || undefined,
      baseModelId: baseModelId || undefined,
    })
  }

  const handleFileChange = (nextFile: File | null) => {
    setFile(nextFile)
    if (nextFile && !name.trim()) {
      const fileName = nextFile.name
      const lastDot = fileName.lastIndexOf('.')
      const baseName = lastDot > 0 ? fileName.slice(0, lastDot) : fileName
      setName(baseName)
    }
  }

  if (!isOpen) {
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
          <h2 className="text-lg font-semibold">Import model</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Model file
            </label>
            <FileUpload
              file={file}
              onFileChange={handleFileChange}
              disabled={isSaving}
              required
            />
          </div>

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
              disabled={isSaving}
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
              disabled={isSaving}
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
              isDisabled={isSaving}
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
          {isSaving ? (
            <LoadingSpinner size="sm" message="Importing model..." />
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
