import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ExecutionsService,
  type feature_Execution,
  type feature_ExecutionApplyItem,
  type feature_ExecutionSkipIf,
  type feature_ExecutionStatus,
  type feature_Feature,
  FeaturesService,
  type model_Edition,
} from '@hub-api'
import { useAppState } from '../../../context/useAppState.ts'
import { useAuthStore } from '../../../store/authStore.ts'
import { Button } from '../../core/Button.tsx'
import { ErrorMessage } from '../../core/ErrorMessage.tsx'
import { CreateFeatureExecutionModal } from './CreateFeatureExecutionModal.tsx'
import { formatEditionLabel } from '../../../utils/editions.ts'
import { useAllEditionsQuery } from '../../../queries/editions.ts'

const EXECUTION_STATUS_LABELS: Record<feature_ExecutionStatus, string> = {
  success: 'Completed',
  failed: 'Failed',
  in_progress: 'In progress',
  canceling: 'Canceling',
  canceled: 'Canceled',
}

const EXECUTION_SKIP_IF_OPTIONS: feature_ExecutionSkipIf[] = [
  'feature_exist',
  'revision_exist',
  'human_reviewed',
]

const EXECUTION_SKIP_IF_LABELS: Record<feature_ExecutionSkipIf, string> = {
  feature_exist: 'Feature exist',
  revision_exist: 'Revision exist',
  human_reviewed: 'Human reviewed',
}

const formatDate = (value?: string) => {
  if (!value) return 'Unknown'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString()
}

