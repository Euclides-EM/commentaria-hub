import { useEffect, useMemo, useState } from 'react'
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
import type { common_ALTOPart } from '@hub-api'
import { startCase } from 'lodash'

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
  if (!category) return 'Uncategorized'
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
  if (result.location?.page === 0 && result.location.text_block_id) {
    return result.location.text_block_id
  }
  if (result.location?.page) {
    return `p. ${result.location.page}`
  }
  return null
}

const getResultJumpTarget = (
  result: common_ALTOPart,
): number | string | null => {
  if (result.location?.page === 0 && result.location.text_block_id) {
    return result.location.text_block_id
  }
  if (result.location?.page) {
    return result.location.page
  }
  return null
}

export function AnnotationSearchMenu() {
  const { state, jumpToPage, setSearchResultHighlight } = useAppState()
  const [searchTerm, setSearchTerm] = useLocalStorageState(
    'annotationSearchTerm',
    { defaultValue: '' },
  )
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState(searchTerm)
  const [selectedCategories, setSelectedCategories] = useLocalStorageState<
    string[] | null
  >('annotationSearchCategories', {
    defaultValue: null,
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

  const annotationSearchQuery = useAnnotationSearch(
    state.datasetId,
    state.annotationId,
    normalizedSearch,
    sortedCategories,
  )

  const results = annotationSearchQuery.data?.results ?? []
  const isLoading = annotationSearchQuery.isLoading
  const error = annotationSearchQuery.error
  const hasCategories = (categories?.length ?? 0) > 0
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

  return (
    <div className="flex flex-col min-h-0 h-full mr-1">
      <div className="px-3 pb-3">
        <div className="flex gap-2 items-center">
          <SearchInput
            className="flex-1 min-w-0"
            placeholder="Search..."
            value={searchTerm}
            onChange={(t) => {
              setSearchTerm(t)
              setSearchResultHighlight(null)
            }}
          />
          {hasCategories && (
            <MultiSelectDropdown
              allItems={categories || []}
              selectedItems={selectedCategories}
              setSelectedItems={setSelectedCategories}
              itemsLabel="categories"
              getItemLabel={(category) => category}
            />
          )}
        </div>
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
