import { useMemo } from 'react'
import type { integration_Job } from '@hub-api'
import { useAppState } from '../../context/useAppState'
import { useNonCompletedIntegrationJobsQuery } from '../../queries/integrations'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { SearchInput } from '../core/SearchInput'
import { Timestamp } from '../core/Timestamp'
import useLocalStorageState from 'use-local-storage-state'

type SortKey = 'job' | 'status' | 'updated'
type SortDirection = 'asc' | 'desc'

type SortConfig = {
  key: SortKey
  direction: SortDirection
}

export function JobsTable() {
  const { setState } = useAppState()
  const { data: jobs, isLoading, error } = useNonCompletedIntegrationJobsQuery()
  const [searchQuery, setSearchQuery] = useLocalStorageState<string>(
    'jobsSearch',
    {
      defaultValue: '',
    },
  )
  const [sortConfig, setSortConfig] = useLocalStorageState<SortConfig>(
    'jobsSort',
    {
      defaultValue: {
        key: 'updated',
        direction: 'desc',
      },
    },
  )

  const rows = useMemo(() => jobs ?? [], [jobs])

  const filteredRows = useMemo(() => {
    const trimmed = searchQuery.trim().toLowerCase()
    if (!trimmed) {
      return rows
    }
    return rows.filter((job) => {
      const haystack = [
        job.id,
        job.name,
        job.description,
        job.status,
        job.task,
        job.annotation?.dataset_id,
        job.annotation?.id,
        job.target?.platform,
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
      return haystack.includes(trimmed)
    })
  }, [rows, searchQuery])

  const sortedRows = useMemo(() => {
    const getSortValue = (job: integration_Job, key: SortKey) => {
      switch (key) {
        case 'job':
          return (job.name || job.id || '').toLowerCase()
        case 'status':
          return (job.status || '').toLowerCase()
        case 'updated': {
          const raw = job.updated_at || job.created_at
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
            <h2 className="text-lg font-semibold text-gray-900">Jobs</h2>
            <p className="text-xs text-gray-500">
              {filteredRows.length}
              {searchQuery && ` of ${rows.length}`}{' '}
              {rows.length === 1 ? 'job' : 'jobs'}
            </p>
          </div>
          <SearchInput
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Search jobs..."
            className="w-[22rem] max-w-full"
          />
        </div>
      </div>

      <div className="overflow-auto px-2 py-4">
        <div className="flex flex-col">
          {isLoading && <LoadingSpinner message="Loading jobs..." />}

          {!isLoading && error && (
            <div className="mb-4">
              <ErrorMessage message={String(error)} />
            </div>
          )}

          {!isLoading && !error && rows.length === 0 && (
            <div className="text-sm text-gray-500">No active jobs found.</div>
          )}

          {!isLoading &&
            !error &&
            rows.length > 0 &&
            filteredRows.length === 0 && (
              <div className="text-sm text-gray-500">
                No jobs match "{searchQuery.trim()}".
              </div>
            )}

          {!isLoading && !error && filteredRows.length > 0 && (
            <div>
              <div className="overflow-auto rounded-lg border border-gray-200 bg-white">
                <table className="min-w-full text-sm table-auto">
                  <thead className="bg-gray-50 text-xs text-gray-500">
                    <tr>
                      <th className="px-4 py-3 text-left whitespace-nowrap">
                        {renderSortHeader('Job', 'job')}
                      </th>
                      <th className="px-4 py-3 text-left whitespace-nowrap">
                        Job details
                      </th>
                      <th className="px-4 py-3 text-left whitespace-nowrap">
                        Annotation
                      </th>
                      <th className="px-4 py-3 text-left whitespace-nowrap">
                        {renderSortHeader('Status', 'status')}
                      </th>
                      <th className="px-4 py-3 text-left whitespace-nowrap">
                        Task
                      </th>
                      <th className="px-4 py-3 text-left whitespace-nowrap">
                        Target
                      </th>
                      <th className="px-4 py-3 text-left whitespace-nowrap">
                        {renderSortHeader('Updated', 'updated')}
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {sortedRows.map((job, index) => (
                      <tr
                        key={`${job.id || job.created_at || 'job'}-${index}`}
                        className="hover:bg-gray-50"
                      >
                        <td className="px-4 py-3 text-left whitespace-nowrap">
                          <div className="font-medium text-gray-900">
                            {job.name || job.id || 'Untitled'}
                          </div>
                          {job.description && (
                            <div className="text-xs text-gray-500">
                              {job.description}
                            </div>
                          )}
                        </td>
                        <td className="px-4 py-3 text-left text-gray-700">
                          {job.details || '-'}
                        </td>
                        <td className="px-4 py-3 text-left whitespace-nowrap">
                          {job.annotation?.dataset_id && job.annotation?.id ? (
                            <button
                              type="button"
                              className="text-teal-700 hover:text-teal-900 hover:underline"
                              onClick={() =>
                                setState({
                                  datasetId: job.annotation?.dataset_id || '',
                                  annotationId: job.annotation?.id || '',
                                })
                              }
                            >
                              {job.annotation.id}
                            </button>
                          ) : (
                            <span className="text-gray-400">-</span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                          {job.status || '-'}
                        </td>
                        <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                          {job.task || '-'}
                        </td>
                        <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                          {job.target?.platform || '-'}
                        </td>
                        <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                          <Timestamp
                            hideFullDate
                            date={job.updated_at || job.created_at}
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
