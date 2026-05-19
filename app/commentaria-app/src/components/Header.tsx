import { useEffect, useState } from 'react'
import { HealthService, StoreService } from '@hub-api'
import { useAuthStore } from '../store/authStore'
import { BreadcrumbNav } from './BreadcrumbNav.tsx'
import { Button } from './core/Button'
import { useAppState } from '../context/useAppState.ts'
import { API_BASE_URL } from '../config/api.ts'
import { useQueryClient } from '@tanstack/react-query'
import { nonCompletedIntegrationJobsQueryKey } from '../queries/integrations.ts'

interface HeaderProps {
  onShowLogin: () => void
}

export function Header({ onShowLogin }: HeaderProps) {
  const { token, username, clearAuth } = useAuthStore()
  const { setState } = useAppState()
  const queryClient = useQueryClient()
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const [cleanupRequested, setCleanupRequested] = useState(false)
  const [facsimileImportStatus, setFacsimileImportStatus] = useState<
    'idle' | 'running' | 'done' | 'failed'
  >('idle')
  const [serverSha, setServerSha] = useState<string | null>(null)
  const backendUrl = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8085'
  const serverShaLabel =
    !token || !username ? 'Unavailable' : serverSha || 'Loading...'

  useEffect(() => {
    if (!token || !username) {
      return
    }

    let cancelled = false

    Promise.resolve().then(() => {
      if (!cancelled) {
        setServerSha(null)
      }
    })

    void HealthService.getHealth()
      .then((health) => {
        if (cancelled) {
          return
        }
        setServerSha(health.commit_sha || 'Unavailable')
      })
      .catch(() => {
        if (cancelled) {
          return
        }
        setServerSha('Unavailable')
      })

    return () => {
      cancelled = true
    }
  }, [token, username])

  const runStoreCleanup = () => {
    setCleanupRequested(true)
    void StoreService.deleteStoreCleanupLocal({}).catch(() => undefined)
  }

  const importFacsimilesFromDrive = () => {
    if (!token || facsimileImportStatus === 'running') {
      return
    }
    setFacsimileImportStatus('running')
    void fetch(`${API_BASE_URL}/facsimilies/import-from-drive?async=true`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(await response.text())
        }
        setFacsimileImportStatus('done')
        void queryClient.invalidateQueries({
          queryKey: nonCompletedIntegrationJobsQueryKey(),
        })
      })
      .catch(() => {
        setFacsimileImportStatus('failed')
      })
  }

  const closeMenu = () => {
    setIsMenuOpen(false)
    setCleanupRequested(false)
    if (facsimileImportStatus !== 'running') {
      setFacsimileImportStatus('idle')
    }
  }

  return (
    <header className="relative z-40 bg-white border-b border-gray-200 px-4 py-3">
      <div className="flex items-center justify-between flex-wrap">
        <div className="flex items-center gap-3">
          <div className="flex flex-col gap-1 justify-center items-center shrink-0">
            <h1 className="text-sm font-semibold text-gray-500">
              {import.meta.env.VITE_BACKEND_URL.includes('localhost')
                ? 'LOCALHOST'
                : 'Commentaria'}{' '}
              in Eucliedem
            </h1>
            <h1 className="text-xl font-semibold text-gray-800 tracking-widest annotations-shimmer">
              Annotations Hub
            </h1>
          </div>
          <BreadcrumbNav />
        </div>

        <div className="flex items-center gap-3 ml-auto">
          {token && username ? (
            <div
              className="relative"
              onMouseEnter={() => setIsMenuOpen(true)}
              onMouseLeave={closeMenu}
            >
              <button
                type="button"
                aria-label={`Open user menu for ${username}`}
                onClick={() => (isMenuOpen ? closeMenu() : setIsMenuOpen(true))}
                className="h-8 w-8 rounded-full bg-teal-600 text-white text-sm font-semibold flex items-center justify-center uppercase border border-teal-500 shadow-sm transition-transform hover:scale-105 hover:border-teal-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-teal-400"
              >
                {username.charAt(0)}
              </button>
              {isMenuOpen ? (
                <div className="absolute right-0 top-full z-50 pt-2">
                  <div className="w-80 rounded-md border border-gray-200 bg-white shadow-lg p-2">
                    <div className="px-2 py-2 text-xs text-gray-500">
                      Signed in as
                      <div className="text-sm font-semibold text-gray-900 truncate">
                        {username}
                      </div>
                    </div>
                    <div className="px-2 pb-2 text-xs text-gray-500 grid grid-cols-2 gap-3">
                      <div className="min-w-0">
                        Server URL
                        <div className="text-gray-800 break-all">
                          {backendUrl}
                        </div>
                      </div>
                      <div className="min-w-0">
                        Server SHA
                        <div className="text-gray-800 break-all">
                          {serverShaLabel}
                        </div>
                      </div>
                    </div>
                    <div className="border-t border-gray-100 my-1" />
                    <Button
                      variant="primary"
                      onClick={() => setState({ viewMode: 'backups' })}
                      className="w-full px-2 py-1 text-xs transition-colors mb-1"
                    >
                      Backups
                    </Button>
                    <Button
                      variant="regular"
                      onClick={importFacsimilesFromDrive}
                      disabled={facsimileImportStatus === 'running'}
                      className="w-full px-2 py-1 text-xs transition-colors mb-1"
                    >
                      {facsimileImportStatus === 'running'
                        ? 'Importing facsimiles...'
                        : facsimileImportStatus === 'done'
                          ? 'Import queued'
                          : facsimileImportStatus === 'failed'
                            ? 'Import failed'
                            : 'Import Facsimiles'}
                    </Button>
                    <Button
                      variant="regular"
                      onClick={runStoreCleanup}
                      className="w-full px-2 py-1 text-xs transition-colors mb-1"
                    >
                      {cleanupRequested ? 'Cleanup requested' : 'Cleanup Store'}
                    </Button>
                    <Button
                      variant="regular"
                      onClick={() => {
                        setState({ viewMode: 'logs' })
                        closeMenu()
                      }}
                      className="w-full px-2 py-1 text-xs transition-colors mb-1"
                    >
                      Logs
                    </Button>
                    <Button
                      variant="danger"
                      onClick={clearAuth}
                      className="w-full px-2 py-1 text-xs transition-colors"
                    >
                      Sign out
                    </Button>
                  </div>
                </div>
              ) : null}
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
