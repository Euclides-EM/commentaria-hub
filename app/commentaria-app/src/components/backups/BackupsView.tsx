import { type ChangeEvent, useRef, useState } from 'react'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import {
  useBackupsQuery,
  useCreateBackupFromZipMutation,
  useCreateBackupMutation,
  useRestoreBackupMutation,
} from '../../queries/backups'
import { RestoreBackupModal } from '../modal/RestoreBackupModal'
import { API_BASE_URL } from '../../config/api'
import { timeAgo } from '../../utils/timeAgo'
import { useAuthStore } from '../../store/authStore'

const getBackupCreatedAt = (backupId: string): string | null => {
  const match = backupId.match(/(\d{8}[tT]\d{6}[zZ])/)
  if (!match) {
    return null
  }
  const raw = match[1].toUpperCase()
  const iso = `${raw.slice(0, 4)}-${raw.slice(4, 6)}-${raw.slice(6, 8)}T${raw.slice(9, 11)}:${raw.slice(11, 13)}:${raw.slice(13, 15)}Z`
  return Number.isNaN(new Date(iso).getTime()) ? null : iso
}

const isNetworkError = (error: unknown): boolean => {
  if (!error || typeof error !== 'object') {
    return false
  }
  const err = error as { code?: string; message?: string }
  if (err.code === 'ERR_NETWORK') {
    return true
  }
  const message = (err.message || '').toLowerCase()
  return (
    message.includes('network error') ||
    message.includes('failed to fetch') ||
    message.includes('load failed')
  )
}

export function BackupsView() {
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const { data: backups, isLoading, error } = useBackupsQuery()
  const createMutation = useCreateBackupMutation()
  const createFromZipMutation = useCreateBackupFromZipMutation()
  const restoreMutation = useRestoreBackupMutation()
  const zipInputRef = useRef<HTMLInputElement>(null)
  const [downloadError, setDownloadError] = useState<unknown>(null)
  const [restoreTargetId, setRestoreTargetId] = useState<string | null>(null)
  const hasMutationError = Boolean(
    createMutation.error || createFromZipMutation.error || downloadError,
  )

  const handleDownload = (backupId: string) => {
    setDownloadError(null)
    const downloadUrl = `${API_BASE_URL}/backups/${encodeURIComponent(backupId)}`
    window.open(downloadUrl, '_blank', 'noopener,noreferrer')
  }

  const handleZipSelected = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }
    try {
      await createFromZipMutation.mutateAsync(file)
    } finally {
      event.target.value = ''
    }
  }

  const handleConfirmRestore = async () => {
    if (!restoreTargetId) {
      return
    }
    try {
      await restoreMutation.mutateAsync(restoreTargetId)
      setRestoreTargetId(null)
    } catch (error) {
      if (isNetworkError(error)) {
        window.location.reload()
      }
      return
    }
  }

  return (
    <div className="w-full h-full flex flex-col px-8">
      <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white gap-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-900">Backups</h2>
          <p className="text-xs text-gray-500">
            {backups?.length || 0} available backup
            {(backups?.length || 0) === 1 ? '' : 's'}
          </p>
        </div>
        {isAuthenticated && (
          <div className="flex items-center gap-2">
            <input
              ref={zipInputRef}
              type="file"
              accept=".zip,application/zip"
              className="hidden"
              onChange={handleZipSelected}
            />
            <Button
              variant="primary"
              className="px-3 py-2 text-sm"
              onClick={() => createMutation.mutate()}
              disabled={
                createMutation.isPending || createFromZipMutation.isPending
              }
            >
              {createMutation.isPending ? 'Creating...' : 'Create backup'}
            </Button>
            <Button
              variant="primary"
              className="px-3 py-2 text-sm"
              onClick={() => zipInputRef.current?.click()}
              disabled={
                createMutation.isPending || createFromZipMutation.isPending
              }
            >
              {createFromZipMutation.isPending
                ? 'Uploading...'
                : 'Upload from zip'}
            </Button>
          </div>
        )}
      </div>

      <div className="overflow-auto px-2 py-4">
        {isLoading && <LoadingSpinner message="Loading backups..." />}
        {!isLoading && <ErrorMessage error={error} />}
        {!isLoading && !error && hasMutationError && (
          <div className="mb-4">
            <ErrorMessage
              error={
                createMutation.error ||
                createFromZipMutation.error ||
                downloadError
              }
            />
          </div>
        )}
        {!isLoading && !error && backups?.length === 0 && (
          <div className="text-sm text-gray-500">No backups found.</div>
        )}
        {!isLoading && !error && (backups?.length || 0) > 0 && (
          <div className="overflow-auto rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full text-sm table-auto">
              <thead className="bg-gray-50 text-xs text-gray-500">
                <tr>
                  <th className="px-4 py-3 text-left whitespace-nowrap">ID</th>
                  <th className="px-4 py-3 text-left whitespace-nowrap">
                    Created
                  </th>
                  <th className="px-4 py-3 text-right whitespace-nowrap">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {backups?.map((backupId) => {
                  const createdAt = getBackupCreatedAt(backupId)
                  return (
                    <tr key={backupId} className="hover:bg-gray-50">
                      <td className="px-4 py-3 text-left">
                        <span className="font-mono text-xs text-gray-700">
                          {backupId}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-left whitespace-nowrap text-gray-700">
                        {createdAt ? (
                          <span title={new Date(createdAt).toLocaleString()}>
                            {timeAgo.format(new Date(createdAt))}
                          </span>
                        ) : (
                          <span className="text-gray-400">-</span>
                        )}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-2">
                          <Button
                            variant="primary"
                            className="px-2 py-1 text-xs"
                            onClick={() => handleDownload(backupId)}
                          >
                            Download
                          </Button>
                          {isAuthenticated && (
                            <Button
                              variant="danger"
                              className="px-2 py-1 text-xs"
                              onClick={() => setRestoreTargetId(backupId)}
                              disabled={restoreMutation.isPending}
                            >
                              Restore
                            </Button>
                          )}
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {isAuthenticated && (
        <RestoreBackupModal
          isOpen={!!restoreTargetId}
          backupId={restoreTargetId || ''}
          isRestoring={restoreMutation.isPending}
          error={restoreMutation.error}
          onCancel={() => {
            if (!restoreMutation.isPending) {
              setRestoreTargetId(null)
            }
          }}
          onConfirm={handleConfirmRestore}
        />
      )}
    </div>
  )
}
