import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'

interface DeleteAnnotationModalProps {
  isOpen: boolean
  annotationLabel: string
  error?: string | null
  isDeleting: boolean
  onCancel: () => void
  onConfirm: () => void
}

export function DeleteAnnotationModal({
  isOpen,
  annotationLabel,
  error,
  isDeleting,
  onCancel,
  onConfirm,
}: DeleteAnnotationModalProps) {
  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={isDeleting ? undefined : onCancel}
    >
      <div
        className="bg-white rounded-lg w-full max-w-md m-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">
            Delete annotation
          </h2>
        </div>
        <div className="px-5 py-4 text-sm text-gray-700 space-y-2">
          {isDeleting ? (
            <LoadingSpinner size="sm" message="Deleting annotation..." />
          ) : (
            <p>
              Are you sure you want to delete{' '}
              <span className="font-semibold">{annotationLabel}</span>? This
              action cannot be undone.
            </p>
          )}
          <ErrorMessage message={error} />
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            onClick={onCancel}
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isDeleting}
          >
            Cancel
          </Button>
          <Button
            onClick={onConfirm}
            variant="danger"
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isDeleting}
          >
            {isDeleting ? 'Deleting...' : 'Delete'}
          </Button>
        </div>
      </div>
    </div>
  )
}
