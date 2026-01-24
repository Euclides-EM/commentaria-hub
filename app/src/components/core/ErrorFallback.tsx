import { ErrorMessage } from './ErrorMessage'

interface ErrorFallbackProps {
  error: Error
  onRetry?: () => void
  message?: string
}

export function ErrorFallback({ error, onRetry, message }: ErrorFallbackProps) {
  return (
    <div className="flex flex-col items-center justify-center p-6 text-center">
      <div className="text-red-500 text-2xl mb-2">⚠️</div>
      <h3 className="text-lg font-semibold text-gray-900 mb-2">
        {message || 'Something went wrong'}
      </h3>
      <div className="mb-4">
        <ErrorMessage error={error} variant="muted" />
      </div>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 bg-black text-white rounded-lg hover:bg-gray-800 font-semibold text-sm"
        >
          Try again
        </button>
      )}
    </div>
  )
}
