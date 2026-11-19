import { useEffect, useMemo, useRef, useState } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import Select from 'react-select'
import { parseAsString, useQueryState } from 'nuqs'
import { type feature_Result } from '@hub-api'
import { useAppState } from '../../context/useAppState'
import { useAuthStore } from '../../store/authStore'
import {
  useDatasetsQuery,
  useDatasetFeaturesQuery,
} from '../../queries/datasets'
import { useAnnotationsQuery } from '../../queries/annotations'
import { useAllEditionsQuery } from '../../queries/editions'
import { useGlobalFeaturesQuery } from '../../queries/features'
import { useFeatureResultsQuery } from '../../queries/featureResults'
import { SearchInput } from '../core/SearchInput'
import { ErrorMessage } from '../core/ErrorMessage'
import { Button } from '../core/Button'
import {
  EditionFilterSelect,
  type EditionFilterOption,
} from '../featureResults/EditionFilterSelect'
import { selectStyles } from '../../styles/selectStyles'
import { formatEditionLabel } from '../../utils/editions'

type ScopeOption = 'editions' | 'dataset'

type FeatureOption = {
  value: string
  label: string
  color: string
}

type EntityOption = {
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

const TABLE_MIN_WIDTH = 1120
const ROW_ESTIMATE = 60

const COLUMN_CLASS_NAMES = {
  pageKey: 'w-28 shrink-0',
  editionDetails: 'min-w-64 flex-[1.15]',
  featureName: 'min-w-56 flex-[0.95]',
  featureDescription: 'min-w-72 flex-[1.25]',
  featureRevision: 'w-32 shrink-0',
  value: 'min-w-72 flex-[1.05]',
} as const

const normalizeText = (value: string | null | undefined) => value?.trim() || ''

const normalizeSearchValue = (value: string | null | undefined) =>
  normalizeText(value).toLowerCase()

const formatValue = (result: feature_Result) =>
  (result.values ?? [])
    .map((item) => normalizeText(item.surface))
    .filter(Boolean)
    .join(', ')

const formatCount = (value: number) => value.toLocaleString()

const escapeCsvValue = (value: string | null | undefined) => {
  const normalized = normalizeText(value)
  return `"${normalized.replace(/"/g, '""')}"`
}

export function FeatureResultsBrowser() {
  const { getUrlForState } = useAppState()
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const [selectedScope, setSelectedScope] = useState<ScopeOption>('editions')
  const [selectedDatasetOption, setSelectedDatasetOption] =
    useState<EntityOption | null>(null)
  const [selectedAnnotationOption, setSelectedAnnotationOption] =
    useState<EntityOption | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedFeatureOption, setSelectedFeatureOption] =
    useState<FeatureOption | null>(null)
  const [featureId, setFeatureId] = useQueryState(
    'featureId',
    parseAsString.withDefault('').withOptions({ history: 'replace' }),
  )
  const [editionKey, setEditionKey] = useQueryState(
    'editionKey',
    parseAsString.withDefault('').withOptions({ history: 'replace' }),
  )
  const [selectedEditionOption, setSelectedEditionOption] =
    useState<EditionFilterOption | null>(null)
  const [sortKey, setSortKey] = useState<SortKey>('editionDetails')
  const [sortDirection, setSortDirection] = useState<SortDirection>('asc')
  const tableContainerRef = useRef<HTMLDivElement | null>(null)

  const datasetId =
    selectedScope === 'dataset' ? selectedDatasetOption?.value || '' : ''
  const annotationId =
    selectedScope === 'dataset' ? selectedAnnotationOption?.value || '' : ''

  const datasetsQuery = useDatasetsQuery()
  const annotationsQuery = useAnnotationsQuery(datasetId)
  const featuresQuery = useDatasetFeaturesQuery(
    datasetId,
    selectedScope === 'dataset' ? !!datasetId : false,
  )
  const globalFeatureDefinitionsQuery = useGlobalFeaturesQuery(
    ['revisions'],
    selectedScope === 'editions',
  )

  const featureResultsQuery = useFeatureResultsQuery({
    scope: selectedScope,
    datasetId,
    annotationId,
    enabled: true,
  })
  const editionsQuery = useAllEditionsQuery()

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

  const featureDefinitions = useMemo(
    () =>
      selectedScope === 'dataset'
        ? (featuresQuery.data ?? [])
        : (globalFeatureDefinitionsQuery.data ?? []),
    [selectedScope, featuresQuery.data, globalFeatureDefinitionsQuery.data],
  )

  const featureOptions = useMemo(
    () =>
      [...featureDefinitions]
        .filter((feature) => !!feature.id)
        .sort((left, right) =>
          (left.name || left.id!).localeCompare(
            right.name || right.id!,
            undefined,
            {
              sensitivity: 'base',
            },
          ),
        )
        .map((feature) => ({
          value: feature.id!,
          label: feature.name || feature.id!,
          color: feature.color || '',
        })),
    [featureDefinitions],
  )

  useEffect(() => {
    if (!featureId || featureOptions.length === 0) return
    const option = featureOptions.find((item) => item.value === featureId)
    if (option) {
      setSelectedFeatureOption(option)
    }
  }, [featureId, featureOptions])

  const featureById = useMemo(() => {
    const map = new Map<
      string,
      {
        name: string
        color: string
        description: string
      }
    >()

    for (const feature of featureDefinitions) {
      if (!feature.id) continue
      map.set(feature.id, {
        name: feature.name || feature.id,
        color: feature.color || '',
        description: feature.description || '',
      })
    }

    return map
  }, [featureDefinitions])

  const latestRevisionByFeatureId = useMemo(() => {
    const map = new Map<string, string>()

    for (const feature of featureDefinitions) {
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
  }, [featureDefinitions])

  const editionDetailsByKey = useMemo(() => {
    const map = new Map<string, string>()
    for (const item of editionsQuery.data ?? []) {
      map.set(item.key!, formatEditionLabel(item))
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
          editionDetails: editionDetailsByKey.get(result.key || '') || '',
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
        row.result.key !== selectedEditionOption.value
      ) {
        return false
      }
      if (!trimmedQuery) {
        return true
      }
      const haystack = [
        row.result.key,
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
    const options: EditionFilterOption[] = []

    for (const row of rows) {
      const pageKey = normalizeText(row.result.key)
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

  useEffect(() => {
    if (!editionKey || editionOptions.length === 0) return
    const option = editionOptions.find((item) => item.value === editionKey)
    if (option) {
      setSelectedEditionOption(option)
    }
  }, [editionKey, editionOptions])

  const sortedRows = useMemo(() => {
    const getSortValue = (row: FeatureResultRow, key: SortKey) => {
      switch (key) {
        case 'pageKey':
          return normalizeSearchValue(row.result.key)
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

  const handleExport = () => {
    const header = [
      'Page/Key',
      'Edition Details',
      'Feature Name',
      'Feature Description',
      'Feature Revision',
      'Value',
    ]
    const lines = [
      header.map((value) => escapeCsvValue(value)).join(','),
      ...sortedRows.map((row) =>
        [
          row.result.key || '',
          row.editionDetails,
          row.featureName || row.result.name || '',
          row.featureDescription || row.result.description || '',
          row.featureRevision,
          row.value,
        ]
          .map((value) => escapeCsvValue(value))
          .join(','),
      ),
    ]
    const blob = new Blob([`${lines.join('\n')}\n`], {
      type: 'text/csv;charset=utf-8;',
    })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    const suffix =
      selectedScope === 'dataset'
        ? normalizeText(annotationId) || 'feature-results'
        : 'editions-feature-results'

    link.href = url
    link.download = `feature-results-${suffix}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
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

  const getPageLink = (pageKey: string | null | undefined) => {
    const nextPageKey = normalizeText(pageKey)
    if (
      selectedScope !== 'dataset' ||
      !datasetId ||
      !annotationId ||
      !nextPageKey
    ) {
      return null
    }
    return getUrlForState({
      viewMode: null,
      datasetId,
      annotationId,
      annotationTab: 'text',
      currentPageOrKey: nextPageKey,
    })
  }

  // eslint-disable-next-line react-hooks/incompatible-library
  const rowVirtualizer = useVirtualizer({
    count: sortedRows.length,
    getScrollElement: () => tableContainerRef.current,
    estimateSize: () => ROW_ESTIMATE,
    overscan: 10,
  })

  const virtualRows = rowVirtualizer.getVirtualItems()

  const selectorError =
    datasetsQuery.error instanceof Error
      ? datasetsQuery.error.message
      : annotationsQuery.error instanceof Error
        ? annotationsQuery.error.message
        : null

  const error =
    selectorError ||
    (featuresQuery.error instanceof Error
      ? featuresQuery.error.message
      : globalFeatureDefinitionsQuery.error instanceof Error
        ? globalFeatureDefinitionsQuery.error.message
        : featureResultsQuery.error instanceof Error
          ? featureResultsQuery.error.message
          : editionsQuery.error instanceof Error
            ? editionsQuery.error.message
            : featuresQuery.error ||
                globalFeatureDefinitionsQuery.error ||
                featureResultsQuery.error ||
                editionsQuery.error
              ? 'Failed to load feature results.'
              : null)

  const isLoading =
    datasetsQuery.isLoading ||
    (selectedScope === 'dataset' && annotationsQuery.isLoading) ||
    (selectedScope === 'dataset'
      ? featuresQuery.isLoading
      : globalFeatureDefinitionsQuery.isLoading) ||
    featureResultsQuery.isLoading ||
    editionsQuery.isLoading

  const needsDatasetSelection =
    selectedScope === 'dataset' && !selectedDatasetOption?.value
  const needsAnnotationSelection =
    selectedScope === 'dataset' &&
    !!selectedDatasetOption?.value &&
    !selectedAnnotationOption?.value

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col bg-white">
      <div className="px-3 py-2 border-b border-gray-200 bg-gray-50 flex items-center justify-between gap-3">
        <div className="text-sm font-semibold">Feature Results</div>
        <div className="flex items-center gap-2 shrink-0">
          {isAuthenticated && (
            <Button
              type="button"
              className="px-2 py-1 text-xs"
              onClick={handleExport}
              disabled={sortedRows.length === 0}
            >
              Export CSV
            </Button>
          )}
        </div>
      </div>

      <div className="flex-1 min-h-0 flex flex-col overflow-hidden p-4 gap-4">
        <div className="shrink-0 flex items-center gap-4 flex-wrap">
          <select
            value={selectedScope}
            onChange={(event) => {
              const nextScope = event.target.value as ScopeOption
              setSelectedScope(nextScope)
              setSelectedDatasetOption(null)
              setSelectedAnnotationOption(null)
              setSelectedFeatureOption(null)
              void setFeatureId(null)
              setSelectedEditionOption(null)
              void setEditionKey(null)
            }}
            className="h-9 px-3 text-sm border border-gray-300 rounded-md bg-white"
          >
            <option value="editions">Editions</option>
            <option value="dataset">Dataset</option>
          </select>
          {selectedScope === 'dataset' && (
            <Select
              value={selectedDatasetOption}
              onChange={(option) => {
                setSelectedDatasetOption(option)
                setSelectedAnnotationOption(null)
                setSelectedFeatureOption(null)
                void setFeatureId(null)
                setSelectedEditionOption(null)
                void setEditionKey(null)
              }}
              options={datasetOptions}
              placeholder="Select dataset..."
              styles={selectStyles<EntityOption>({ controlWidth: 260 })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          )}
          {selectedScope === 'dataset' && selectedDatasetOption && (
            <Select
              value={selectedAnnotationOption}
              onChange={(option) => {
                setSelectedAnnotationOption(option)
                setSelectedFeatureOption(null)
                void setFeatureId(null)
                setSelectedEditionOption(null)
                void setEditionKey(null)
              }}
              options={annotationOptions}
              placeholder="Select annotation..."
              styles={selectStyles<EntityOption>({ controlWidth: 300 })}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          )}
          <SearchInput
            placeholder="Search feature results"
            className="w-full max-w-105"
            value={searchQuery}
            onChange={setSearchQuery}
          />
          <Select
            value={selectedFeatureOption}
            onChange={(option) => {
              setSelectedFeatureOption(option)
              void setFeatureId(option?.value || null)
            }}
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
            isDisabled={featureOptions.length === 0}
          />
          <EditionFilterSelect
            value={selectedEditionOption}
            onChange={(option) => {
              setSelectedEditionOption(option)
              void setEditionKey(option?.value || '')
            }}
            options={editionOptions}
          />
          <div className="text-xs text-gray-500 shrink-0">
            {searchQuery.trim() ||
            selectedFeatureOption ||
            selectedEditionOption
              ? `Listing ${formatCount(filteredRows.length)} of ${formatCount(rows.length)} results`
              : `Listing ${formatCount(rows.length)} results`}
          </div>
        </div>

        <ErrorMessage message={error} />

        {needsDatasetSelection ? (
          <div className="text-sm text-gray-500">
            Select a dataset to load feature results.
          </div>
        ) : needsAnnotationSelection ? (
          <div className="text-sm text-gray-500">
            Select an annotation to load dataset-scoped feature results.
          </div>
        ) : isLoading ? (
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
          <div
            ref={tableContainerRef}
            className="flex-1 min-h-0 overflow-y-auto overflow-x-clip rounded-lg border border-gray-200 bg-white"
          >
            <div className="h-full" style={{ minWidth: TABLE_MIN_WIDTH }}>
              <div className="sticky top-0 z-10 flex border-b border-gray-200 bg-gray-50 text-xs text-gray-500 shadow-[0_1px_0_0_rgba(229,231,235,1)]">
                <div
                  className={`${COLUMN_CLASS_NAMES.pageKey} px-4 py-3 text-left whitespace-nowrap`}
                >
                  {renderSortHeader('Page/Key', 'pageKey')}
                </div>
                <div
                  className={`${COLUMN_CLASS_NAMES.editionDetails} px-4 py-3 text-left whitespace-nowrap`}
                >
                  {renderSortHeader('Edition Details', 'editionDetails')}
                </div>
                <div
                  className={`${COLUMN_CLASS_NAMES.featureName} px-4 py-3 text-left whitespace-nowrap`}
                >
                  {renderSortHeader('Feature Name', 'featureName')}
                </div>
                <div
                  className={`${COLUMN_CLASS_NAMES.featureDescription} px-4 py-3 text-left whitespace-nowrap`}
                >
                  {renderSortHeader(
                    'Feature Description',
                    'featureDescription',
                  )}
                </div>
                <div
                  className={`${COLUMN_CLASS_NAMES.featureRevision} px-4 py-3 text-left whitespace-nowrap`}
                >
                  Feature Revision
                </div>
                <div
                  className={`${COLUMN_CLASS_NAMES.value} px-4 py-3 text-left whitespace-nowrap`}
                >
                  Value
                </div>
              </div>

              <div
                className="relative"
                style={{ height: rowVirtualizer.getTotalSize() }}
              >
                {virtualRows.map((virtualRow) => {
                  const row = sortedRows[virtualRow.index]
                  return (
                    <div
                      key={`${row.result.id || row.result.feature_id || 'feature-result'}-${row.result.key || ''}-${virtualRow.index}`}
                      ref={rowVirtualizer.measureElement}
                      data-index={virtualRow.index}
                      className="absolute left-0 top-0 flex w-full border-b border-gray-200 bg-white text-sm hover:bg-gray-50"
                      style={{
                        transform: `translateY(${virtualRow.start}px)`,
                      }}
                    >
                      <div
                        className={`${COLUMN_CLASS_NAMES.pageKey} flex items-start px-4 py-3 text-xs text-gray-700 font-mono`}
                      >
                        {getPageLink(row.result.key) ? (
                          <a
                            href={getPageLink(row.result.key) || undefined}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="max-w-28 truncate whitespace-nowrap text-left text-sky-700 hover:text-sky-900 hover:underline"
                            title={`Open ${row.result.key} in page view`}
                          >
                            {row.result.key}
                          </a>
                        ) : (
                          <div
                            className="max-w-28 truncate whitespace-nowrap"
                            title={row.result.key || '-'}
                          >
                            {row.result.key || '-'}
                          </div>
                        )}
                      </div>
                      <div
                        className={`${COLUMN_CLASS_NAMES.editionDetails} flex items-start px-4 py-3 text-gray-700 leading-5`}
                      >
                        {row.editionDetails || '-'}
                      </div>
                      <div
                        className={`${COLUMN_CLASS_NAMES.featureName} flex items-start px-4 py-3 text-gray-700`}
                      >
                        <div className="flex items-center gap-2 whitespace-nowrap leading-5">
                          <span
                            className="inline-block h-2.5 w-2.5 shrink-0 rounded-full border border-gray-300"
                            style={{
                              backgroundColor: row.featureColor || '#d1d5db',
                            }}
                          />
                          <span>
                            {row.featureName || row.result.name || '-'}
                          </span>
                        </div>
                      </div>
                      <div
                        className={`${COLUMN_CLASS_NAMES.featureDescription} flex items-start px-4 py-3 text-gray-700 leading-5`}
                      >
                        {row.featureDescription ||
                          row.result.description ||
                          '-'}
                      </div>
                      <div
                        className={`${COLUMN_CLASS_NAMES.featureRevision} flex items-start px-4 py-3 text-xs text-gray-700 font-mono`}
                      >
                        <div
                          className="max-w-32 truncate whitespace-nowrap"
                          title={row.featureRevision || '-'}
                        >
                          {row.featureRevision || '-'}
                        </div>
                      </div>
                      <div
                        className={`${COLUMN_CLASS_NAMES.value} flex items-start px-4 py-3 text-gray-700 leading-5`}
                      >
                        {row.value || '-'}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        )}
      </div>
    </section>
  )
}
