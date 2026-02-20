import { type FormEvent, useEffect, useState } from 'react'
import {
  ApiError,
  IntegrationService,
  type integration_JobTarget,
} from '@hub-api'
import { useQueryClient } from '@tanstack/react-query'
import { Button } from '../../core/Button.tsx'
import { LoadingSpinner } from '../../core/LoadingSpinner.tsx'
import { useAppState } from '../../../context/useAppState.ts'
import Select from 'react-select'
import { selectStyles } from '../../../styles/selectStyles.ts'
import { ErrorMessage } from '../../core/ErrorMessage'
import useLocalStorageState from 'use-local-storage-state'
import { runningIntegrationJobsQueryKey } from '../../../queries/integrations.ts'

type ExportMode = 'roboflow' | 'escriptorium'

interface ExportAnnotationModalProps {
  isOpen: boolean
  onClose: () => void
  exportTargets?: ExportAnnotationTarget[]
  defaultGroundTruthChecked?: boolean
}

export type ExportAnnotationTarget = {
  datasetId: string
  annotationId: string
}

type RoboflowSettings = Required<
  Pick<
    integration_JobTarget,
    'api_key' | 'workspace_url' | 'project_id' | 'is_not_ground_truth'
  >
>

type EscriptoriumSettings = Required<
  Pick<
    integration_JobTarget,
    'base_path' | 'document' | 'username' | 'password' | 'is_not_ground_truth'
  >
>

const exportOptions = [
  { value: 'roboflow', label: 'Upload to Roboflow' },
  { value: 'escriptorium', label: 'Upload to Escriptorium' },
] as const

