import { useMemo, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  ExecutionsService,
  type feature_Execution,
  type feature_ExecutionApplyItem,
  type feature_ExecutionSkipIf,
  type feature_ExecutionStatus,
  type feature_Feature,
  type model_Edition,
} from '@hub-api'
import Select from 'react-select'
import { useAuthStore } from '../../store/authStore'
import {
  useDatasetsQuery,
  useDatasetImageKeysQuery,
  useDatasetFeaturesQuery,
} from '../../queries/datasets'
import { useAnnotationsQuery } from '../../queries/annotations'
import { useAllEditionsQuery } from '../../queries/editions'
import { useFeaturesForExecutionsQuery } from '../../queries/features'
import { useFeatureExecutionsQuery } from '../../queries/executions'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { selectStyles } from '../../styles/selectStyles'
import { formatEditionLabel } from '../../utils/editions'
import { hasAnnotationPages } from '../../utils/editions'
import { parsePageEntries } from '../../utils/pages'
import { CreateFeatureExecutionModal } from '../annotation/featureExecutions/CreateFeatureExecutionModal'
import { FeatureExecutionLauncherModal } from './FeatureExecutionLauncherModal'

type ScopeFilter = 'all' | 'dataset' | 'editions'
type EntityOption = {
  value: string
  label: string
}

type ExecutionTarget =
  | { scope: 'editions' }
  | { scope: 'dataset'; datasetId: string; annotationId: string }

const EXECUTION_SKIP_IF_OPTIONS: feature_ExecutionSkipIf[] = [
  'feature_exist',
  'revision_exist',
  'human_reviewed',
  'value_not_empty',
]

const EXECUTION_SKIP_IF_LABELS: Record<feature_ExecutionSkipIf, string> = {
  feature_exist: 'Feature exist',
  revision_exist: 'Revision exist',
  human_reviewed: 'Human reviewed',
  value_not_empty: 'Value not empty',
}

const EXECUTION_STATUS_LABELS: Record<feature_ExecutionStatus, string> = {
  success: 'Completed',
  failed: 'Failed',
  in_progress: 'In progress',
  canceling: 'Canceling',
  canceled: 'Canceled',
}

const formatDate = (value?: string) => {
  if (!value) return 'Unknown'
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return value
  return parsed.toLocaleString()
}

