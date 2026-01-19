import { Suspense, type ReactNode, type ErrorInfo } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { ErrorFallback } from './ErrorFallback'
import { SuspenseFallback } from './SuspenseFallback'

interface AsyncBoundaryProps {
  children: ReactNode
  errorMessage?: string
  loadingMessage?: string
  onError?: (error: unknown, errorInfo: ErrorInfo) => void
  onRetry?: () => void
}

export function AsyncBoundary({
  children,
  errorMessage = 'Something went wrong',
  loadingMessage = 'Loading...',
  onError,
  onRetry,
}: AsyncBoundaryProps) {
  return (
    <ErrorBoundary
      FallbackComponent={({ error, resetErrorBoundary }) => (
        <ErrorFallback
          error={error as Error}
          message={errorMessage}
          onRetry={onRetry || resetErrorBoundary}
        />
      )}
      onError={onError}
    >
      <Suspense fallback={<SuspenseFallback message={loadingMessage} />}>
        {children}
      </Suspense>
    </ErrorBoundary>
  )
}
