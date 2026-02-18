import { useMemo } from 'react'
import { useQueries } from '@tanstack/react-query'
import type { annotation_Annotation } from '../../api'
import { AnnotationsService } from '../../api'
import { useAppState } from '../../context/useAppState'
import { useDatasetsQuery } from '../../queries/datasets'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { SearchInput } from '../core/SearchInput'
import { Timestamp } from '../core/Timestamp'
import useLocalStorageState from 'use-local-storage-state'

type GroundTruthRow = {
  datasetId: string
  datasetName: string
  annotation: annotation_Annotation
}

type SortKey = 'annotation' | 'dataset' | 'updated'
type SortDirection = 'asc' | 'desc'

type SortConfig = {
  key: SortKey
  direction: SortDirection
}

export function GroundTruthsTable() {
  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const { setState } = useAppState()
  const [searchQuery, setSearchQuery] = useLocalStorageState<string>(
    'groundTruthsSearch',
    {
      defaultValue: '',
    },
  )
  const [sortConfig, setSortConfig] = useLocalStorageState<SortConfig>(
    'groundTruthsSort',
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

  const isLoadingAnnotations = annotationQueries.some((query) => query.isLoading)
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

    const result: GroundTruthRow[] = []
    annotationQueries.forEach((query, index) => {
      const datasetId = datasetIds[index]
      if (!datasetId || !query.data) {
        return
      }

      const datasetName = datasetNameById.get(datasetId) || datasetId
      query.data.forEach((annotation) => {
        if (annotation.id && annotation.ground_truth) {
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
    if (!trimmed) {
      return rows
    }
    return rows.filter((row) => {
      const haystack = [
        row.annotation.id,
        row.annotation.name,
        row.datasetId,
        row.datasetName,
        row.annotation.pages,
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
      return haystack.includes(trimmed)
    })
  }, [rows, searchQuery])

  const filteredDatasetCount = useMemo(
    () => new Set(filteredRows.map((row) => row.datasetId)).size,
    [filteredRows],
  )

  const sortedRows = useMemo(() => {
    const getSortValue = (row: GroundTruthRow, key: SortKey) => {
      switch (key) {
        case 'annotation':
          return (row.annotation.name || row.annotation.id || '').toLowerCase()
        case 'dataset':
          return row.datasetName.toLowerCase()
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
      <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white gap-4">
        <div className="flex items-center gap-6">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Ground Truths</h2>
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
            placeholder="Search ground truths..."
            className="w-[22rem] max-w-full"
          />
        </div>
      </div>

      <div className="overflow-auto px-2 py-4">
        <div className="flex flex-col">
          {(datasetsLoading || isLoadingAnnotations) && (
            <LoadingSpinner message="Loading ground truths..." />
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
              <div className="text-sm text-gray-500">
                No ground-truth annotations found.
              </div>
            )}

          {!datasetsLoading &&
            !isLoadingAnnotations &&
            !hasError &&
            rows.length > 0 &&
            filteredRows.length === 0 && (
              <div className="text-sm text-gray-500">
                No ground truths match "{searchQuery.trim()}".
              </div>
            )}

          {!datasetsLoading &&
            !isLoadingAnnotations &&
            !hasError &&
            filteredRows.length > 0 && (
              <div>
                <div className="overflow-auto rounded-lg border border-gray-200 bg-white">
                  <table className="min-w-full text-sm table-auto">
                    <thead className="bg-gray-50 text-xs text-gray-500">
                      <tr>
                        <th className="px-4 py-3 text-left whitespace-nowrap">
                          {renderSortHeader('Annotation', 'annotation')}
                        </th>
                        <th className="px-4 py-3 text-left whitespace-nowrap">
                          {renderSortHeader('Dataset', 'dataset')}
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
                          className="hover:bg-gray-50"
                        >
                          <td className="px-4 py-3 text-left whitespace-nowrap">
                            <button
                              type="button"
                              className="block w-full text-left font-medium text-teal-700 hover:text-teal-900 hover:underline"
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
                              className="block w-full text-left text-teal-700 hover:text-teal-900 hover:underline"
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
                          <td className="px-4 py-3 text-gray-700 w-full">
                            {row.annotation.pages || 'All'}
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
    </div>
  )
}
