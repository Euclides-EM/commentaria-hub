import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import Select from 'react-select'
import { FeatureResultsService, type feature_Result } from '@hub-api'
import { useAppState } from '../../../context/useAppState.ts'
import { SearchInput } from '../../core/SearchInput.tsx'
import { ErrorMessage } from '../../core/ErrorMessage.tsx'
import { useDatasetFeaturesQuery } from '../../../queries/datasets.ts'
import { selectStyles } from '../../../styles/selectStyles.ts'
import { formatEditionLabel } from '../../../utils/editions.ts'
import {
  listAllEditions,
  mapEditionsToItems,
} from '../../../utils/editionItems.ts'

type FeatureOption = {
  value: string
  label: string
  color: string
}

type EditionOption = {
  value: string
  label: string
}

type FeatureResultRow = {
  result: feature_Result
  featureName: string
  featureColor: string
  featureDescription: string
  featureRevision: string
  editionDetails: string
  value: string
}

type SortKey =
  | 'pageKey'
  | 'editionDetails'
  | 'featureName'
  | 'featureDescription'

type SortDirection = 'asc' | 'desc'

const normalizeText = (value: string | null | undefined) => value?.trim() || ''

const normalizeSearchValue = (value: string | null | undefined) =>
  normalizeText(value).toLowerCase()

const formatValue = (result: feature_Result) =>
  (result.values ?? [])
    .map((item) => normalizeText(item.surface))
    .filter(Boolean)
    .join(', ')

const formatCount = (value: number) => value.toLocaleString()

