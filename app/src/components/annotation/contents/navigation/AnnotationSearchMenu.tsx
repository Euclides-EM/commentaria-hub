import { useEffect, useMemo, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import {
  useAnnotationCategories,
  useAnnotationSearch,
} from '../../../../queries/annotations.ts'
import { useAppState } from '../../../../context/useAppState.ts'
import { MultiSelectDropdown } from '../../../core/MultiSelectDropdown.tsx'
import type { model_AnnotationPart } from '../../../../api'
import { SearchInput } from '../../../core/SearchInput.tsx'

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

const getResultKey = (result: model_AnnotationPart, index: number) =>
  `${result.location?.page ?? 'p'}-${result.category ?? 'c'}-${index}`

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

  const {
    data: searchResults,
    isLoading,
    error,
  } = useAnnotationSearch(
    state.datasetId,
    state.annotationId,
    normalizedSearch,
    sortedCategories,
  )

  const results = searchResults?.results ?? []

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
          <MultiSelectDropdown
            allItems={categories || []}
            selectedItems={selectedCategories}
            setSelectedItems={setSelectedCategories}
            itemsLabel="categories"
            getItemLabel={(category) => category}
            disabled={!categories || categories.length === 0}
          />
        </div>
      </div>

      <div className="overflow-auto px-3 pb-3 flex-1 min-h-0">
        {!normalizedSearch && (
          <div className="text-gray-500 text-xs italic text-center py-6">
            Type to search
          </div>
        )}
        {normalizedSearch && isLoading && (
          <div className="text-gray-500 text-xs text-center py-6">
            Searching...
          </div>
        )}
        {normalizedSearch && error && (
          <div className="text-red-600 text-xs text-center py-6">
            {error.message}
          </div>
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
            {results.map((result, index) => (
              <div
                key={getResultKey(result, index)}
                className="border border-gray-200 rounded-lg p-2 text-xs bg-white hover:bg-gray-50 transition-colors cursor-pointer"
                onClick={() => {
                  if (result.location?.page) {
                    jumpToPage(result.location.page)
                  }
                  setSearchResultHighlight(
                    result.location?.text_block_id ?? null,
                  )
                }}
              >
                <div className="flex items-center justify-between text-[11px] text-gray-500 mb-1">
                  <span>{result.category || 'Uncategorized'}</span>
                  {result.location?.page && (
                    <span>p. {result.location.page}</span>
                  )}
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
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
