import { type FormEvent, useState } from 'react'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'

interface CreateAnnotationGroupModalProps {
  isOpen: boolean
  isSubmitting: boolean
  error?: string | null
  onClose: () => void
  onSubmit: (values: { name: string; description?: string }) => Promise<void>
  title?: string
  submitLabel?: string
  submittingLabel?: string
  description?: string
  initialName?: string
  initialDescription?: string
  selectedCount?: number
}

export function CreateAnnotationGroupModal({
  isOpen,
  isSubmitting,
  error,
  onClose,
  onSubmit,
  title = 'Create group',
  submitLabel = 'Create',
  submittingLabel = 'Creating...',
  description,
  initialName = '',
  initialDescription = '',
  selectedCount,
}: CreateAnnotationGroupModalProps) {
  if (!isOpen) {
    return null
  }

  return (
    <CreateAnnotationGroupModalContent
      key={`${initialName}\u0000${initialDescription}`}
      isSubmitting={isSubmitting}
      error={error}
      onClose={onClose}
      onSubmit={onSubmit}
      title={title}
      submitLabel={submitLabel}
      submittingLabel={submittingLabel}
      description={description}
      initialName={initialName}
      initialDescription={initialDescription}
      selectedCount={selectedCount}
    />
  )
}

type CreateAnnotationGroupModalContentProps = Omit<
  CreateAnnotationGroupModalProps,
  'isOpen'
>

function CreateAnnotationGroupModalContent({
  selectedCount,
  isSubmitting,
  error,
  onClose,
  onSubmit,
  title,
  submitLabel,
  submittingLabel,
  description,
  initialName = '',
  initialDescription = '',
}: CreateAnnotationGroupModalContentProps) {
  const [name, setName] = useState(initialName)
  const [descriptionValue, setDescriptionValue] = useState(initialDescription)

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmedName = name.trim()
    if (!trimmedName) {
      return
    }
    await onSubmit({
      name: trimmedName,
      description: descriptionValue.trim() || undefined,
    })
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
          <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
        </div>
        <div className="px-5 py-4 space-y-3 text-sm text-gray-700">
          {description ? (
            <p>{description}</p>
          ) : typeof selectedCount === 'number' ? (
            <p>
              Create a group with{' '}
              <span className="font-semibold">{selectedCount}</span> selected{' '}
              {selectedCount === 1 ? 'annotation' : 'annotations'}.
            </p>
          ) : null}
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
              value={descriptionValue}
              onChange={(e) => setDescriptionValue(e.target.value)}
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
            {isSubmitting ? submittingLabel : submitLabel}
          </Button>
        </div>
      </form>
    </div>
  )
}
