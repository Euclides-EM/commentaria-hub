import { Button } from '../core/Button'

interface NoticeModalProps {
  isOpen: boolean
  title: string
  message: string
  buttonLabel?: string
  onClose: () => void
}

export function NoticeModal({
  isOpen,
  title,
  message,
  buttonLabel = 'Close',
  onClose,
}: NoticeModalProps) {
  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-lg w-full max-w-md m-4 shadow-lg"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">{title}</h2>
        </div>
        <div className="px-5 py-4 text-sm text-gray-700">
          <p>{message}</p>
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex justify-end">
          <Button
            onClick={onClose}
            variant="primary"
            className="px-3 py-1.5 text-sm font-semibold"
          >
            {buttonLabel}
          </Button>
        </div>
      </div>
    </div>
  )
}
