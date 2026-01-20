import { useState } from 'react'
import { useAuthStore } from '../store/authStore'

interface LoginModalProps {
  onClose: () => void
  onSuccess: () => void
}

export function LoginModal({ onClose, onSuccess }: LoginModalProps) {
  const [token, setToken] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState('')
  const setAuth = useAuthStore((state) => state.setAuth)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token.trim()) return

    setIsLoading(true)
    setError('')

    try {
      const { AuthenticationService } =
        await import('../api/services/AuthenticationService')
      const userInfo = await AuthenticationService.postAuthValidate({
        requestBody: { token: token.trim() },
      })
      const displayName = userInfo.username || userInfo.email || 'Unknown User'
      setAuth(token.trim(), displayName)
      onSuccess()
    } catch (error) {
      console.error('Authentication failed:', error)
      setError('Invalid token. Please check your GitHub PAT and try again.')
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
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              disabled={isLoading}
            />
          </div>

          {error && <div className="text-red-600 text-sm">{error}</div>}

          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isLoading || !token.trim()}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isLoading ? 'Validating...' : 'Sign In'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
