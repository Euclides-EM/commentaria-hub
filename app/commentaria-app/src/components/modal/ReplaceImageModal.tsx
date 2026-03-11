import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'

interface ReplaceImageModalProps {
  isOpen: boolean
  title?: string
  body: string
  confirmLabel?: string
  loadingMessage?: string
  isReplacing: boolean
  error?: string | null
  onCancel: () => void
  onConfirm: () => void
}

export function ReplaceImageModal({
  isOpen,
  title = 'Replace image',
  body,
  confirmLabel = 'Choose image',
  loadingMessage = 'Replacing image...',
  isReplacing,
  error,
  onCancel,
  onConfirm,
}: ReplaceImageModalProps) {
  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={isReplacing ? undefined : onCancel}
    >
      <div
        className="bg-white rounded-lg w-full max-w-md m-4 shadow-lg"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
        </div>
        <div className="px-5 py-4 text-sm text-gray-700 space-y-2">
          {isReplacing ? (
            <LoadingSpinner size="sm" message={loadingMessage} />
          ) : (
            <p>{body}</p>
          )}
          <ErrorMessage message={error} />
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            onClick={onCancel}
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isReplacing}
          >
            Cancel
          </Button>
          <Button
            onClick={onConfirm}
            variant="primary"
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isReplacing}
          >
            {confirmLabel}
          </Button>
        </div>
      </div>
    </div>
  )
}
