import { type JSX, useEffect, useMemo, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import {
  useAnnotationCategories,
  useAnnotationSearch,
} from '../queries/annotations'
import { useAppState } from '../context/AppStateContext.tsx'
import { MultiSelectDropdown } from './MultiSelectDropdown'
import type { model_AnnotationPart } from '../api'

const renderHighlightedContent = (content: string, pattern: string) => {
  if (!pattern) {
    return content
  }
  try {
    const regex = new RegExp(pattern, 'gi')
    const matches = Array.from(content.matchAll(regex))
    if (matches.length === 0) return content
    const nodes: Array<string | JSX.Element> = []
    let lastIndex = 0
    matches.forEach((match, index) => {
      const start = match.index ?? 0
      const end = start + match[0].length
      if (start > lastIndex) nodes.push(content.slice(lastIndex, start))
      nodes.push(
        <mark
          key={`${start}-${end}-${index}`}
          className="bg-yellow-200/70 text-gray-900 rounded-sm px-0.5"
        >
          {match[0]}
        </mark>,
      )
      lastIndex = end
    })
    if (lastIndex < content.length) nodes.push(content.slice(lastIndex))
    return nodes
  } catch {
    return content
  }
}

const getResultKey = (result: model_AnnotationPart, index: number) =>
  `${result.location?.page ?? 'p'}-${result.category ?? 'c'}-${index}`

export function AnnotationSearchMenu() {
  const { state, jumpToPage } = useAppState()
  const [searchTerm, setSearchTerm] = useLocalStorageState(
    'annotationSearchTerm',
    { defaultValue: '' },
  )
  const [debouncedSearchTerm, setDebouncedSearchTerm] = useState(searchTerm)
  const [selectedCategories, setSelectedCategories] = useLocalStorageState<
    string[]
  >('annotationSearchCategories', {
    defaultValue: [],
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
      selectedCategories.length === 0 ||
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

  const getDropdownLabel = () => {
    if (!categories || categories.length === 0) {
      return 'No categories'
    }
    if (
      selectedCategories.length === 0 ||
      selectedCategories.length === categories.length
    )
      return 'All categories'
    if (selectedCategories.length === 1) return selectedCategories[0]
    return `${selectedCategories.length} categories`
  }

  const results = searchResults?.results ?? []

  return (
    <div className="flex flex-col min-h-0 h-full">
      <div className="px-3 pb-3">
        <div className="flex gap-2 items-center">
          <input
            className="flex-1 min-w-0 border border-gray-300 rounded-lg px-3 py-2 font-mono text-xs"
            placeholder="Search..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
          <MultiSelectDropdown
            allItems={categories || []}
            selectedItems={selectedCategories}
            onToggleItem={(category) => {
              if (
                selectedCategories.length === 0 ||
                selectedCategories.includes(category)
              ) {
                setSelectedCategories(
                  (selectedCategories.length === 0
                    ? categories || []
                    : selectedCategories
                  ).filter((item) => item !== category),
                )
              } else {
                const next = [...selectedCategories, category]
                setSelectedCategories(
                  next.length === (categories || []).length ? [] : next,
                )
              }
            }}
            getButtonLabel={() => getDropdownLabel()}
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
            {results.map((result, index) => (
              <div
                key={getResultKey(result, index)}
                className="border border-gray-200 rounded-lg p-2 text-xs bg-white hover:bg-gray-50 transition-colors cursor-pointer"
                onClick={() =>
                  result.location?.page && jumpToPage(result.location.page)
                }
              >
                <div className="flex items-center justify-between text-[11px] text-gray-500 mb-1">
                  <span>{result.category || 'Uncategorized'}</span>
                  {result.location?.page && (
                    <span>p. {result.location.page}</span>
                  )}
                </div>
                <div className="text-gray-800 leading-snug">
                  {result.content
                    ? renderHighlightedContent(result.content, normalizedSearch)
                    : 'No content'}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
