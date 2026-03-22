import { useEffect, useMemo, useState } from 'react'
import Select from 'react-select'
import useLocalStorageState from 'use-local-storage-state'
import {
  useAnnotationCategories,
  useAnnotationSearch,
} from '../../../../queries/annotations.ts'
import { useDatasetFeaturesQuery } from '../../../../queries/datasets.ts'
import { useAppState } from '../../../../context/useAppState.ts'
import { MultiSelectDropdown } from '../../../core/MultiSelectDropdown.tsx'
import { SearchInput } from '../../../core/SearchInput.tsx'
import { LoadingSpinner } from '../../../core/LoadingSpinner.tsx'
import { ErrorMessage } from '../../../core/ErrorMessage'
import { selectStyles } from '../../../../styles/selectStyles.ts'
import type { annotation_SearchWithin, common_ALTOPart } from '@hub-api'
import { startCase } from 'lodash'

const annotationSearchWithinOptions: annotation_SearchWithin[] = [
  'categories',
  'transcription',
  'translation',
  'biblio_metadata',
]

const annotationSearchWithinLabels: Record<annotation_SearchWithin, string> = {
  categories: 'Categories',
  transcription: 'Transcription',
  translation: 'Translation',
  biblio_metadata: 'Bibliographic Metadata',
}

type SearchWithinOption = {
  value: annotation_SearchWithin
  label: string
}

const buildSnippet = (content: string, maxLength = 64) => {
  const startMatch = content.match(/<em[^>]*>/i)
  if (!startMatch) {
    if (content.length <= maxLength) return content
    return `${content.slice(0, maxLength)}...`
  }

  const startTag = startMatch[0]
  const startIndex = content.indexOf(startTag)
  const endIndex = content.toLowerCase().indexOf('</em>', startIndex)
  if (endIndex === -1) return content

  const beforeText = content.slice(0, startIndex)
  const matchHtml = content.slice(startIndex, endIndex + 5)
  const afterText = content.slice(endIndex + 5)
  const matchText = matchHtml.replace(/<[^>]*>/g, '')
  const remaining = maxLength - matchText.length

  if (remaining <= 0) {
    const truncated = matchText.slice(0, maxLength)
    const suffix = matchText.length > maxLength ? '...' : ''
    return `${startTag}${truncated}${suffix}</em>`
  }

  const beforeLen = Math.floor(remaining / 2)
  const afterLen = remaining - beforeLen
  const beforeTrim =
    beforeText.length > beforeLen
      ? beforeText.slice(beforeText.length - beforeLen)
      : beforeText
  const afterTrim =
    afterText.length > afterLen ? afterText.slice(0, afterLen) : afterText
  const prefix = beforeText.length > beforeLen ? '...' : ''
  const suffix = afterText.length > afterLen ? '...' : ''

  return `${prefix}${beforeTrim}${matchHtml}${afterTrim}${suffix}`
}

const getResultKey = (result: common_ALTOPart, index: number) => {
  return `${result.location?.page ?? 'p'}-${result.category ?? 'c'}-${index}`
}

const getFormattedCategory = (
  category: string | undefined,
  featureNameById: Map<string, string>,
  shouldDeriveFromPattern: boolean,
) => {
  if (!category) {
    return ''
  }
  if (!shouldDeriveFromPattern) return category

  const match = category.match(/^feature\.([^.]+)\.(surface|property\.(.+))$/)
  if (!match) return category

  const featureId = match[1]
  const suffix = match[2]
  const propertyName = match[3]
  const featureName = featureNameById.get(featureId) || featureId

  if (suffix === 'surface') {
    return featureName
  }

  if (propertyName) {
    return `${featureName} (${startCase(propertyName)})`
  }

  return featureName
}

const getResultLocationDisplay = (result: common_ALTOPart) => {
  if (result.location?.text_block_id) {
    return result.location.text_block_id
  }
  if (result.location?.page) {
    if (Number.isInteger(Number(result.location.page))) {
      return `p. ${result.location.page}`
    } else {
      return result.location.page
    }
  }
  return null
}

const getResultJumpTarget = (
  result: common_ALTOPart,
): number | string | null => {
  if (result.location?.page === '0' && result.location.text_block_id) {
    return result.location.text_block_id
  }
  if (result.location?.page) {
    const pageNumber = Number(result.location.page)
    return Number.isNaN(pageNumber) ? result.location.page : pageNumber
  }
  return null
}

