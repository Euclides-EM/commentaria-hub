import { useMemo, useState } from 'react'
import { useQueries } from '@tanstack/react-query'
import type {
  annotation_Annotation,
  annotationrule_PipelineStage,
} from '@hub-api'
import { AnnotationsService } from '@hub-api'
import { useAppState } from '../../context/useAppState'
import { useDatasetsQuery } from '../../queries/datasets'
import { usePipelineStages } from '../../queries/metadata'
import { getStageDisplayName } from '../../utils/stages'
import { ExportAnnotationModal } from '../annotation/details/ExportAnnotationModal'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { MultiSelectDropdown } from '../core/MultiSelectDropdown'
import { SearchInput } from '../core/SearchInput'
import { Timestamp } from '../core/Timestamp'
import useLocalStorageState from 'use-local-storage-state'
import { TITLE_PAGES_DATASET_ID } from '../../utils/editions.ts'

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

export function AnnotationsTable() {
  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { data: stages } = usePipelineStages()
  const { setState } = useAppState()
  const [isExportOpen, setIsExportOpen] = useState(false)
  const [selectedTargets, setSelectedTargets] = useState<SelectedTargets>({})
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
  const allVisibleSelected =
    sortedRows.length > 0 &&
    sortedRows.every(
      (row) => effectiveSelectedTargets[rowKey(row)] !== undefined,
    )

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

  return (
    <div className="w-full h-full flex flex-col px-8">
      <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white gap-4 cursor-default">
        <div className="flex items-center gap-6 cursor-default">
          <div>
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
        <div className="flex items-center gap-2">
          <Button
            type="button"
            onClick={() => setIsExportOpen(true)}
            disabled={selectedCount === 0}
            className="px-3 py-1.5 text-sm"
          >
            Export selected ({selectedCount})
          </Button>
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
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {sortedRows.map((row) => (
                        <tr
                          key={`${row.datasetId}:${row.annotation.id}`}
                          className="hover:bg-gray-50 cursor-default"
                        >
                          <td className="px-4 py-3 text-left whitespace-nowrap">
                            <input
                              type="checkbox"
                              checked={
                                effectiveSelectedTargets[rowKey(row)] !==
                                undefined
                              }
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
                              ? getStageDisplayName(
                                  row.annotation.pipeline_stage,
                                )
                              : 'None'}
                          </td>
                          <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                            {row.annotation.ground_truth ? 'Yes' : 'No'}
                          </td>
                          <td className="px-4 py-3 text-gray-700 w-full">
                            {row.annotation.dataset_id ===
                            TITLE_PAGES_DATASET_ID
                              ? '-'
                              : row.annotation.pages || '-'}
                          </td>
                          <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                            <Timestamp
                              hideFullDate
                              date={
                                row.annotation.updated_at ||
                                row.annotation.created_at
                              }
                            />
                          </td>
                        </tr>
                      ))}
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
      />
    </div>
  )
}
