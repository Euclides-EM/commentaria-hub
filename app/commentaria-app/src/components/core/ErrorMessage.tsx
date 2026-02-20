import { ApiError } from '@hub-api'

type ErrorVariant = 'inline' | 'compact' | 'centered' | 'empty' | 'muted'

interface ErrorMessageProps {
  error?: unknown
  message?: string | null
  variant?: ErrorVariant
}

const formatError = (error: unknown) => {
  if (!error) return ''
  if (error instanceof ApiError) {
    if (typeof error.body === 'string') return error.body
    if (error.body) return JSON.stringify(error.body)
    return error.message
  }
  if (error instanceof Error) return error.message
  if (typeof error === 'string') return error
  return String(error)
}

const variantStyles: Record<ErrorVariant, string> = {
  inline: 'text-sm text-red-600',
  compact: 'text-sm text-red-500 p-2',
  centered: 'text-xs text-red-600 text-center py-6',
  empty: 'w-full m-10 font-medium text-center text-red-800',
  muted: 'text-sm text-gray-600',
}

export function ErrorMessage({
  error,
  message,
  variant = 'inline',
}: ErrorMessageProps) {
  const text = message != null ? message : formatError(error)
  if (!text) return null
  return <div className={variantStyles[variant]}>{text}</div>
}
