import { useState } from 'react'
import { useAuthStore } from '../../store/authStore'
import { ApiError, OpenAPI } from '../../api'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'

interface LoginModalProps {
  onClose: () => void
  onSuccess: () => void
}

const withTempToken = async <T,>(token: string, fn: () => Promise<T>) => {
  const originalToken = OpenAPI.TOKEN
  OpenAPI.TOKEN = token
  try {
    return await fn()
  } finally {
    OpenAPI.TOKEN = originalToken
  }
}

export function LoginModal({ onClose, onSuccess }: LoginModalProps) {
  const [token, setToken] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')
  const setAuth = useAuthStore((state) => state.setAuth)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token) {
      return
    }

    setIsLoading(true)
    setError('')

    try {
      const { AuthenticationService } =
        await import('../../api/services/AuthenticationService')
      const userInfo = await withTempToken(
        token,
        async () => await AuthenticationService.postAuthValidate(),
      )
      const displayName = userInfo.username || userInfo.email || 'Unknown User'
      setAuth(token, displayName)
      onSuccess()
    } catch (error) {
      console.error('Authentication failed:', error)
      if (error instanceof ApiError) {
        setError(
          typeof error.body === 'string'
            ? error.body
            : error.body
              ? JSON.stringify(error.body)
              : error.message,
        )
      } else {
        setError('Invalid token. Please check your GitHub PAT and try again.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-white rounded-lg p-6 w-full max-w-md">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold text-gray-900">Login</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <svg
              className="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <img
              src="/diagram.png"
              alt="Diagram"
              className="mx-auto h-64 w-auto mb-4"
            />
            <label
              htmlFor="modal-token"
              className="block text-sm font-medium text-gray-700 mb-2"
            >
              GitHub Personal Access Token
            </label>
            <input
              id="modal-token"
              name="token"
              type="password"
              autoComplete="current-password"
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
              value={token}
              onChange={(e) => setToken(e.target.value.trim())}
              disabled={isLoading}
            />
          </div>

          <ErrorMessage message={error} />

          <div className="flex gap-3">
            <Button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={isLoading || !token.trim()}
              variant="primary"
              className="flex-1 px-4 py-2"
            >
              {isLoading ? 'Validating...' : 'Sign In'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  )
}
