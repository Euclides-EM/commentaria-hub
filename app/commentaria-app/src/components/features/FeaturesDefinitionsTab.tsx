import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import {
  EditionFeaturesService,
  type feature_Feature,
  type feature_Revision,
} from '@hub-api'
import { useAppState } from '../../context/useAppState'
import { useAuthStore } from '../../store/authStore'
import {
  useFeaturePropertiesQuery,
  useDatasetsQuery,
} from '../../queries/datasets'
import {
  useGlobalFeaturesQuery,
  useAllDatasetsFeaturesQueries,
} from '../../queries/features'
import { normalizeFeatureProperties } from '../../utils/featureProperties'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { MultiSelectDropdown } from '../core/MultiSelectDropdown'
import { SearchInput } from '../core/SearchInput'
import { FeatureCard, type FeatureEditState } from '../dataset/FeatureCard'
import { CreateRevisionModal } from '../dataset/CreateRevisionModal'

type FeatureRow = {
  datasetId: string | null
  datasetName: string
  feature: feature_Feature
}

const areFeatureEditsEqual = (
  left: Record<string, FeatureEditState>,
  right: Record<string, FeatureEditState>,
) => {
  const leftKeys = Object.keys(left)
  const rightKeys = Object.keys(right)

  if (leftKeys.length !== rightKeys.length) {
    return false
  }

  for (const key of leftKeys) {
    const leftEdit = left[key]
    const rightEdit = right[key]
    if (!rightEdit) {
      return false
    }
    if (
      leftEdit.name !== rightEdit.name ||
      leftEdit.description !== rightEdit.description ||
      leftEdit.color !== rightEdit.color ||
      leftEdit.properties.join('|') !== rightEdit.properties.join('|')
    ) {
      return false
    }
  }

  return true
}

const getScopeLabel = (row: FeatureRow) => {
  const scopeType = row.feature.scope?.type
  if (scopeType === 'dataset') {
    return 'Dataset'
  }
  if (scopeType === 'editions') {
    return 'Editions'
  }
  return row.datasetId ? 'Dataset' : 'Editions'
}

