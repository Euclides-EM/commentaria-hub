import { Fragment, useMemo, useState } from 'react'
import { useQueries } from '@tanstack/react-query'
import type {
  annotation_Annotation,
  annotation_Group,
  annotationrule_PipelineStage,
} from '@hub-api'
import { AnnotationsService } from '@hub-api'
import { useAppState } from '../../context/useAppState'
import {
  useAnnotationGroupsQuery,
  useCreateAnnotationGroupMutation,
  useDeleteAnnotationGroupMutation,
  useUpdateAnnotationGroupMutation,
} from '../../queries/annotationGroups'
import { useDatasetsQuery } from '../../queries/datasets'
import { usePipelineStages } from '../../queries/metadata'
import { getStageDisplayName } from '../../utils/stages'
import { ExportAnnotationModal } from '../annotation/details/ExportAnnotationModal'
import { CreateAnnotationGroupModal } from './CreateAnnotationGroupModal'
import { AddToAnnotationGroupsModal } from './AddToAnnotationGroupsModal'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { MultiSelectDropdown } from '../core/MultiSelectDropdown'
import { SearchInput } from '../core/SearchInput'
import { Timestamp } from '../core/Timestamp'
import useLocalStorageState from 'use-local-storage-state'
import { hasAnnotationPages } from '../../utils/editions.ts'
import { useAuthStore } from '../../store/authStore'

type AnnotationRow = {
  datasetId: string
  datasetName: string
  annotation: annotation_Annotation
}

type SelectedTarget = { datasetId: string; annotationId: string }
type SelectedTargets = Record<string, SelectedTarget>

type SortKey = 'annotation' | 'dataset' | 'stage' | 'updated'
type SortDirection = 'asc' | 'desc'

type SortConfig = {
  key: SortKey
  direction: SortDirection
}

type GroundTruthFilter = 'all' | 'true' | 'false'
type HiddenFilter = 'all' | 'true' | 'false'
type GroupSection = {
  group: annotation_Group | null
  rows: AnnotationRow[]
}
type AnnotationGroupWithId = annotation_Group & { id: string }

