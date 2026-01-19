import { LoadingSpinner } from './LoadingSpinner'

interface SuspenseFallbackProps {
  message?: string
}

export function SuspenseFallback({
  message = 'Loading...',
}: SuspenseFallbackProps) {
  return (
    <div className="flex items-center justify-center min-h-32">
      <LoadingSpinner size="md" message={message} />
    </div>
  )
}