export function ExportAnnotationModal({
  isOpen,
  onClose,
  exportTargets,
  defaultGroundTruthChecked = false,
}: ExportAnnotationModalProps) {
  const queryClient = useQueryClient()
  const { annotation } = useAppState()
  const [mode, setMode] = useState<ExportMode>('roboflow')
  const [roboflow, setRoboflow] = useLocalStorageState<RoboflowSettings>(
    'export-roboflow',
    {
      defaultValue: {
        api_key: '',
        workspace_url: 'mia-workplace',
        project_id: '',
        is_not_ground_truth: false,
      },
    },
  )
  const [escriptorium, setEscriptorium] =
    useLocalStorageState<EscriptoriumSettings>('export-escriptorium', {
      defaultValue: {
        base_path: '',
        document: '',
        username: '',
        password: '',
        is_not_ground_truth: false,
      },
    })
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const exportCount = exportTargets?.length || 0
  const modalTitle =
    exportCount > 1 ? `Export ${exportCount} annotations` : 'Export annotation'

  useEffect(() => {
    if (isOpen) {
      setError(null)
      setLoading(false)
      if (defaultGroundTruthChecked) {
        setRoboflow((prev) => ({
          ...prev,
          is_not_ground_truth: false,
        }))
        setEscriptorium((prev) => ({
          ...prev,
          is_not_ground_truth: false,
        }))
      }
    }
  }, [defaultGroundTruthChecked, isOpen, setEscriptorium, setRoboflow])

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    const resolvedTargets =
      exportTargets && exportTargets.length > 0
        ? exportTargets
        : annotation?.id && annotation.dataset_id
          ? [{ datasetId: annotation.dataset_id, annotationId: annotation.id }]
          : []

    if (resolvedTargets.length === 0) {
      return
    }

    try {
      setError(null)
      setLoading(true)

      const exportDetails =
        mode === 'roboflow'
          ? {
              platform: 'Roboflow' as const,
              api_key: roboflow.api_key,
              workspace_url: roboflow.workspace_url,
              project_id: roboflow.project_id,
              is_not_ground_truth: roboflow.is_not_ground_truth,
            }
          : {
              platform: 'EScriptorium' as const,
              base_path: escriptorium.base_path,
              document: escriptorium.document,
              username: escriptorium.username,
              password: escriptorium.password,
              is_not_ground_truth: escriptorium.is_not_ground_truth,
            }

      await IntegrationService.postIntegrationsJobs({
        job: {
          jobs: resolvedTargets.map((target) => ({
            annotation: {
              dataset_id: target.datasetId,
              id: target.annotationId,
            },
            task: 'Export',
            target: { ...exportDetails },
          })),
        },
      })

      void queryClient.invalidateQueries({
        queryKey: runningIntegrationJobsQueryKey(),
      })
      onClose()
    } catch (e) {
      console.error('Failed to export annotation:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    } finally {
      setLoading(false)
    }
  }

  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={loading ? undefined : onClose}
    >
      <form
        className="bg-white rounded-lg max-w-xl w-full min-h-[50vh] max-h-[85vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">{modalTitle}</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-4 text-sm">
          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Destination
            </label>
            <Select
              options={exportOptions}
              value={exportOptions.find((option) => option.value === mode)}
              onChange={(option) =>
                setMode((option?.value as ExportMode) || 'roboflow')
              }
              isDisabled={loading}
              className="text-sm"
              styles={selectStyles<{ value: ExportMode; label: string }>({
                controlWidth: 256,
              })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
            />
          </div>

          {mode === 'roboflow' && (
            <div className="space-y-3">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  API key
                </label>
                <input
                  type="password"
                  autoComplete="current-password"
                  value={roboflow.api_key}
                  onChange={(e) =>
                    setRoboflow((prev) => ({
                      ...prev,
                      api_key: e.target.value,
                    }))
                  }
                  className="w-full p-2 border border-gray-300 rounded-md"
                  disabled={loading}
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Workspace URL
                </label>
                <input
                  type="text"
                  autoComplete="url"
                  value={roboflow.workspace_url}
                  onChange={(e) =>
                    setRoboflow((prev) => ({
                      ...prev,
                      workspace_url: e.target.value,
                    }))
                  }
                  className="w-full p-2 border border-gray-300 rounded-md"
                  placeholder="your-workspace"
                  disabled={loading}
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Project ID
                </label>
                <input
                  type="text"
                  autoComplete="on"
                  value={roboflow.project_id}
                  onChange={(e) =>
                    setRoboflow((prev) => ({
                      ...prev,
                      project_id: e.target.value,
                    }))
                  }
                  className="w-full p-2 border border-gray-300 rounded-md"
                  disabled={loading}
                  required
                />
              </div>
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={roboflow.is_not_ground_truth}
                  onChange={(e) =>
                    setRoboflow((prev) => ({
                      ...prev,
                      is_not_ground_truth: !e.target.checked,
                    }))
                  }
                  className="h-4 w-4"
                  disabled={loading}
                />
                Mark as ground truth
              </label>
            </div>
          )}

          {mode === 'escriptorium' && (
            <div className="space-y-3">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Base URL
                </label>
                <input
                  type="url"
                  autoComplete="url"
                  value={escriptorium.base_path}
                  onChange={(e) =>
                    setEscriptorium((prev) => ({
                      ...prev,
                      base_path: e.target.value,
                    }))
                  }
                  className="w-full p-2 border border-gray-300 rounded-md"
                  placeholder="https://escriptorium.example.com"
                  disabled={loading}
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Document ID
                </label>
                <input
                  type="text"
                  autoComplete="on"
                  value={escriptorium.document}
                  onChange={(e) =>
                    setEscriptorium((prev) => ({
                      ...prev,
                      document: e.target.value,
                    }))
                  }
                  className="w-full p-2 border border-gray-300 rounded-md"
                  disabled={loading}
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Username
                </label>
                <input
                  type="text"
                  autoComplete="username"
                  value={escriptorium.username}
                  onChange={(e) =>
                    setEscriptorium((prev) => ({
                      ...prev,
                      username: e.target.value,
                    }))
                  }
                  className="w-full p-2 border border-gray-300 rounded-md"
                  disabled={loading}
                  required
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Password
                </label>
                <input
                  type="password"
                  autoComplete="current-password"
                  value={escriptorium.password}
                  onChange={(e) =>
                    setEscriptorium((prev) => ({
                      ...prev,
                      password: e.target.value,
                    }))
                  }
                  className="w-full p-2 border border-gray-300 rounded-md"
                  disabled={loading}
                  required
                />
              </div>
              <label className="flex items-center gap-2 text-sm text-gray-700">
                <input
                  type="checkbox"
                  checked={escriptorium.is_not_ground_truth}
                  onChange={(e) =>
                    setEscriptorium((prev) => ({
                      ...prev,
                      is_not_ground_truth: !e.target.checked,
                    }))
                  }
                  className="h-4 w-4"
                  disabled={loading}
                />
                Mark as ground truth
              </label>
            </div>
          )}

          <ErrorMessage message={error} />
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          {loading ? (
            <LoadingSpinner size="sm" message="Exporting annotation..." />
          ) : (
            <>
              <Button
                type="button"
                onClick={onClose}
                className="px-3 py-1.5 text-sm"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="primary"
                className="px-3 py-1.5 text-sm"
              >
                Export
              </Button>
            </>
          )}
        </div>
      </form>
    </div>
  )
}