export function FeatureExecutionsBrowser() {
  const queryClient = useQueryClient()
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const [actionError, setActionError] = useState<string | null>(null)
  const [cancelingExecutionId, setCancelingExecutionId] = useState<
    string | null
  >(null)
  const [executionStatusFilter, setExecutionStatusFilter] = useState<
    'all' | feature_ExecutionStatus
  >('all')
  const [scopeFilter, setScopeFilter] = useState<ScopeFilter>('all')
  const [selectedDatasetOption, setSelectedDatasetOption] =
    useState<EntityOption | null>(null)
  const [selectedAnnotationOption, setSelectedAnnotationOption] =
    useState<EntityOption | null>(null)
  const [expandedExecutionEditions, setExpandedExecutionEditions] = useState<
    Record<string, boolean>
  >({})
  const [isLaunchModalOpen, setIsLaunchModalOpen] = useState(false)
  const [pendingExecutionTarget, setPendingExecutionTarget] =
    useState<ExecutionTarget | null>(null)
  const [isCreatingExecution, setIsCreatingExecution] = useState(false)

  const datasetId = selectedDatasetOption?.value || ''
  const annotationId = selectedAnnotationOption?.value || ''
  const launchDatasetId =
    pendingExecutionTarget?.scope === 'dataset'
      ? pendingExecutionTarget.datasetId
      : ''
  const launchAnnotationId =
    pendingExecutionTarget?.scope === 'dataset'
      ? pendingExecutionTarget.annotationId
      : ''

  const datasetsQuery = useDatasetsQuery()
  const annotationsQuery = useAnnotationsQuery(datasetId)
  const datasetIds = useMemo(
    () =>
      (datasetsQuery.data || []).flatMap((dataset) =>
        dataset.id ? [dataset.id] : [],
      ),
    [datasetsQuery.data],
  )
  const launchAnnotationsQuery = useAnnotationsQuery(launchDatasetId)

  const [globalFeaturesQuery, ...datasetFeatureQueries] =
    useFeaturesForExecutionsQuery(datasetIds)

  const launchDatasetFeaturesQuery = useDatasetFeaturesQuery(
    launchDatasetId,
    pendingExecutionTarget?.scope === 'dataset' && !!launchDatasetId,
  )

  const executionsQuery = useFeatureExecutionsQuery({
    scope: scopeFilter,
    datasetId: datasetId,
  })

  const editionsQuery = useAllEditionsQuery()
  const editionsByKey = useMemo(() => {
    const map = new Map<string, model_Edition>()
    for (const item of editionsQuery.data ?? []) {
      map.set(item.key!, item)
    }
    return map
  }, [editionsQuery.data])

  const launchAnnotation = useMemo(
    () =>
      (launchAnnotationsQuery.data ?? []).find(
        (item) => item.id === launchAnnotationId,
      ) ?? null,
    [launchAnnotationId, launchAnnotationsQuery.data],
  )
  const launchAnnotationPageEntries = useMemo(
    () =>
      launchAnnotation ? parsePageEntries(launchAnnotation.pages || '') : [],
    [launchAnnotation],
  )
  const launchShouldLoadAnnotationImageKeys =
    !!launchAnnotation && !hasAnnotationPages(launchAnnotation)
  const launchAnnotationImageKeysQuery = useDatasetImageKeysQuery(
    launchDatasetId,
    launchShouldLoadAnnotationImageKeys,
    launchAnnotationPageEntries.length > 0 ? launchAnnotationPageEntries : null,
  )
  const launchAnnotationKeys = useMemo(() => {
    if (!launchAnnotation) {
      return []
    }
    if (launchAnnotationPageEntries.length > 0) {
      return [...new Set(launchAnnotationPageEntries)].sort((left, right) =>
        left.localeCompare(right, undefined, { numeric: true }),
      )
    }
    return (launchAnnotationImageKeysQuery.data ?? []).map((image) => image.key)
  }, [
    launchAnnotation,
    launchAnnotationImageKeysQuery.data,
    launchAnnotationPageEntries,
  ])
  const launchAnnotationEditionItems = useMemo(
    () =>
      launchAnnotationKeys.map((key) => {
        const edition = editionsByKey.get(key)
        if (edition) return edition
        return { key } as model_Edition
      }),
    [launchAnnotationKeys, editionsByKey],
  )

  const datasetOptions = useMemo(
    () =>
      [...(datasetsQuery.data ?? [])]
        .filter((dataset) => !!dataset.id)
        .map((dataset) => ({
          value: dataset.id!,
          label: dataset.name || dataset.id!,
        }))
        .sort((left, right) =>
          left.label.localeCompare(right.label, undefined, {
            sensitivity: 'base',
          }),
        ),
    [datasetsQuery.data],
  )

  const annotationOptions = useMemo(
    () =>
      [...(annotationsQuery.data ?? [])]
        .filter((annotation) => !!annotation.id)
        .map((annotation) => ({
          value: annotation.id!,
          label: annotation.name || annotation.id!,
        }))
        .sort((left, right) =>
          left.label.localeCompare(right.label, undefined, {
            sensitivity: 'base',
          }),
        ),
    [annotationsQuery.data],
  )

  const featureInfoById = useMemo(() => {
    const map: Record<string, { name: string; color?: string }> = {}

    for (const feature of globalFeaturesQuery?.data ?? []) {
      if (!feature.id) continue
      map[feature.id] = {
        name: feature.name || feature.id,
        color: feature.color || undefined,
      }
    }

    datasetFeatureQueries.forEach((query) => {
      for (const feature of query.data ?? []) {
        if (!feature.id) continue
        map[feature.id] = {
          name: feature.name || feature.id,
          color: feature.color || undefined,
        }
      }
    })

    return map
  }, [datasetFeatureQueries, globalFeaturesQuery?.data])

  const executionModalFeatures =
    pendingExecutionTarget?.scope === 'dataset'
      ? (launchDatasetFeaturesQuery.data ?? [])
      : (globalFeaturesQuery.data ?? [])
  const executionModalEditionItems =
    pendingExecutionTarget?.scope === 'dataset'
      ? launchAnnotationEditionItems
      : (editionsQuery.data ?? [])
  const executionModalLoadingFeatures =
    pendingExecutionTarget?.scope === 'dataset'
      ? launchDatasetFeaturesQuery.isLoading
      : globalFeaturesQuery.isLoading
  const executionModalLoadingEditions =
    pendingExecutionTarget?.scope === 'dataset'
      ? launchAnnotationsQuery.isLoading ||
        launchAnnotationImageKeysQuery.isLoading ||
        editionsQuery.isLoading
      : editionsQuery.isLoading
  const executionModalError =
    pendingExecutionTarget?.scope === 'dataset'
      ? launchDatasetFeaturesQuery.error instanceof Error
        ? launchDatasetFeaturesQuery.error.message
        : launchAnnotationsQuery.error instanceof Error
          ? launchAnnotationsQuery.error.message
          : launchAnnotationImageKeysQuery.error instanceof Error
            ? launchAnnotationImageKeysQuery.error.message
            : editionsQuery.error instanceof Error
              ? editionsQuery.error.message
              : null
      : globalFeaturesQuery.error instanceof Error
        ? globalFeaturesQuery.error.message
        : editionsQuery.error instanceof Error
          ? editionsQuery.error.message
          : null

  const datasetNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const dataset of datasetsQuery.data ?? []) {
      if (dataset.id) {
        map.set(dataset.id, dataset.name || dataset.id)
      }
    }
    return map
  }, [datasetsQuery.data])

  const annotationNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const annotation of annotationsQuery.data ?? []) {
      if (annotation.id) {
        map.set(annotation.id, annotation.name || annotation.id)
      }
    }
    return map
  }, [annotationsQuery.data])

  const launchAnnotationNameById = useMemo(() => {
    const map = new Map<string, string>()
    for (const annotation of launchAnnotationsQuery.data ?? []) {
      if (annotation.id) {
        map.set(annotation.id, annotation.name || annotation.id)
      }
    }
    return map
  }, [launchAnnotationsQuery.data])

  const filteredExecutions = useMemo(() => {
    const executionList = executionsQuery.data ?? []
    return executionList.filter((execution) => {
      if (
        executionStatusFilter !== 'all' &&
        execution.status !== executionStatusFilter
      ) {
        return false
      }
      if (datasetId && execution.scope?.dataset_id !== datasetId) {
        return false
      }
      return !(annotationId && execution.scope?.annotation_id !== annotationId);

    })
  }, [annotationId, datasetId, executionStatusFilter, executionsQuery.data])

  const cancelExecutionMutation = useMutation({
    mutationFn: (executionId: string) =>
      ExecutionsService.putFeatureExecutionsCancel({ executionId }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['executions'] })
    },
  })

  const executionsError = useMemo(() => {
    if (actionError) return actionError
    if (datasetsQuery.error instanceof Error) return datasetsQuery.error.message
    if (annotationsQuery.error instanceof Error)
      return annotationsQuery.error.message
    if (globalFeaturesQuery?.error instanceof Error)
      return globalFeaturesQuery.error.message
    const datasetFeaturesError = datasetFeatureQueries.find(
      (query) => query.error instanceof Error,
    )
    if (datasetFeaturesError?.error instanceof Error) {
      return datasetFeaturesError.error.message
    }
    if (executionsQuery.error instanceof Error)
      return executionsQuery.error.message
    if (editionsQuery.error instanceof Error) return editionsQuery.error.message
    if (
      datasetsQuery.error ||
      annotationsQuery.error ||
      globalFeaturesQuery?.error ||
      datasetFeatureQueries.some((query) => query.error) ||
      executionsQuery.error ||
      editionsQuery.error
    ) {
      return 'Failed to load feature executions data.'
    }
    return null
  }, [
    actionError,
    annotationsQuery.error,
    datasetFeatureQueries,
    datasetsQuery.error,
    editionsQuery.error,
    executionsQuery.error,
    globalFeaturesQuery?.error,
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

  const toggleExecutionEditions = (executionCardKey: string) => {
    setExpandedExecutionEditions((previous) => ({
      ...previous,
      [executionCardKey]: !previous[executionCardKey],
    }))
  }

  const handleLaunchContinue = (target: ExecutionTarget) => {
    setPendingExecutionTarget(target)
    setIsLaunchModalOpen(false)
  }

  const handleLaunchClose = () => {
    setIsLaunchModalOpen(false)
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
    if (!pendingExecutionTarget) {
      return
    }

    setActionError(null)
    setIsCreatingExecution(true)
    const executionFeatures =
      pendingExecutionTarget.scope === 'dataset'
        ? (launchDatasetFeaturesQuery.data ?? [])
        : (globalFeaturesQuery.data ?? [])
    const featureDefinitionById = new Map(
      executionFeatures
        .filter((feature): feature is feature_Feature & { id: string } =>
          Boolean(feature.id),
        )
        .map((feature) => [feature.id, feature]),
    )
    const apply: feature_ExecutionApplyItem[] = []
    for (const featureId of selectedFeatureIds) {
      const feature = featureDefinitionById.get(featureId)
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
      scope:
        pendingExecutionTarget.scope === 'dataset'
          ? {
              type: 'dataset',
              dataset_id: pendingExecutionTarget.datasetId,
              annotation_id: pendingExecutionTarget.annotationId,
            }
          : {
              type: 'editions',
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
      await ExecutionsService.postFeatureExecutions({
        execution: executionPayload,
      })
      setActionError(null)
      setPendingExecutionTarget(null)
    } catch (error) {
      setActionError(
        error instanceof Error ? error.message : 'Failed to create execution.',
      )
    } finally {
      setIsCreatingExecution(false)
    }
  }

  const executionsLoading =
    datasetsQuery.isLoading ||
    (scopeFilter !== 'editions' && annotationsQuery.isLoading) ||
    !!globalFeaturesQuery?.isLoading ||
    datasetFeatureQueries.some((query) => query.isLoading) ||
    executionsQuery.isLoading

  return (
    <section className="border border-gray-300 rounded-xl flex flex-col bg-white">
      <div className="px-3 py-2 rounded-t-xl border-b border-gray-200 bg-gray-50 flex items-center justify-between gap-3">
        <div className="text-sm font-semibold">Feature Executions</div>
        {isAuthenticated && (
          <Button
            type="button"
            variant="primary"
            className="px-2 py-1 text-xs"
            onClick={() => setIsLaunchModalOpen(true)}
          >
            Execute
          </Button>
        )}
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-4 space-y-3">
        <div className="flex items-center gap-3 flex-wrap">
          <select
            value={scopeFilter}
            onChange={(event) => {
              setScopeFilter(event.target.value as ScopeFilter)
              setSelectedDatasetOption(null)
              setSelectedAnnotationOption(null)
            }}
            className="h-8 px-2 text-xs border border-gray-300 rounded-md bg-white"
          >
            <option value="all">All scopes</option>
            <option value="editions">Editions</option>
            <option value="dataset">Dataset</option>
          </select>
          {scopeFilter !== 'editions' && (
            <Select
              value={selectedDatasetOption}
              onChange={(option) => {
                setSelectedDatasetOption(option)
                setSelectedAnnotationOption(null)
              }}
              options={datasetOptions}
              placeholder="Filter by dataset..."
              styles={selectStyles<EntityOption>({ controlWidth: 240 })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          )}
          {scopeFilter !== 'editions' && selectedDatasetOption && (
            <Select
              value={selectedAnnotationOption}
              onChange={(option) => setSelectedAnnotationOption(option)}
              options={annotationOptions}
              placeholder="Filter by annotation..."
              styles={selectStyles<EntityOption>({ controlWidth: 280 })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          )}
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
              : 'No executions match the selected filters.'}
          </div>
        ) : (
          filteredExecutions.map(
            (execution: feature_Execution, index: number) => {
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
              const executionDatasetId = execution.scope?.dataset_id || ''
              const executionAnnotationId = execution.scope?.annotation_id || ''
              const executionDatasetName =
                datasetNameById.get(executionDatasetId) || executionDatasetId
              const executionAnnotationName =
                annotationNameById.get(executionAnnotationId) ||
                executionAnnotationId

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
                      {isAuthenticated && canCancel && (
                        <Button
                          variant="danger"
                          type="button"
                          className="px-2 py-1 text-xs"
                          onClick={() =>
                            executionId &&
                            void handleCancelExecution(executionId)
                          }
                          disabled={isCanceling}
                        >
                          {isCanceling ? 'Canceling...' : 'Cancel'}
                        </Button>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-2 text-xs text-gray-500 flex-wrap">
                    <span className="rounded-full bg-gray-100 px-2 py-0.5 text-gray-700">
                      {execution.scope?.type === 'dataset'
                        ? 'Dataset'
                        : 'Editions'}
                    </span>
                    {executionDatasetName && (
                      <span>{executionDatasetName}</span>
                    )}
                    {executionAnnotationName && (
                      <span>{executionAnnotationName}</span>
                    )}
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
                        onClick={() =>
                          toggleExecutionEditions(executionCardKey)
                        }
                        className="text-xs text-gray-700 hover:text-gray-900"
                      >
                        {showExecutionEditions ? '▾' : '▸'} Including{' '}
                        {executionKeys.length}{' '}
                        {executionKeys.length === 1 ? 'edition' : 'editions'}.
                      </button>
                      {showExecutionEditions && (
                        <div className="border border-gray-200 rounded-md max-h-52 overflow-auto divide-y divide-gray-100">
                          {executionKeys.map((editionKey: string) => {
                            const item = editionsByKey.get(editionKey)
                            return (
                              <div
                                key={editionKey}
                                className="px-3 py-2 text-xs text-gray-700"
                              >
                                {item ? (
                                  formatEditionLabel(item)
                                ) : (
                                  <span>
                                    {editionKey} - details unavailable
                                  </span>
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
                      {execution.apply.map(
                        (
                          applyItem: feature_ExecutionApplyItem,
                          itemIndex: number,
                        ) => {
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
                                  backgroundColor:
                                    featureInfo?.color || '#d1d5db',
                                }}
                              />
                              <span>{label}</span>
                            </span>
                          )
                        },
                      )}
                    </div>
                  )}
                </article>
              )
            },
          )
        )}
      </div>

      <FeatureExecutionLauncherModal
        isOpen={isLaunchModalOpen}
        onClose={handleLaunchClose}
        onContinue={handleLaunchContinue}
        datasetOptions={datasetOptions}
        loadingDatasets={datasetsQuery.isLoading}
      />

      <CreateFeatureExecutionModal
        isOpen={pendingExecutionTarget !== null}
        onClose={() => {
          setPendingExecutionTarget(null)
          setActionError(null)
        }}
        onSubmit={handleCreateExecution}
        features={executionModalFeatures}
        editionItems={executionModalEditionItems}
        skipIfOptions={EXECUTION_SKIP_IF_OPTIONS}
        skipIfLabels={EXECUTION_SKIP_IF_LABELS}
        loadingFeatures={executionModalLoadingFeatures}
        loadingEditions={executionModalLoadingEditions}
        isSubmitting={isCreatingExecution}
        errorMessage={actionError || executionModalError}
        context={
          pendingExecutionTarget
            ? {
                scope: pendingExecutionTarget.scope,
                datasetName:
                  pendingExecutionTarget.scope === 'dataset'
                    ? datasetNameById.get(pendingExecutionTarget.datasetId)
                    : undefined,
                annotationName:
                  pendingExecutionTarget.scope === 'dataset'
                    ? launchAnnotationNameById.get(
                        pendingExecutionTarget.annotationId,
                      )
                    : undefined,
              }
            : undefined
        }
      />
    </section>
  )
}