export function AnnotationsTable() {
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: stages } = usePipelineStages()
  const {
    data: annotationGroups = [],
    isLoading: annotationGroupsLoading,
    error: annotationGroupsError,
  } = useAnnotationGroupsQuery()
  const createGroupMutation = useCreateAnnotationGroupMutation()
  const updateGroupMutation = useUpdateAnnotationGroupMutation()
  const deleteGroupMutation = useDeleteAnnotationGroupMutation()
  const { setState } = useAppState()
  const [isExportOpen, setIsExportOpen] = useState(false)
  const [isCreateGroupOpen, setIsCreateGroupOpen] = useState(false)
  const [isAddToGroupsOpen, setIsAddToGroupsOpen] = useState(false)
  const [collapsedSections, setCollapsedSections] = useState<
    Record<string, boolean>
  >({})
  const [selectedTargets, setSelectedTargets] = useState<SelectedTargets>({})
  const [groupActionError, setGroupActionError] = useState<string | null>(null)
  const [searchQuery, setSearchQuery] = useLocalStorageState<string>(
    'annotationsSearch',
    {
      defaultValue: '',
    },
  )
  const [selectedStages, setSelectedStages] = useLocalStorageState<
    annotationrule_PipelineStage[] | null
  >('annotationsFilterStages', {
    defaultValue: null,
  })
  const [groundTruthFilter, setGroundTruthFilter] =
    useLocalStorageState<GroundTruthFilter>('annotationsGroundTruthFilter', {
      defaultValue: 'all',
    })
  const [hiddenFilter, setHiddenFilter] = useLocalStorageState<HiddenFilter>(
    'annotationsHiddenFilter',
    {
      defaultValue: 'false',
    },
  )
  const [sortConfig, setSortConfig] = useLocalStorageState<SortConfig>(
    'annotationsSort',
    {
      defaultValue: {
        key: 'updated',
        direction: 'desc',
      },
    },
  )
  const [groupingEnabled, setGroupingEnabled] = useLocalStorageState<boolean>(
    'annotationsGroupingEnabled',
    {
      defaultValue: false,
    },
  )

  const datasetIds = useMemo(
    () =>
      (datasets || []).flatMap((dataset) => (dataset.id ? [dataset.id] : [])),
    [datasets],
  )

  const annotationQueries = useQueries({
    queries: datasetIds.map((datasetId) => ({
      queryKey: ['annotations', datasetId] as const,
      queryFn: () =>
        AnnotationsService.getDatasetsAnnotations({
          dataSetId: datasetId,
        }),
      enabled: datasetIds.length > 0,
    })),
  })

  const isLoadingAnnotations = annotationQueries.some(
    (query) => query.isLoading,
  )
  const hasError = annotationQueries.some((query) => query.isError)
  const queryErrorMessage =
    annotationQueries.find((query) => query.error)?.error?.toString() || null

  const rows = useMemo(() => {
    const datasetNameById = new Map<string, string>()
    datasets?.forEach((dataset) => {
      if (dataset.id) {
        datasetNameById.set(dataset.id, dataset.name || dataset.id)
      }
    })

    const result: AnnotationRow[] = []
    annotationQueries.forEach((query, index) => {
      const datasetId = datasetIds[index]
      if (!datasetId || !query.data) {
        return
      }

      const datasetName = datasetNameById.get(datasetId) || datasetId
      query.data.forEach((annotation) => {
        if (annotation.id) {
          result.push({
            datasetId,
            datasetName,
            annotation,
          })
        }
      })
    })

    return result.sort((a, b) => {
      const aTime = a.annotation.updated_at || a.annotation.created_at
      const bTime = b.annotation.updated_at || b.annotation.created_at
      return new Date(bTime || 0).getTime() - new Date(aTime || 0).getTime()
    })
  }, [annotationQueries, datasetIds, datasets])

  const filteredRows = useMemo(() => {
    const trimmed = searchQuery.trim().toLowerCase()
    return rows.filter((row) => {
      const stage = row.annotation.pipeline_stage
      const matchesStage =
        selectedStages == null ||
        (stage != null && selectedStages.includes(stage))
      if (!matchesStage) {
        return false
      }
      const matchesGroundTruth =
        groundTruthFilter === 'all' ||
        (groundTruthFilter === 'true'
          ? !!row.annotation.ground_truth
          : !row.annotation.ground_truth)
      if (!matchesGroundTruth) {
        return false
      }
      const matchesHidden =
        hiddenFilter === 'all' ||
        (hiddenFilter === 'true'
          ? !!row.annotation.hidden
          : !row.annotation.hidden)
      if (!matchesHidden) {
        return false
      }
      if (!trimmed) {
        return true
      }
      const haystack = [
        row.annotation.id,
        row.annotation.name,
        row.annotation.description,
        row.datasetId,
        row.datasetName,
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
      return haystack.includes(trimmed)
    })
  }, [groundTruthFilter, hiddenFilter, rows, searchQuery, selectedStages])

  const filteredDatasetCount = useMemo(
    () => new Set(filteredRows.map((row) => row.datasetId)).size,
    [filteredRows],
  )

  const rowKey = (row: AnnotationRow) => `${row.datasetId}:${row.annotation.id}`

  const validRowKeys = useMemo(
    () => new Set(rows.map((row) => rowKey(row))),
    [rows],
  )

  const pruneSelectedTargets = (targets: SelectedTargets): SelectedTargets => {
    const next: SelectedTargets = {}
    for (const [key, target] of Object.entries(targets)) {
      if (validRowKeys.has(key)) {
        next[key] = target
      }
    }
    return next
  }

  const sortedRows = useMemo(() => {
    const getSortValue = (row: AnnotationRow, key: SortKey) => {
      switch (key) {
        case 'annotation':
          return (row.annotation.name || row.annotation.id || '').toLowerCase()
        case 'dataset':
          return row.datasetName.toLowerCase()
        case 'stage':
          return row.annotation.pipeline_stage
            ? getStageDisplayName(row.annotation.pipeline_stage).toLowerCase()
            : ''
        case 'updated': {
          const raw = row.annotation.updated_at || row.annotation.created_at
          const time = raw ? new Date(raw).getTime() : 0
          return Number.isNaN(time) ? 0 : time
        }
        default:
          return ''
      }
    }

    const data = [...filteredRows]
    data.sort((a, b) => {
      const aValue = getSortValue(a, sortConfig.key)
      const bValue = getSortValue(b, sortConfig.key)
      if (aValue < bValue) return sortConfig.direction === 'asc' ? -1 : 1
      if (aValue > bValue) return sortConfig.direction === 'asc' ? 1 : -1
      return 0
    })
    return data
  }, [filteredRows, sortConfig.direction, sortConfig.key])

  const effectiveSelectedTargets = pruneSelectedTargets(selectedTargets)

  const selectedCount = Object.keys(effectiveSelectedTargets).length
  const selectedRows = Object.values(effectiveSelectedTargets)
  const selectedReferences = useMemo(
    () =>
      selectedRows.map((target) => ({
        dataset_id: target.datasetId,
        id: target.annotationId,
      })),
    [selectedRows],
  )
  const allVisibleSelected =
    sortedRows.length > 0 &&
    sortedRows.every(
      (row) => effectiveSelectedTargets[rowKey(row)] !== undefined,
    )
  const isGroupMutationPending =
    createGroupMutation.isPending ||
    updateGroupMutation.isPending ||
    deleteGroupMutation.isPending

  const sortedGroups = useMemo(
    () =>
      [...annotationGroups].sort((a, b) =>
        (a.name || a.id || '').localeCompare(b.name || b.id || ''),
      ),
    [annotationGroups],
  )
  const groupOptions = useMemo<AnnotationGroupWithId[]>(
    () =>
      sortedGroups.filter(
        (group): group is AnnotationGroupWithId =>
          typeof group.id === 'string' && group.id.length > 0,
      ),
    [sortedGroups],
  )
  const groupOptionItems = useMemo(
    () =>
      groupOptions.map((group) => ({
        id: group.id,
        label: group.name || group.id,
      })),
    [groupOptions],
  )

  const groupsByRowKey = useMemo(() => {
    const map = new Map<string, annotation_Group[]>()
    sortedGroups.forEach((group) => {
      ;(group.annotations || []).forEach((ref) => {
        if (!ref.dataset_id || !ref.id) {
          return
        }
        const key = `${ref.dataset_id}:${ref.id}`
        const existing = map.get(key)
        if (existing) {
          existing.push(group)
          return
        }
        map.set(key, [group])
      })
    })
    return map
  }, [sortedGroups])

  const groupedSections = useMemo<GroupSection[]>(() => {
    const rowsByGroupId = new Map<string, AnnotationRow[]>()
    const ungrouped: AnnotationRow[] = []
    sortedRows.forEach((row) => {
      const groups = (groupsByRowKey.get(rowKey(row)) || []).filter(
        (group) => !!group.id,
      )
      if (!groups.length) {
        ungrouped.push(row)
        return
      }
      groups.forEach((group) => {
        if (!group.id) {
          return
        }
        const existingRows = rowsByGroupId.get(group.id)
        if (existingRows) {
          existingRows.push(row)
          return
        }
        rowsByGroupId.set(group.id, [row])
      })
    })
    const result: GroupSection[] = sortedGroups
      .filter((group) => group.id && (rowsByGroupId.get(group.id) || []).length)
      .map((group) => ({
        group,
        rows: rowsByGroupId.get(group.id!) || [],
      }))
    if (ungrouped.length > 0) {
      result.push({
        group: null,
        rows: ungrouped,
      })
    }
    return result
  }, [groupsByRowKey, sortedGroups, sortedRows])

  const mergeAnnotationReferences = (
    existing: Array<{ dataset_id?: string; id?: string }>,
    additions: Array<{ dataset_id?: string; id?: string }>,
  ) => {
    const merged = new Map<string, { dataset_id: string; id: string }>()
    ;[...existing, ...additions].forEach((ref) => {
      if (!ref.dataset_id || !ref.id) {
        return
      }
      merged.set(`${ref.dataset_id}:${ref.id}`, {
        dataset_id: ref.dataset_id,
        id: ref.id,
      })
    })
    return [...merged.values()]
  }

  const getErrorMessage = (value: unknown) =>
    value instanceof Error ? value.message : String(value)

  const handleAddSelectedToGroups = async (groupIds: string[]) => {
    if (groupIds.length === 0 || selectedReferences.length === 0) {
      return
    }
    try {
      setGroupActionError(null)
      for (const groupId of groupIds) {
        const group = groupOptions.find((item) => item.id === groupId)
        if (!group) {
          continue
        }
        await updateGroupMutation.mutateAsync({
          groupId: group.id,
          group: {
            ...group,
            annotations: mergeAnnotationReferences(
              group.annotations || [],
              selectedReferences,
            ),
          },
        })
      }
      setSelectedTargets({})
      setIsAddToGroupsOpen(false)
    } catch (error) {
      setGroupActionError(getErrorMessage(error))
    }
  }

  const handleCreateGroupFromSelected = async ({
    name,
    description,
  }: {
    name: string
    description?: string
  }) => {
    if (selectedReferences.length === 0) {
      setGroupActionError('Select at least one annotation.')
      return
    }
    try {
      setGroupActionError(null)
      await createGroupMutation.mutateAsync({
        name,
        description,
        annotations: selectedReferences,
      })
      setSelectedTargets({})
      setIsCreateGroupOpen(false)
    } catch (error) {
      setGroupActionError(getErrorMessage(error))
    }
  }

  const handleUngroup = async (group: annotation_Group | null) => {
    if (!group?.id) {
      return
    }
    try {
      setGroupActionError(null)
      await deleteGroupMutation.mutateAsync(group.id)
    } catch (error) {
      setGroupActionError(getErrorMessage(error))
    }
  }

  const handleRemoveFromGroup = async (
    row: AnnotationRow,
    group: annotation_Group | null,
  ) => {
    if (!group?.id || !row.annotation.id) {
      return
    }
    try {
      setGroupActionError(null)
      await updateGroupMutation.mutateAsync({
        groupId: group.id,
        group: {
          ...group,
          annotations: (group.annotations || []).filter(
            (ref) =>
              !(
                ref.dataset_id === row.datasetId && ref.id === row.annotation.id
              ),
          ),
        },
      })
    } catch (error) {
      setGroupActionError(getErrorMessage(error))
    }
  }

  const getSectionKey = (section: GroupSection) =>
    section.group?.id || '__annotations-ungrouped'

  const isSectionCollapsed = (section: GroupSection) =>
    collapsedSections[getSectionKey(section)] === true

  const toggleSectionCollapsed = (section: GroupSection) => {
    const key = getSectionKey(section)
    setCollapsedSections((current) => ({
      ...current,
      [key]: !current[key],
    }))
  }

  const toggleSort = (key: SortKey) => {
    setSortConfig((current) => {
      if (current.key === key) {
        return {
          key,
          direction: current.direction === 'asc' ? 'desc' : 'asc',
        }
      }
      return {
        key,
        direction: 'asc',
      }
    })
  }

  const renderSortHeader = (label: string, key: SortKey) => {
    const isActive = sortConfig.key === key
    const arrow = isActive ? (sortConfig.direction === 'asc' ? '▲' : '▼') : null
    return (
      <button
        type="button"
        onClick={() => toggleSort(key)}
        className={`inline-flex items-center gap-1 ${isActive ? 'text-gray-800' : 'text-gray-500 hover:text-gray-700'}`}
      >
        <span>{label}</span>
        {arrow && <span className="text-[10px]">{arrow}</span>}
      </button>
    )
  }

  const showSelectionControls = isAuthenticated
  const showGroupMutationControls = isAuthenticated && groupingEnabled
  const visibleColumnCount = [
    showSelectionControls ? 'select' : null,
    'annotation',
    'dataset',
    'stage',
    'groundTruth',
    'pages',
    'updated',
    showGroupMutationControls ? 'groupAction' : null,
  ].filter(Boolean).length

  const renderRow = (row: AnnotationRow, group: annotation_Group | null) => (
    <tr
      key={
        groupingEnabled
          ? `${group?.id || 'ungrouped'}:${row.datasetId}:${row.annotation.id}`
          : `${row.datasetId}:${row.annotation.id}`
      }
      className="hover:bg-gray-50 cursor-default"
    >
      {showSelectionControls && (
        <td className="px-4 py-3 text-left whitespace-nowrap">
          <input
            type="checkbox"
            checked={effectiveSelectedTargets[rowKey(row)] !== undefined}
            onChange={(e) => {
              if (!row.annotation.id) {
                return
              }
              const key = rowKey(row)
              setSelectedTargets((current) => {
                const next = pruneSelectedTargets(current)
                if (e.target.checked) {
                  return {
                    ...next,
                    [key]: {
                      datasetId: row.datasetId,
                      annotationId: row.annotation.id!,
                    },
                  }
                }
                delete next[key]
                return next
              })
            }}
            className="h-4 w-4"
            aria-label={`Select ${row.annotation.name || row.annotation.id}`}
          />
        </td>
      )}
      <td className="px-4 py-3 text-left whitespace-nowrap">
        <button
          type="button"
          className="inline text-left font-medium text-teal-700 hover:text-teal-900 hover:underline cursor-pointer"
          onClick={() =>
            setState({
              datasetId: row.datasetId,
              annotationId: row.annotation.id || '',
            })
          }
        >
          {row.annotation.name || row.annotation.id}
        </button>
      </td>
      <td className="px-4 py-3 text-left whitespace-nowrap">
        <button
          type="button"
          className="inline text-left text-teal-700 hover:text-teal-900 hover:underline cursor-pointer"
          onClick={() =>
            setState({
              datasetId: row.datasetId,
              annotationId: '',
            })
          }
        >
          {row.datasetName}
        </button>
      </td>
      <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
        {row.annotation.pipeline_stage
          ? getStageDisplayName(row.annotation.pipeline_stage)
          : 'None'}
      </td>
      <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
        {row.annotation.ground_truth ? 'Yes' : 'No'}
      </td>
      <td className="px-4 py-3 text-gray-700 w-full">
        {hasAnnotationPages(row.annotation) ? row.annotation.pages : '-'}
      </td>
      <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
        <Timestamp
          hideFullDate
          date={row.annotation.updated_at || row.annotation.created_at}
        />
      </td>
      {showGroupMutationControls && (
        <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
          {group?.id ? (
            <Button
              type="button"
              onClick={() => void handleRemoveFromGroup(row, group)}
              disabled={isGroupMutationPending}
              className="px-2 py-1 text-xs"
            >
              Remove
            </Button>
          ) : (
            <span className="text-gray-400">-</span>
          )}
        </td>
      )}
    </tr>
  )

  return (
    <div className="w-full h-full flex flex-col px-8">
      <div className="px-6 py-4 border-b border-gray-200 bg-white cursor-default">
        <div className="flex flex-wrap items-center gap-4">
          <div className="mr-2">
            <h2 className="text-lg font-semibold text-gray-900">Annotations</h2>
            <p className="text-xs text-gray-500">
              {filteredRows.length}
              {searchQuery && ` of ${rows.length}`}{' '}
              {rows.length === 1 ? 'annotation' : 'annotations'} from{' '}
              {filteredDatasetCount}{' '}
              {filteredDatasetCount === 1 ? 'dataset' : 'datasets'}
            </p>
          </div>
          <SearchInput
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Search annotations..."
            className="w-[22rem] max-w-full"
          />
          {stages && (
            <MultiSelectDropdown
              allItems={stages}
              selectedItems={selectedStages}
              setSelectedItems={setSelectedStages}
              itemsLabel="stages"
              getItemLabel={(stage) => getStageDisplayName(stage)}
            />
          )}
          <label className="flex items-center gap-2 text-sm text-gray-700 cursor-default">
            <span>Ground truth</span>
            <select
              value={groundTruthFilter}
              onChange={(e) =>
                setGroundTruthFilter(e.target.value as GroundTruthFilter)
              }
              className="h-8 rounded-md border border-gray-400 bg-white px-2 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100 cursor-pointer"
              aria-label="Filter by ground truth"
            >
              <option value="all">All</option>
              <option value="true">Ground truth only</option>
              <option value="false">Not ground truth</option>
            </select>
          </label>
          <label className="flex items-center gap-2 text-sm text-gray-700 cursor-default">
            <span>Hidden</span>
            <select
              value={hiddenFilter}
              onChange={(e) => setHiddenFilter(e.target.value as HiddenFilter)}
              className="h-8 rounded-md border border-gray-400 bg-white px-2 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100 cursor-pointer"
              aria-label="Filter by hidden"
            >
              <option value="false">Not hidden</option>
              <option value="all">All</option>
              <option value="true">Hidden only</option>
            </select>
          </label>
        </div>
        <div className="mt-3 flex flex-wrap items-center gap-2">
          <label className="flex items-center gap-2 text-sm text-gray-700 cursor-pointer">
            <input
              type="checkbox"
              checked={groupingEnabled}
              onChange={(e) => setGroupingEnabled(e.target.checked)}
              className="h-4 w-4"
            />
            Group by groups
          </label>
          {isAuthenticated && (
            <>
              {groupOptions.length > 0 && (
                <Button
                  type="button"
                  onClick={() => {
                    setGroupActionError(null)
                    setIsAddToGroupsOpen(true)
                  }}
                  disabled={selectedCount === 0 || isGroupMutationPending}
                  className="px-3 py-1.5 text-sm"
                >
                  Add to groups
                </Button>
              )}
              <Button
                type="button"
                onClick={() => {
                  setGroupActionError(null)
                  setIsCreateGroupOpen(true)
                }}
                disabled={selectedCount === 0 || isGroupMutationPending}
                className="px-3 py-1.5 text-sm"
              >
                New group
              </Button>
              <Button
                type="button"
                onClick={() => setIsExportOpen(true)}
                disabled={selectedCount === 0}
                className="px-3 py-1.5 text-sm"
              >
                Export selected ({selectedCount})
              </Button>
            </>
          )}
        </div>
      </div>

      <div className="overflow-auto px-2 py-4">
        <div className="flex flex-col">
          {(datasetsLoading || isLoadingAnnotations) && (
            <LoadingSpinner message="Loading annotations..." />
          )}

          {!datasetsLoading && !isLoadingAnnotations && hasError && (
            <div className="mb-4">
              <ErrorMessage
                message={
                  queryErrorMessage ||
                  "Failed to load one or more datasets' annotations."
                }
              />
            </div>
          )}
          {((annotationGroupsError &&
            !annotationGroupsLoading &&
            !groupActionError) ||
            groupActionError) && (
            <div className="mb-4">
              <ErrorMessage
                message={
                  groupActionError ||
                  (annotationGroupsError instanceof Error
                    ? annotationGroupsError.message
                    : String(annotationGroupsError))
                }
              />
            </div>
          )}

          {!datasetsLoading &&
            !isLoadingAnnotations &&
            !hasError &&
            rows.length === 0 && (
              <div className="text-sm text-gray-500">No annotations found.</div>
            )}

          {!datasetsLoading &&
            !isLoadingAnnotations &&
            !hasError &&
            rows.length > 0 &&
            filteredRows.length === 0 && (
              <div className="text-sm text-gray-500">
                No annotations match "{searchQuery.trim()}".
              </div>
            )}

          {!datasetsLoading &&
            !isLoadingAnnotations &&
            !hasError &&
            filteredRows.length > 0 && (
              <div>
                <div className="overflow-auto rounded-lg border border-gray-200 bg-white">
                  <table className="min-w-full text-sm table-auto cursor-default">
                    <thead className="bg-gray-50 text-xs text-gray-500">
                      <tr>
                        {showSelectionControls && (
                          <th className="px-4 py-3 text-left whitespace-nowrap">
                            <input
                              type="checkbox"
                              checked={allVisibleSelected}
                              onChange={(e) => {
                                if (e.target.checked) {
                                  setSelectedTargets((current) => {
                                    const next = pruneSelectedTargets(current)
                                    sortedRows.forEach((row) => {
                                      if (row.annotation.id) {
                                        next[rowKey(row)] = {
                                          datasetId: row.datasetId,
                                          annotationId: row.annotation.id,
                                        }
                                      }
                                    })
                                    return next
                                  })
                                  return
                                }
                                const visibleKeys = new Set(
                                  sortedRows.map((row) => rowKey(row)),
                                )
                                setSelectedTargets((current) => {
                                  const next = pruneSelectedTargets(current)
                                  visibleKeys.forEach((key) => {
                                    delete next[key]
                                  })
                                  return next
                                })
                              }}
                              className="h-4 w-4"
                              aria-label="Select all visible annotations"
                            />
                          </th>
                        )}
                        <th className="px-4 py-3 text-left whitespace-nowrap">
                          {renderSortHeader('Annotation', 'annotation')}
                        </th>
                        <th className="px-4 py-3 text-left whitespace-nowrap">
                          {renderSortHeader('Dataset', 'dataset')}
                        </th>
                        <th className="px-4 py-3 text-left whitespace-nowrap">
                          {renderSortHeader('Stage', 'stage')}
                        </th>
                        <th className="px-4 py-3 text-left whitespace-nowrap">
                          Ground Truth
                        </th>
                        <th className="px-4 py-3 text-left">Pages</th>
                        <th className="px-4 py-3 text-left whitespace-nowrap">
                          {renderSortHeader('Updated', 'updated')}
                        </th>
                        {showGroupMutationControls && (
                          <th className="px-4 py-3 text-left whitespace-nowrap">
                            Group action
                          </th>
                        )}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {groupingEnabled
                        ? groupedSections.map((section) => (
                            <Fragment
                              key={
                                section.group?.id || '__annotations-ungrouped'
                              }
                            >
                              <tr className="bg-gray-100">
                                <td
                                  colSpan={visibleColumnCount}
                                  className="px-4 py-2 text-xs text-gray-700"
                                >
                                  <div className="flex items-center justify-between gap-3">
                                    <div className="flex items-center gap-2 min-w-0">
                                      <button
                                        type="button"
                                        onClick={() =>
                                          toggleSectionCollapsed(section)
                                        }
                                        className="h-6 w-6 inline-flex items-center justify-center rounded border border-gray-300 bg-white text-gray-700 hover:bg-gray-50"
                                        aria-label={
                                          isSectionCollapsed(section)
                                            ? 'Expand group'
                                            : 'Collapse group'
                                        }
                                        title={
                                          isSectionCollapsed(section)
                                            ? 'Expand'
                                            : 'Collapse'
                                        }
                                      >
                                        <span className="text-sm leading-none">
                                          {isSectionCollapsed(section)
                                            ? '▶'
                                            : '▼'}
                                        </span>
                                      </button>
                                      <div className="font-semibold truncate">
                                        {section.group
                                          ? section.group.name ||
                                            section.group.id
                                          : 'Ungrouped'}
                                        <span className="ml-2 font-normal text-gray-500">
                                          ({section.rows.length})
                                        </span>
                                      </div>
                                    </div>
                                    {showGroupMutationControls &&
                                      section.group && (
                                        <Button
                                          type="button"
                                          variant="danger"
                                          onClick={() =>
                                            void handleUngroup(section.group)
                                          }
                                          disabled={
                                            isGroupMutationPending ||
                                            !section.group.id
                                          }
                                          className="px-2 py-1 text-xs shrink-0"
                                        >
                                          Ungroup
                                        </Button>
                                      )}
                                  </div>
                                </td>
                              </tr>
                              {!isSectionCollapsed(section) &&
                                section.rows.map((row) =>
                                  renderRow(row, section.group),
                                )}
                            </Fragment>
                          ))
                        : sortedRows.map((row) => renderRow(row, null))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
        </div>
      </div>
      <ExportAnnotationModal
        isOpen={isExportOpen}
        onClose={() => setIsExportOpen(false)}
        exportTargets={selectedRows}
        onSuccess={() => setSelectedTargets({})}
      />
      <CreateAnnotationGroupModal
        isOpen={isCreateGroupOpen}
        selectedCount={selectedCount}
        isSubmitting={createGroupMutation.isPending}
        error={groupActionError}
        onClose={() => {
          setGroupActionError(null)
          setIsCreateGroupOpen(false)
        }}
        onCreate={handleCreateGroupFromSelected}
      />
      <AddToAnnotationGroupsModal
        isOpen={isAddToGroupsOpen}
        selectedCount={selectedCount}
        groups={groupOptionItems}
        isSubmitting={updateGroupMutation.isPending}
        error={groupActionError}
        onClose={() => {
          setGroupActionError(null)
          setIsAddToGroupsOpen(false)
        }}
        onSubmit={handleAddSelectedToGroups}
      />
    </div>
  )
}