export function AnnotationSearchMenu() {
  const { state, dataset, jumpToPage, setSearchResultHighlight } = useAppState()
  const [searchTerm, setSearchTerm] = useLocalStorageState(
    'annotationSearchTerm',
    { defaultValue: '', storageSync: false },
  )
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState(searchTerm)
  const [selectedCategories, setSelectedCategories] = useLocalStorageState<
    string[] | null
  >('annotationSearchCategories', {
    defaultValue: null,
    storageSync: false,
  })
  const [selectedSearchWithin, setSelectedSearchWithin] = useLocalStorageState<
    annotation_SearchWithin | annotation_SearchWithin[] | null
  >('annotationSearchWithin', {
    defaultValue: null,
    storageSync: false,
  })

  const { data: categories } = useAnnotationCategories(
    state.datasetId,
    state.annotationId,
  )

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      setDebouncedSearchTerm(searchTerm)
    }, 300)
    return () => {
      window.clearTimeout(timeoutId)
    }
  }, [searchTerm])

  const normalizedSearch = debouncedSearchTerm.trim()

  const activeCategories = useMemo(() => {
    if (!categories || categories.length === 0) {
      return []
    }
    if (
      selectedCategories == null ||
      selectedCategories.length === categories.length
    )
      return categories
    return selectedCategories.filter((category) =>
      categories.includes(category),
    )
  }, [categories, selectedCategories])

  const sortedCategories = useMemo(
    () => [...activeCategories].sort(),
    [activeCategories],
  )
  const hasCategories = (categories?.length ?? 0) > 0
  const availableSearchWithinOptions = useMemo(() => {
    return annotationSearchWithinOptions.filter((option) => {
      if (option === 'categories') {
        return hasCategories
      }
      if (option === 'transcription') {
        return !hasCategories
      }
      if (option === 'biblio_metadata') {
        return !dataset?.edition_id
      }
      return true
    })
  }, [dataset?.edition_id, hasCategories])

  const normalizedSelectedSearchWithin = useMemo(() => {
    const currentSelection = Array.isArray(selectedSearchWithin)
      ? selectedSearchWithin
      : selectedSearchWithin
        ? [selectedSearchWithin]
        : []

    return currentSelection.find((option) =>
      availableSearchWithinOptions.includes(option),
    )
  }, [availableSearchWithinOptions, selectedSearchWithin])

  useEffect(() => {
    const nextSelection =
      normalizedSelectedSearchWithin ?? availableSearchWithinOptions[0] ?? null

    if (selectedSearchWithin !== nextSelection) {
      setSelectedSearchWithin(nextSelection)
    }
  }, [
    availableSearchWithinOptions,
    normalizedSelectedSearchWithin,
    selectedSearchWithin,
    setSelectedSearchWithin,
  ])

  const activeSearchWithin = useMemo(() => {
    return normalizedSelectedSearchWithin
      ? [normalizedSelectedSearchWithin]
      : []
  }, [normalizedSelectedSearchWithin])
  const searchWithinSelectOptions = useMemo<SearchWithinOption[]>(
    () =>
      availableSearchWithinOptions.map((item) => ({
        value: item,
        label: annotationSearchWithinLabels[item],
      })),
    [availableSearchWithinOptions],
  )
  const selectedSearchWithinOption = useMemo(
    () =>
      searchWithinSelectOptions.find(
        (option) => option.value === normalizedSelectedSearchWithin,
      ) || null,
    [normalizedSelectedSearchWithin, searchWithinSelectOptions],
  )
  const searchWithinSelectStyles = useMemo(() => {
    const baseStyles = selectStyles<SearchWithinOption>({
      controlWidth: '100%',
    })

    return {
      ...baseStyles,
      control: (
        base: Parameters<NonNullable<typeof baseStyles.control>>[0],
        state: Parameters<NonNullable<typeof baseStyles.control>>[1],
      ) => ({
        ...baseStyles.control?.(base, state),
        minHeight: 28,
        height: 28,
        fontSize: '12px',
        boxShadow: 'none',
      }),
      valueContainer: (
        base: Parameters<NonNullable<typeof baseStyles.valueContainer>>[0],
        state: Parameters<NonNullable<typeof baseStyles.valueContainer>>[1],
      ) => ({
        ...baseStyles.valueContainer?.(base, state),
        height: 26,
        padding: '1px 6px',
      }),
      singleValue: (
        base: Parameters<NonNullable<typeof baseStyles.singleValue>>[0],
        state: Parameters<NonNullable<typeof baseStyles.singleValue>>[1],
      ) => ({
        ...baseStyles.singleValue?.(base, state),
        fontSize: '12px',
      }),
      input: (
        base: Parameters<NonNullable<typeof baseStyles.input>>[0],
        state: Parameters<NonNullable<typeof baseStyles.input>>[1],
      ) => ({
        ...baseStyles.input?.(base, state),
        fontSize: '12px',
      }),
      indicatorsContainer: (
        base: Parameters<NonNullable<typeof baseStyles.indicatorsContainer>>[0],
        state: Parameters<
          NonNullable<typeof baseStyles.indicatorsContainer>
        >[1],
      ) => ({
        ...baseStyles.indicatorsContainer?.(base, state),
        height: 26,
      }),
      dropdownIndicator: (
        base: Parameters<NonNullable<typeof baseStyles.dropdownIndicator>>[0],
        state: Parameters<NonNullable<typeof baseStyles.dropdownIndicator>>[1],
      ) => ({
        ...baseStyles.dropdownIndicator?.(base, state),
        padding: '2px 4px',
      }),
      placeholder: (
        base: Parameters<NonNullable<typeof baseStyles.placeholder>>[0],
        state: Parameters<NonNullable<typeof baseStyles.placeholder>>[1],
      ) => ({
        ...baseStyles.placeholder?.(base, state),
        fontSize: '12px',
      }),
      option: (
        base: Parameters<NonNullable<typeof baseStyles.option>>[0],
        state: Parameters<NonNullable<typeof baseStyles.option>>[1],
      ) => ({
        ...baseStyles.option?.(base, state),
        fontSize: '12px',
        padding: '6px 10px',
      }),
    }
  }, [])

  const annotationSearchQuery = useAnnotationSearch(
    state.datasetId,
    state.annotationId,
    normalizedSearch,
    sortedCategories,
    activeSearchWithin,
  )

  const results = annotationSearchQuery.data?.results ?? []
  const isLoading = annotationSearchQuery.isLoading
  const error = annotationSearchQuery.error
  const featureDefinitionsQuery = useDatasetFeaturesQuery(
    state.datasetId,
    !hasCategories,
  )
  const featureNameById = useMemo(() => {
    const lookup = new Map<string, string>()
    for (const feature of featureDefinitionsQuery.data ?? []) {
      if (!feature.id) continue
      lookup.set(feature.id, feature.name?.trim() || feature.id)
    }
    return lookup
  }, [featureDefinitionsQuery.data])

  const searchWithinSelect = (
    <div
      className="w-full min-w-0 max-w-full"
      style={{ width: 'min(100%, 220px)', maxWidth: '100%' }}
    >
      <Select<SearchWithinOption, false>
        className="w-full min-w-0 max-w-full"
        value={selectedSearchWithinOption}
        onChange={(option) => setSelectedSearchWithin(option?.value || null)}
        options={searchWithinSelectOptions}
        placeholder="Select field..."
        styles={searchWithinSelectStyles}
        menuPortalTarget={document.body}
        menuPosition="fixed"
        isClearable={false}
      />
    </div>
  )

  return (
    <div className="flex flex-col min-h-0 h-full mr-1">
      <div className="px-3 pb-3">
        {hasCategories ? (
          <div className="flex flex-col gap-2">
            <SearchInput
              className="min-w-0"
              placeholder="Search..."
              value={searchTerm}
              onChange={(t) => {
                setSearchTerm(t)
                setSearchResultHighlight(null)
              }}
            />
            <div className="flex gap-2 items-center">
              <MultiSelectDropdown
                allItems={categories || []}
                selectedItems={selectedCategories}
                setSelectedItems={setSelectedCategories}
                itemsLabel="categories"
                getItemLabel={(category) => category}
              />
              {searchWithinSelect}
            </div>
          </div>
        ) : (
          <div className="flex flex-wrap gap-2 items-center">
            <SearchInput
              className="flex-1 min-w-0 w-full"
              placeholder="Search..."
              value={searchTerm}
              onChange={(t) => {
                setSearchTerm(t)
                setSearchResultHighlight(null)
              }}
            />
            {searchWithinSelect}
          </div>
        )}
      </div>

      <div className="overflow-auto px-3 pb-3 flex-1 min-h-0">
        {!normalizedSearch && (
          <div className="text-gray-500 text-xs italic text-center py-6">
            Type to search
          </div>
        )}
        {normalizedSearch && isLoading && (
          <LoadingSpinner size="sm" message="Searching..." />
        )}
        {normalizedSearch && error && (
          <ErrorMessage error={error} variant="centered" />
        )}
        {normalizedSearch && !isLoading && !error && results.length === 0 && (
          <div className="text-gray-500 text-xs italic text-center py-6">
            No results found
          </div>
        )}
        {normalizedSearch && !isLoading && !error && results.length > 0 && (
          <div className="space-y-2">
            <div className="text-xs text-gray-500">
              Listing {results.length} results
            </div>
            {results.map((result, index) => {
              const locationLabel = getResultLocationDisplay(result)
              return (
                <div
                  key={getResultKey(result, index)}
                  className="border border-gray-200 rounded-lg p-2 text-xs bg-white hover:bg-gray-50 transition-colors cursor-pointer"
                  onClick={() => {
                    const jumpTarget = getResultJumpTarget(result)
                    if (jumpTarget != null) {
                      jumpToPage(jumpTarget)
                    }
                    setSearchResultHighlight(result.content || null)
                  }}
                >
                  <div className="flex items-center justify-between text-[11px] text-gray-500 mb-1">
                    <span>
                      {getFormattedCategory(
                        result.category,
                        featureNameById,
                        !hasCategories,
                      )}
                    </span>
                    {locationLabel && <span>{locationLabel}</span>}
                  </div>
                  <div className="text-gray-800 leading-snug">
                    {result.content ? (
                      <span
                        dangerouslySetInnerHTML={{
                          __html: buildSnippet(result.content),
                        }}
                      />
                    ) : (
                      'No content'
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