export function FeatureExecutionsTab() {
  const queryClient = useQueryClient()
  const { state } = useAppState()
  const { datasetId, annotationId } = state
  const isAuthenticated = !!useAuthStore((store) => store.token)

  const [actionError, setActionError] = useState<string | null>(null)
  const [cancelingExecutionId, setCancelingExecutionId] = useState<
    string | null
  >(null)
  const [executionStatusFilter, setExecutionStatusFilter] = useState<
    'all' | feature_ExecutionStatus
  >('all')
  const [expandedExecutionEditions, setExpandedExecutionEditions] = useState<
    Record<string, boolean>
  >({})
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)

  const featuresQueryKey = ['features', 'revisions', datasetId]
  const executionsQueryKey = ['executions', datasetId]

  const featuresQuery = useQuery({
    queryKey: featuresQueryKey,
    queryFn: () =>
      FeaturesService.getDatasetsFeatures({
        dataSetId: datasetId,
        expand: ['revisions'],
      }),
    refetchOnWindowFocus: false,
  })

  const executionsQuery = useQuery({
    queryKey: executionsQueryKey,
    queryFn: () =>
      ExecutionsService.getFeaturesExecutions({ dataset: datasetId }),
    refetchInterval: 5 * 1000,
    refetchOnWindowFocus: false,
  })

  const editionsQuery = useAllEditionsQuery({
    titlePageStatus: ['No', 'Unknown'],
  })
  const editionsByKey = useMemo(() => {
    const map = new Map<string, model_Edition>()
    for (const item of editionsQuery.data ?? []) {
      map.set(item.key!, item)
    }
    return map
  }, [editionsQuery.data])

  const cancelExecutionMutation = useMutation({
    mutationFn: (executionId: string) =>
      ExecutionsService.putFeaturesExecutionsCancel({ executionId }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: executionsQueryKey })
    },
  })

  const createExecutionMutation = useMutation({
    mutationFn: (execution: feature_Execution) =>
      ExecutionsService.postFeaturesExecutions({ execution }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: executionsQueryKey })
    },
  })

  const sortedFeatures = useMemo(
    () =>
      [...(featuresQuery.data ?? [])].sort((left, right) =>
        (left.name || '').localeCompare(right.name || '', undefined, {
          sensitivity: 'base',
        }),
      ),
    [featuresQuery.data],
  )

  const featureInfoById = useMemo(() => {
    const map: Record<string, { name: string; color?: string }> = {}
    for (const feature of sortedFeatures) {
      if (!feature.id) continue
      map[feature.id] = {
        name: feature.name || feature.id,
        color: feature.color || undefined,
      }
    }
    return map
  }, [sortedFeatures])

  const filteredExecutions = useMemo(() => {
    const executionList = executionsQuery.data ?? []
    if (executionStatusFilter === 'all') return executionList
    return executionList.filter(
      (execution) => execution.status === executionStatusFilter,
    )
  }, [executionsQuery.data, executionStatusFilter])

  const executionsError = useMemo(() => {
    if (actionError) return actionError
    if (featuresQuery.error instanceof Error) return featuresQuery.error.message
    if (executionsQuery.error instanceof Error)
      return executionsQuery.error.message
    if (editionsQuery.error instanceof Error) return editionsQuery.error.message
    if (featuresQuery.error || executionsQuery.error || editionsQuery.error) {
      return 'Failed to load feature executions data.'
    }
    return null
  }, [
    actionError,
    editionsQuery.error,
    executionsQuery.error,
    featuresQuery.error,
  ])

  const handleCancelExecution = async (executionId: string) => {
    setCancelingExecutionId(executionId)
    setActionError(null)
    try {
      await cancelExecutionMutation.mutateAsync(executionId)
    } catch (error) {
      setActionError(
        error instanceof Error ? error.message : 'Failed to cancel execution.',
      )
    } finally {
      setCancelingExecutionId(null)
    }
  }

  const handleCreateExecution = async ({
    selectedFeatureIds,
    selectedKeys,
    skipIf,
    pushToOrigin,
  }: {
    selectedFeatureIds: string[]
    selectedKeys: string[]
    skipIf: feature_ExecutionSkipIf[]
    pushToOrigin: boolean
  }) => {
    setActionError(null)
    const featureById = new Map(
      sortedFeatures
        .filter((feature): feature is feature_Feature & { id: string } =>
          Boolean(feature.id),
        )
        .map((feature) => [feature.id, feature]),
    )
    const apply: feature_ExecutionApplyItem[] = []
    for (const featureId of selectedFeatureIds) {
      const feature = featureById.get(featureId)
      if (!feature) continue
      const revisions = [...(feature.revisions ?? [])].sort((left, right) => {
        const leftTime = left.created_at
          ? new Date(left.created_at).getTime()
          : 0
        const rightTime = right.created_at
          ? new Date(right.created_at).getTime()
          : 0
        return rightTime - leftTime
      })
      apply.push({
        feature: featureId,
        revision: revisions[0]?.id,
      })
    }

    const executionPayload: feature_Execution = {
      scope: {
        type: 'dataset',
        dataset_id: datasetId,
        annotation_id: annotationId,
      },
      apply,
      keys: selectedKeys,
      policy:
        skipIf.length || pushToOrigin
          ? {
              skip_if: skipIf.length ? skipIf : undefined,
              push_to_origin: pushToOrigin,
            }
          : undefined,
    }

    try {
      await createExecutionMutation.mutateAsync(executionPayload)
      setIsCreateModalOpen(false)
    } catch (error) {
      setActionError(
        error instanceof Error ? error.message : 'Failed to create execution.',
      )
    }
  }

  const toggleExecutionEditions = (executionCardKey: string) => {
    setExpandedExecutionEditions((previous) => ({
      ...previous,
      [executionCardKey]: !previous[executionCardKey],
    }))
  }

  const executionsLoading = executionsQuery.isLoading
  const creatingExecution = createExecutionMutation.isPending

  return (
    <section className="border border-gray-300 rounded-xl flex flex-col bg-white mb-0 w-[calc(100%-1.5rem)] max-w-[80vw] mx-auto">
      <div className="px-3 py-2 rounded-t-xl border-b border-gray-200 bg-gray-50 flex items-center justify-between gap-3">
        <div className="text-sm font-semibold">Feature Executions</div>
        <div className="flex items-center gap-2">
          {isAuthenticated && (
            <Button
              type="button"
              variant="primary"
              className="px-2 py-1 text-xs"
              onClick={() => setIsCreateModalOpen(true)}
            >
              Execute
            </Button>
          )}
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-4 space-y-3">
        <div className="flex items-center gap-2">
          <label
            htmlFor="execution-status-filter"
            className="text-xs font-medium text-gray-600"
          >
            Status
          </label>
          <select
            id="execution-status-filter"
            value={executionStatusFilter}
            aria-label="Filter executions by status"
            onChange={(event) =>
              setExecutionStatusFilter(
                event.target.value as 'all' | feature_ExecutionStatus,
              )
            }
            className="h-8 px-2 text-xs border border-gray-300 rounded-md bg-white"
          >
            <option value="all">All statuses</option>
            {(
              Object.keys(EXECUTION_STATUS_LABELS) as feature_ExecutionStatus[]
            ).map((status) => (
              <option key={status} value={status}>
                {EXECUTION_STATUS_LABELS[status]}
              </option>
            ))}
          </select>
        </div>

        <ErrorMessage message={executionsError} />

        {executionsLoading ? (
          <div className="text-sm text-gray-500">Loading executions...</div>
        ) : filteredExecutions.length === 0 ? (
          <div className="text-sm text-gray-500">
            {(executionsQuery.data?.length ?? 0) === 0
              ? 'No executions found yet.'
              : 'No executions match the selected status.'}
          </div>
        ) : (
          filteredExecutions.map((execution, index) => {
            const executionId = execution.id ?? ''
            const executionCardKey =
              executionId || execution.created_at || String(index)
            const isCanceling = cancelingExecutionId === executionId
            const canCancel =
              execution.status === 'in_progress' ||
              execution.status === 'canceling'
            const executionKeys = Array.from(new Set(execution.keys ?? []))
            const showExecutionEditions =
              expandedExecutionEditions[executionCardKey] ?? false

            return (
              <article
                key={executionCardKey}
                className="border border-gray-200 rounded-lg bg-white p-4 space-y-2"
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="text-sm font-semibold text-gray-900">
                    {execution.name || executionId || 'Unnamed'}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="inline-flex items-center rounded-full border border-gray-300 bg-gray-50 px-2 py-0.5 text-xs text-gray-700">
                      {execution.status
                        ? EXECUTION_STATUS_LABELS[execution.status] ||
                          execution.status
                        : 'Unknown'}
                    </span>
                    {canCancel && (
                      <Button
                        variant="danger"
                        type="button"
                        className="px-2 py-1 text-xs"
                        onClick={() =>
                          executionId && void handleCancelExecution(executionId)
                        }
                        disabled={isCanceling}
                      >
                        {isCanceling ? 'Canceling...' : 'Cancel'}
                      </Button>
                    )}
                  </div>
                </div>

                {execution.description && (
                  <div className="text-sm text-gray-700">
                    {execution.description}
                  </div>
                )}

                <div className="text-xs text-gray-500">
                  Created: {formatDate(execution.created_at)}
                </div>

                {executionKeys.length > 0 && (
                  <div className="space-y-2">
                    <button
                      type="button"
                      onClick={() => toggleExecutionEditions(executionCardKey)}
                      className="text-xs text-gray-700 hover:text-gray-900"
                    >
                      {showExecutionEditions ? '▾' : '▸'} Including{' '}
                      {executionKeys.length}{' '}
                      {executionKeys.length === 1 ? 'edition' : 'editions'}.
                    </button>
                    {showExecutionEditions && (
                      <div className="border border-gray-200 rounded-md max-h-52 overflow-auto divide-y divide-gray-100">
                        {executionKeys.map((editionKey) => {
                          const item = editionsByKey.get(editionKey)
                          return (
                            <div
                              key={editionKey}
                              className="px-3 py-2 text-xs text-gray-700"
                            >
                              {item ? (
                                formatEditionLabel(item)
                              ) : (
                                <span>{editionKey} - details unavailable</span>
                              )}
                            </div>
                          )
                        })}
                      </div>
                    )}
                  </div>
                )}

                {execution.apply && execution.apply.length > 0 && (
                  <div className="text-xs text-gray-600 flex flex-wrap items-center gap-1.5">
                    <span className="text-gray-500">Features:</span>
                    {execution.apply.map((applyItem, itemIndex) => {
                      const featureId = applyItem.feature ?? ''
                      const featureInfo = featureInfoById[featureId]
                      const label =
                        featureInfo?.name || featureId || 'Unknown feature'
                      return (
                        <span
                          key={`${featureId || 'unknown'}-${itemIndex}`}
                          title={featureId}
                          className="inline-flex items-center gap-1 rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5"
                        >
                          <span
                            className="inline-block w-2 h-2 rounded-full border border-gray-300"
                            style={{
                              backgroundColor: featureInfo?.color || '#d1d5db',
                            }}
                          />
                          <span>{label}</span>
                        </span>
                      )
                    })}
                  </div>
                )}
              </article>
            )
          })
        )}
      </div>

      <CreateFeatureExecutionModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onSubmit={handleCreateExecution}
        features={sortedFeatures}
        editionItems={editionsQuery.data ?? []}
        skipIfOptions={EXECUTION_SKIP_IF_OPTIONS}
        skipIfLabels={EXECUTION_SKIP_IF_LABELS}
        loadingFeatures={featuresQuery.isLoading}
        loadingEditions={editionsQuery.isLoading}
        isSubmitting={creatingExecution}
        errorMessage={executionsError}
      />
    </section>
  )
}
