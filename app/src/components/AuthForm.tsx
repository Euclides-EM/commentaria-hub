import { useState } from 'react'
import { useAuthValidateMutation } from '../queries/auth'
import { useAuthStore } from '../store/authStore'

export function AuthForm() {
  const [token, setToken] = useState('')
  const setAuth = useAuthStore((state) => state.setAuth)
  const authMutation = useAuthValidateMutation()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!token.trim()) return

    try {
      const userInfo = await authMutation.mutateAsync({ token: token.trim() })
      const displayName = userInfo.username || userInfo.email || 'Unknown User'
      setAuth(token.trim(), displayName)
    } catch (error) {
      console.error('Authentication failed:', error)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 px-4">
      <div className="max-w-md w-full space-y-8">
        <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
          Authentication Required
        </h2>
        <img
          src="/diagram.png"
          alt="Diagram"
          className="mx-auto h-full w-auto"
        />
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div>
            <label
              htmlFor="token"
              className="block text-sm font-medium text-gray-700 mb-2"
            >
              GitHub Personal Access Token
            </label>
            <input
              id="token"
              name="token"
              type="password"
              required
              className="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
              placeholder="ghp_xxxxxxxxxxxxxxxxxxxx"
              value={token}
              onChange={(e) => setToken(e.target.value)}
              disabled={authMutation.isPending}
            />
          </div>

          {authMutation.isError && (
            <div className="text-red-600 text-sm text-center">
              Invalid token. Please check your GitHub PAT and try again.
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={authMutation.isPending || !token.trim()}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {authMutation.isPending ? 'Validating...' : 'Sign In'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
