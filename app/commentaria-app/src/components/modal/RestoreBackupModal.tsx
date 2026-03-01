import { Button } from '../core/Button'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { ErrorMessage } from '../core/ErrorMessage'

interface RestoreBackupModalProps {
  isOpen: boolean
  backupId: string
  isRestoring: boolean
  error?: unknown
  onCancel: () => void
  onConfirm: () => void
}

export function RestoreBackupModal({
  isOpen,
  backupId,
  isRestoring,
  error,
  onCancel,
  onConfirm,
}: RestoreBackupModalProps) {
  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={isRestoring ? undefined : onCancel}
    >
      <div
        className="bg-white rounded-lg w-full max-w-md m-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">
            Restore from backup
          </h2>
        </div>
        <div className="px-5 py-4 text-sm text-gray-700 space-y-2">
          {isRestoring ? (
            <LoadingSpinner size="sm" message="Restoring from backup..." />
          ) : (
            <p>
              Restore from <span className="font-semibold">{backupId}</span>?
              This will replace the current system state.
            </p>
          )}
          <ErrorMessage error={error} />
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            onClick={onCancel}
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isRestoring}
          >
            Cancel
          </Button>
          <Button
            onClick={onConfirm}
            variant="danger"
            className="px-3 py-1.5 text-sm font-semibold"
            disabled={isRestoring}
          >
            {isRestoring ? 'Restoring...' : 'Restore'}
          </Button>
        </div>
      </div>
    </div>
  )
}
