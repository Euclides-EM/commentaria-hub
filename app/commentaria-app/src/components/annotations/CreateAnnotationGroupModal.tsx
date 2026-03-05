import { type FormEvent, useEffect, useState } from 'react'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'

interface CreateAnnotationGroupModalProps {
  isOpen: boolean
  selectedCount: number
  isSubmitting: boolean
  error?: string | null
  onClose: () => void
  onCreate: (values: { name: string; description?: string }) => Promise<void>
}

export function CreateAnnotationGroupModal({
  isOpen,
  selectedCount,
  isSubmitting,
  error,
  onClose,
  onCreate,
}: CreateAnnotationGroupModalProps) {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')

  useEffect(() => {
    if (!isOpen) {
      return
    }
    setName('')
    setDescription('')
  }, [isOpen])

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmedName = name.trim()
    if (!trimmedName) {
      return
    }
    await onCreate({
      name: trimmedName,
      description: description.trim() || undefined,
    })
  }

  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={isSubmitting ? undefined : onClose}
    >
      <form
        className="bg-white rounded-lg w-full max-w-md m-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Create group</h2>
        </div>
        <div className="px-5 py-4 space-y-3 text-sm text-gray-700">
          <p>
            Create a group with{' '}
            <span className="font-semibold">{selectedCount}</span> selected{' '}
            {selectedCount === 1 ? 'annotation' : 'annotations'}.
          </p>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Name
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              disabled={isSubmitting}
              required
              autoFocus
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Description (optional)
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full p-2 border border-gray-300 rounded-md"
              rows={3}
              disabled={isSubmitting}
            />
          </div>
          <ErrorMessage message={error} />
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm"
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            className="px-3 py-1.5 text-sm"
            disabled={isSubmitting || !name.trim()}
          >
            {isSubmitting ? 'Creating...' : 'Create'}
          </Button>
        </div>
      </form>
    </div>
  )
}
