import { useAuthStore } from '../store/authStore'
import { BreadcrumbNav } from './BreadcrumbNav'

interface HeaderProps {
  onShowLogin: () => void
}

export function Header({ onShowLogin }: HeaderProps) {
  const { token, username, clearAuth } = useAuthStore()

  return (
    <header className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold text-gray-800">
            Commentaria Hub
          </h1>
          <BreadcrumbNav />
        </div>

        <div className="flex items-center gap-4">
          {token && username ? (
            <>
              <span className="text-sm text-gray-600">
                Logged in as: <strong>{username}</strong>
              </span>
              <button
                onClick={clearAuth}
                className="px-3 py-1 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-md transition-colors"
              >
                Logout
              </button>
            </>
          ) : (
            <button
              onClick={onShowLogin}
              className="px-4 py-2 text-sm bg-blue-600 hover:bg-blue-700 text-white rounded-md transition-colors"
            >
              Login
            </button>
          )}
        </div>
      </div>
    </header>
  )
}
