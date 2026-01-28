import { useAuthStore } from '../store/authStore'
import { BreadcrumbNav } from './BreadcrumbNav.tsx'
import { Button } from './core/Button'

interface HeaderProps {
  onShowLogin: () => void
}

export function Header({ onShowLogin }: HeaderProps) {
  const { token, username, clearAuth } = useAuthStore()

  return (
    <header className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="flex items-center justify-between flex-wrap">
        <div className="flex items-center gap-3">
          <div className="flex flex-col gap-1 justify-center items-center">
            <h1 className="text-sm font-semibold text-gray-500">
              Commentaria in Eucliedem
            </h1>
            <h1 className="text-xl font-semibold text-gray-800 tracking-widest">
              Annotations Hub
            </h1>
          </div>
          <BreadcrumbNav />
        </div>

        <div className="flex items-center gap-3 ml-auto">
          {token && username ? (
            <div className="relative group">
              <button
                type="button"
                aria-label={`Open user menu for ${username}`}
                className="h-8 w-8 rounded-full bg-teal-600 text-white text-sm font-semibold flex items-center justify-center uppercase border border-teal-500 shadow-sm transition-transform group-hover:scale-105 group-hover:border-teal-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-teal-400"
              >
                {username.charAt(0)}
              </button>
              <div className="absolute right-0 mt-2 w-48 rounded-md border border-gray-200 bg-white shadow-lg p-2 hidden group-hover:block group-focus-within:block">
                <div className="px-2 py-2 text-xs text-gray-500">
                  Signed in as
                  <div className="text-sm font-semibold text-gray-900 truncate">
                    {username}
                  </div>
                </div>
                <div className="border-t border-gray-100 my-1" />
                <Button
                  variant="danger"
                  onClick={clearAuth}
                  className="w-full px-2 py-1 text-xs transition-colors"
                >
                  Sign out
                </Button>
              </div>
            </div>
          ) : (
            <Button
              onClick={onShowLogin}
              variant="primary"
              className="px-4 py-2 text-sm"
            >
              Login
            </Button>
          )}
        </div>
      </div>
    </header>
  )
}