export function FeatureResultsTab() {
  const { state } = useAppState()
  const datasetId = state.datasetId
  const annotationId = state.annotationId
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedFeatureOption, setSelectedFeatureOption] =
    useState<FeatureOption | null>(null)
  const [selectedEditionOption, setSelectedEditionOption] =
    useState<EditionOption | null>(null)
  const [sortKey, setSortKey] = useState<SortKey>('pageKey')
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc')

  const featuresQuery = useDatasetFeaturesQuery(datasetId, !!datasetId)
  const featureResultsQuery = useQuery({
    queryKey: ['featureResults', datasetId, annotationId],
    queryFn: () =>
      FeatureResultsService.getDatasetsAnnotationsResults({
        dataSetId: datasetId,
        id: annotationId,
      }),
    enabled: !!datasetId && !!annotationId,
    refetchOnWindowFocus: false,
  })
  const editionsQuery = useQuery({
    queryKey: ['editions', 'all', 'items'],
    queryFn: async () => mapEditionsToItems(await listAllEditions()),
    refetchOnWindowFocus: false,
  })

  const featureOptions = useMemo(
    () =>
      [...(featuresQuery.data ?? [])]
        .filter((feature): feature is { id: string; name?: string | null } =>
          Boolean(feature.id),
        )
        .sort((left, right) =>
          (left.name || left.id).localeCompare(
            right.name || right.id,
            undefined,
            {
              sensitivity: 'base',
            },
          ),
        )
        .map((feature) => ({
          value: feature.id,
          label: feature.name || feature.id,
          color: feature.color || '',
        })),
    [featuresQuery.data],
  )

  const featureById = useMemo(() => {
    const map = new Map<
      string,
      {
        name: string
        color: string
        description: string
      }
    >()

    for (const feature of featuresQuery.data ?? []) {
      if (!feature.id) continue
      map.set(feature.id, {
        name: feature.name || feature.id,
        color: feature.color || '',
        description: feature.description || '',
      })
    }

    return map
  }, [featuresQuery.data])

  const latestRevisionByFeatureId = useMemo(() => {
    const map = new Map<string, string>()

    for (const feature of featuresQuery.data ?? []) {
      if (!feature.id) continue
      const latestRevision = [...(feature.revisions ?? [])].sort(
        (left, right) => {
          const leftTime = left.created_at
            ? new Date(left.created_at).getTime()
            : 0
          const rightTime = right.created_at
            ? new Date(right.created_at).getTime()
            : 0
          return rightTime - leftTime
        },
      )[0]

      if (latestRevision?.id) {
        map.set(feature.id, latestRevision.id)
      }
    }

    return map
  }, [featuresQuery.data])

  const editionDetailsByKey = useMemo(() => {
    const map = new Map<string, string>()
    for (const item of editionsQuery.data ?? []) {
      map.set(item.key, formatEditionLabel(item))
    }
    return map
  }, [editionsQuery.data])

  const rows = useMemo<FeatureResultRow[]>(
    () =>
      (featureResultsQuery.data ?? []).map((result) => {
        const featureId = result.feature_id || ''
        const feature = featureById.get(featureId)
        return {
          result,
          featureName: feature?.name || '',
          featureColor: feature?.color || '',
          featureDescription: feature?.description || '',
          featureRevision:
            normalizeText(result.source?.revision) ||
            normalizeText(latestRevisionByFeatureId.get(featureId)) ||
            '',
          editionDetails: editionDetailsByKey.get(result.page_key || '') || '',
          value: formatValue(result),
        }
      }),
    [
      editionDetailsByKey,
      featureById,
      featureResultsQuery.data,
      latestRevisionByFeatureId,
    ],
  )

  const filteredRows = useMemo(() => {
    const trimmedQuery = normalizeSearchValue(searchQuery)
    return rows.filter((row) => {
      if (
        selectedFeatureOption &&
        row.result.feature_id !== selectedFeatureOption.value
      ) {
        return false
      }
      if (
        selectedEditionOption &&
        row.result.page_key !== selectedEditionOption.value
      ) {
        return false
      }
      if (!trimmedQuery) {
        return true
      }
      const haystack = [
        row.result.page_key,
        row.result.feature_id,
        row.featureName,
        row.featureDescription,
        row.featureRevision,
        row.editionDetails,
        row.value,
      ]
        .map((value) => normalizeSearchValue(value))
        .filter(Boolean)
        .join(' ')
      return haystack.includes(trimmedQuery)
    })
  }, [rows, searchQuery, selectedEditionOption, selectedFeatureOption])

  const editionOptions = useMemo(() => {
    const seen = new Set<string>()
    const options: EditionOption[] = []

    for (const row of rows) {
      const pageKey = normalizeText(row.result.page_key)
      if (!pageKey || !row.editionDetails || seen.has(pageKey)) {
        continue
      }
      seen.add(pageKey)
      options.push({
        value: pageKey,
        label: editionDetailsByKey.get(pageKey) || pageKey,
      })
    }

    return options.sort((left, right) =>
      left.label.localeCompare(right.label, undefined, {
        sensitivity: 'base',
      }),
    )
  }, [editionDetailsByKey, rows])

  const sortedRows = useMemo(() => {
    const getSortValue = (row: FeatureResultRow, key: SortKey) => {
      switch (key) {
        case 'pageKey':
          return normalizeSearchValue(row.result.page_key)
        case 'editionDetails':
          return normalizeSearchValue(row.editionDetails)
        case 'featureName':
          return normalizeSearchValue(row.featureName || row.result.name)
        case 'featureDescription':
          return normalizeSearchValue(
            row.featureDescription || row.result.description,
          )
      }
    }

    const data = [...filteredRows]
    data.sort((left, right) => {
      const leftValue = getSortValue(left, sortKey)
      const rightValue = getSortValue(right, sortKey)
      if (leftValue < rightValue) return sortDirection === 'asc' ? -1 : 1
      if (leftValue > rightValue) return sortDirection === 'asc' ? 1 : -1
      return 0
    })
    return data
  }, [filteredRows, sortDirection, sortKey])

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDirection((current) => (current === 'asc' ? 'desc' : 'asc'))
      return
    }
    setSortKey(key)
    setSortDirection('asc')
  }

  const renderSortHeader = (label: string, key: SortKey) => {
    const isActive = sortKey === key
    const arrow = isActive ? (sortDirection === 'asc' ? '▲' : '▼') : null

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

  const error =
    featuresQuery.error instanceof Error
      ? featuresQuery.error.message
      : featureResultsQuery.error instanceof Error
        ? featureResultsQuery.error.message
        : editionsQuery.error instanceof Error
          ? editionsQuery.error.message
          : featuresQuery.error ||
              featureResultsQuery.error ||
              editionsQuery.error
            ? 'Failed to load feature results.'
            : null

  const isLoading =
    featuresQuery.isLoading ||
    featureResultsQuery.isLoading ||
    editionsQuery.isLoading

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col bg-white m-3 mb-0 w-[calc(100%-1.5rem)] mx-auto">
      <div className="px-3 py-2 border-b border-gray-200 bg-gray-50 flex items-center gap-3">
        <div className="text-sm font-semibold">Feature Results</div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-4 space-y-4">
        <div className="flex items-center gap-4">
          <SearchInput
            placeholder="Search feature results"
            className="w-full max-w-[420px]"
            value={searchQuery}
            onChange={setSearchQuery}
          />
          <Select
            value={selectedFeatureOption}
            onChange={(option) => setSelectedFeatureOption(option)}
            options={featureOptions}
            placeholder="Filter by feature..."
            formatOptionLabel={(option) => (
              <div className="flex items-center gap-2">
                <span
                  className="inline-block h-2.5 w-2.5 shrink-0 rounded-full border border-gray-300"
                  style={{ backgroundColor: option.color || '#d1d5db' }}
                />
                <span>{option.label}</span>
              </div>
            )}
            styles={selectStyles<FeatureOption>({ controlWidth: 260 })}
            menuPortalTarget={document.body}
            menuPosition="fixed"
            isClearable
          />
          {editionOptions.length > 0 && (
            <Select
              value={selectedEditionOption}
              onChange={(option) => setSelectedEditionOption(option)}
              options={editionOptions}
              placeholder="Filter by edition..."
              styles={selectStyles<EditionOption>({ controlWidth: 320 })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          )}
          <div className="text-xs text-gray-500 shrink-0">
            {searchQuery.trim() ||
            selectedFeatureOption ||
            selectedEditionOption
              ? `Listing ${formatCount(filteredRows.length)} of ${formatCount(rows.length)} results`
              : `Listing ${formatCount(rows.length)} results`}
          </div>
        </div>

        <ErrorMessage message={error} />

        {isLoading ? (
          <div className="text-sm text-gray-500">
            Loading feature results...
          </div>
        ) : rows.length === 0 ? (
          <div className="text-sm text-gray-500">
            No feature results found yet.
          </div>
        ) : filteredRows.length === 0 ? (
          <div className="text-sm text-gray-500">
            No feature results match the current filters.
          </div>
        ) : (
          <div className="overflow-auto rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full text-sm table-auto">
              <thead className="bg-gray-50 text-xs text-gray-500">
                <tr>
                  <th className="w-28 px-4 py-3 text-left whitespace-nowrap">
                    {renderSortHeader('Page/Key', 'pageKey')}
                  </th>
                  <th className="px-4 py-3 text-left whitespace-nowrap">
                    {renderSortHeader('Edition Details', 'editionDetails')}
                  </th>
                  <th className="px-4 py-3 text-left whitespace-nowrap">
                    {renderSortHeader('Feature Name', 'featureName')}
                  </th>
                  <th className="px-4 py-3 text-left whitespace-nowrap">
                    {renderSortHeader(
                      'Feature Description',
                      'featureDescription',
                    )}
                  </th>
                  <th className="w-32 px-4 py-3 text-left whitespace-nowrap">
                    Feature Revision
                  </th>
                  <th className="px-4 py-3 text-left whitespace-nowrap">
                    Value
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {sortedRows.map((row, index) => (
                  <tr
                    key={`${row.result.id || row.result.feature_id || 'feature-result'}-${row.result.page_key || ''}-${index}`}
                    className="hover:bg-gray-50 align-top"
                  >
                    <td className="w-28 px-4 py-3 text-gray-700 font-mono">
                      <div
                        className="max-w-28 truncate whitespace-nowrap"
                        title={row.result.page_key || '-'}
                      >
                        {row.result.page_key || '-'}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-700 min-w-64">
                      {row.editionDetails || '-'}
                    </td>
                    <td className="px-4 py-3 text-gray-700 whitespace-nowrap">
                      <div className="flex items-center gap-2">
                        <span
                          className="inline-block h-2.5 w-2.5 shrink-0 rounded-full border border-gray-300"
                          style={{
                            backgroundColor: row.featureColor || '#d1d5db',
                          }}
                        />
                        <span>{row.featureName || row.result.name || '-'}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-700 min-w-64">
                      {row.featureDescription || row.result.description || '-'}
                    </td>
                    <td className="w-32 px-4 py-3 text-gray-700 font-mono">
                      <div
                        className="max-w-32 truncate whitespace-nowrap"
                        title={row.featureRevision || '-'}
                      >
                        {row.featureRevision || '-'}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-gray-700 min-w-64">
                      {row.value || '-'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </section>
  )
}