export function FeaturesDefinitionsTab() {
  const queryClient = useQueryClient()
  const { getUrlForState } = useAppState()
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const {
    data: datasets,
    isLoading: datasetsLoading,
    error: datasetsError,
  } = useDatasetsQuery()
  const { data: availableProperties = [], isLoading: isLoadingProperties } =
    useFeaturePropertiesQuery(true)
  const datasetIds = useMemo(
    () =>
      (datasets || []).flatMap((dataset) => (dataset.id ? [dataset.id] : [])),
    [datasets],
  )
  const globalFeaturesQuery = useGlobalFeaturesQuery()
  const datasetFeatureQueries = useAllDatasetsFeaturesQueries(datasetIds)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedDatasets, setSelectedDatasets] = useState<string[] | null>(
    null,
  )
  const [selectedScopes, setSelectedScopes] = useState<string[] | null>(null)
  const [featureEdits, setFeatureEdits] = useState<
    Record<string, FeatureEditState>
  >({})
  const [expandedFeatures, setExpandedFeatures] = useState<
    Record<string, boolean>
  >({})
  const [editingFeatures, setEditingFeatures] = useState<
    Record<string, boolean>
  >({})
  const [busyFeatureId, setBusyFeatureId] = useState<string | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)
  const [revisionModal, setRevisionModal] = useState<{
    featureId: string
    latestRevision?: feature_Revision
  } | null>(null)

  const rows = useMemo(() => {
    const datasetNameById = new Map<string, string>()
    datasets?.forEach((dataset) => {
      if (dataset.id) {
        datasetNameById.set(dataset.id, dataset.name || dataset.id)
      }
    })

    const result: FeatureRow[] = []

    ;(globalFeaturesQuery?.data || []).forEach((feature) => {
      result.push({
        datasetId: null,
        datasetName: '',
        feature,
      })
    })

    datasetFeatureQueries.forEach((query, index) => {
      const datasetId = datasetIds[index]
      if (!datasetId || !query.data) {
        return
      }

      const datasetName = datasetNameById.get(datasetId) || datasetId
      query.data.forEach((feature) => {
        result.push({
          datasetId,
          datasetName,
          feature,
        })
      })
    })

    return result.sort((a, b) =>
      (a.feature.name || a.feature.id || '').localeCompare(
        b.feature.name || b.feature.id || '',
        undefined,
        { sensitivity: 'base' },
      ),
    )
  }, [datasetFeatureQueries, datasetIds, datasets, globalFeaturesQuery?.data])

  useEffect(() => {
    setFeatureEdits((previous) => {
      const next: Record<string, FeatureEditState> = {}
      for (const row of rows) {
        const feature = row.feature
        if (!feature.id) continue
        if (editingFeatures[feature.id] && previous[feature.id]) {
          next[feature.id] = previous[feature.id]
          continue
        }
        next[feature.id] = {
          name: feature.name || '',
          description: feature.description || '',
          color: feature.color || '',
          properties: normalizeFeatureProperties(feature.properties ?? []),
        }
      }
      return areFeatureEditsEqual(previous, next) ? previous : next
    })
  }, [rows, editingFeatures])

  const datasetFilterItems = useMemo(
    () =>
      Array.from(
        new Set(rows.map((row) => row.datasetName).filter((name) => !!name)),
      ).sort(),
    [rows],
  )
  const scopeFilterItems = useMemo(
    () => Array.from(new Set(rows.map((row) => getScopeLabel(row)))).sort(),
    [rows],
  )

  const filteredRows = useMemo(() => {
    const trimmed = searchQuery.trim().toLowerCase()
    const isAllDatasetsSelected =
      selectedDatasets == null ||
      selectedDatasets.length === datasetFilterItems.length
    return rows.filter((row) => {
      const isEditionsScope = getScopeLabel(row) === 'Editions'
      const matchesDataset =
        selectedDatasets == null ||
        selectedDatasets.includes(row.datasetName) ||
        (isAllDatasetsSelected && isEditionsScope)
      const matchesScope =
        selectedScopes == null || selectedScopes.includes(getScopeLabel(row))
      if (!matchesDataset || !matchesScope) {
        return false
      }
      if (!trimmed) {
        return true
      }
      const haystack = [
        row.datasetId,
        row.datasetName,
        row.feature.id,
        row.feature.name,
        row.feature.description,
        row.feature.color,
        getScopeLabel(row),
        ...(row.feature.properties || []),
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
      return haystack.includes(trimmed)
    })
  }, [
    datasetFilterItems.length,
    rows,
    searchQuery,
    selectedDatasets,
    selectedScopes,
  ])

  const updateFeatureMutation = useMutation({
    mutationFn: ({
      featureId,
      feature,
    }: {
      featureId: string
      feature: feature_Feature
    }) =>
      EditionFeaturesService.putFeatures({
        featureId,
        feature,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['features'] })
    },
  })

  const deleteFeatureMutation = useMutation({
    mutationFn: (featureId: string) =>
      EditionFeaturesService.deleteFeatures1({
        featureId,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['features'] })
    },
  })

  const isLoading =
    datasetsLoading ||
    isLoadingProperties ||
    !!globalFeaturesQuery?.isLoading ||
    datasetFeatureQueries.some((query) => query.isLoading)
  const queryError =
    datasetsError ||
    globalFeaturesQuery?.error ||
    datasetFeatureQueries.find((query) => query.error)?.error ||
    null
  const error = actionError || (queryError ? String(queryError) : null)

  const handleUpdateFeature = async (
    feature: feature_Feature,
    event?: React.FormEvent<HTMLFormElement>,
  ) => {
    event?.preventDefault()
    if (!feature.id) return
    const form = featureEdits[feature.id]
    if (!form?.name?.trim()) {
      setActionError('Feature name is required.')
      return
    }
    setBusyFeatureId(feature.id)
    setActionError(null)
    try {
      await updateFeatureMutation.mutateAsync({
        featureId: feature.id,
        feature: {
          name: form.name.trim(),
          description: form.description.trim() || undefined,
          color: form.color || undefined,
          is_default: feature.is_default ?? false,
          is_list: feature.is_list ?? true,
          properties: normalizeFeatureProperties(form.properties),
          scope: feature.scope,
        },
      })
      setEditingFeatures((prev) => ({ ...prev, [feature.id as string]: false }))
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : 'Failed to update feature.',
      )
    } finally {
      setBusyFeatureId(null)
    }
  }

  const handleDeleteFeature = async (feature: feature_Feature) => {
    if (!feature.id) return
    if (!window.confirm(`Delete "${feature.name || 'this feature'}"?`)) return
    setBusyFeatureId(feature.id)
    setActionError(null)
    try {
      await deleteFeatureMutation.mutateAsync(feature.id)
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : 'Failed to delete feature.',
      )
    } finally {
      setBusyFeatureId(null)
    }
  }

  const handleCancelEdit = (feature: feature_Feature) => {
    if (!feature.id) return
    setFeatureEdits((prev) => ({
      ...prev,
      [feature.id as string]: {
        name: feature.name || '',
        description: feature.description || '',
        color: feature.color || '',
        properties: normalizeFeatureProperties(feature.properties ?? []),
      },
    }))
    setEditingFeatures((prev) => ({ ...prev, [feature.id as string]: false }))
  }

  const renderDatasetLink = (row: FeatureRow) => {
    if (!row.datasetId) return null
    return (
      <a
        href={getUrlForState({
          viewMode: null,
          datasetId: row.datasetId,
          annotationId: '',
        })}
        target="_blank"
        rel="noopener noreferrer"
        className="text-teal-700 hover:text-teal-900 hover:underline text-xs font-medium"
      >
        {row.datasetName}
      </a>
    )
  }

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col bg-white">
      <div className="px-4 py-4 border-b border-gray-200 bg-white flex items-center justify-between gap-4">
        <div className="flex items-center gap-6 flex-wrap">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Features</h2>
            <p className="text-xs text-gray-500">
              {searchQuery.trim() || selectedScopes || selectedDatasets
                ? `Listing ${filteredRows.length} of ${rows.length} features`
                : `Listing ${rows.length} features`}
            </p>
          </div>
          <SearchInput
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Search features..."
            className="w-[22rem] max-w-full"
          />
          <MultiSelectDropdown
            allItems={scopeFilterItems}
            selectedItems={selectedScopes}
            setSelectedItems={setSelectedScopes}
            itemsLabel="scopes"
            getItemLabel={(item) => item}
            minWidth="140px"
          />
          <MultiSelectDropdown
            allItems={datasetFilterItems}
            selectedItems={selectedDatasets}
            setSelectedItems={setSelectedDatasets}
            itemsLabel="datasets"
            getItemLabel={(item) => item}
            minWidth="180px"
          />
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="flex flex-col gap-5">
          {isLoading && <LoadingSpinner message="Loading features..." />}

          {!isLoading && error && <ErrorMessage message={error} />}

          {!isLoading && !error && rows.length === 0 && (
            <div className="text-sm text-gray-500">No features found.</div>
          )}

          {!isLoading &&
            !error &&
            rows.length > 0 &&
            filteredRows.length === 0 && (
              <div className="text-sm text-gray-500">
                No features match the current filters.
              </div>
            )}

          {!isLoading &&
            !error &&
            filteredRows.map((row) => {
              const featureId = row.feature.id ?? ''
              const edits = featureId ? featureEdits[featureId] : undefined
              const isEditing = featureId ? editingFeatures[featureId] : false
              const isDirty =
                edits &&
                (edits.name !== (row.feature.name || '') ||
                  edits.description !== (row.feature.description || '') ||
                  edits.color !== (row.feature.color || '') ||
                  edits.properties.join('|') !==
                    normalizeFeatureProperties(
                      row.feature.properties ?? [],
                    ).join('|'))

              return (
                <div
                  key={`${row.datasetId || 'global'}-${featureId || row.feature.name}`}
                >
                  <div className="mb-2 flex items-center gap-3 text-xs text-gray-500">
                    <span className="rounded-full bg-gray-100 px-2.5 py-1 font-medium text-gray-700">
                      {getScopeLabel(row)}
                    </span>
                    {renderDatasetLink(row)}
                  </div>
                  <FeatureCard
                    feature={row.feature}
                    edits={edits}
                    isEditing={!!isEditing}
                    isExpanded={
                      featureId ? !!expandedFeatures[featureId] : false
                    }
                    isSaving={busyFeatureId === featureId}
                    isDirty={!!isDirty}
                    isAuthenticated={isAuthenticated}
                    availableProperties={availableProperties}
                    isLoadingProperties={isLoadingProperties}
                    onSubmit={(event) =>
                      handleUpdateFeature(row.feature, event)
                    }
                    onEdit={() =>
                      featureId &&
                      setEditingFeatures((prev) => ({
                        ...prev,
                        [featureId]: true,
                      }))
                    }
                    onCancelEdit={() => handleCancelEdit(row.feature)}
                    onDelete={() => handleDeleteFeature(row.feature)}
                    onToggleExpand={() =>
                      featureId &&
                      setExpandedFeatures((prev) => ({
                        ...prev,
                        [featureId]: !prev[featureId],
                      }))
                    }
                    onEditField={(update) =>
                      featureId &&
                      setFeatureEdits((prev) => ({
                        ...prev,
                        [featureId]: { ...prev[featureId], ...update },
                      }))
                    }
                    onNewRevision={(latestRevision) =>
                      setRevisionModal({ featureId, latestRevision })
                    }
                  />
                </div>
              )
            })}
        </div>
      </div>

      <CreateRevisionModal
        isOpen={revisionModal !== null}
        onClose={() => setRevisionModal(null)}
        featureId={revisionModal?.featureId ?? ''}
        latestRevision={revisionModal?.latestRevision}
      />
    </section>
  )
}
