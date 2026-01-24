import { useAppState } from '../context/AppStateContext.tsx'
import type { model_Annotation } from '../api'
import { selectStyles } from '../styles/selectStyles.ts'
import Select from 'react-select'
import { useEffect, useMemo } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { IndexMenu } from './IndexMenu.tsx'

const expandRange = (range: string): number[] => {
  const parts = range.trim().split('-')

  if (parts.length !== 2) {
    const num = parseInt(range.trim())
    return isNaN(num) ? [] : [num]
  }

  const min = parseInt(parts[0].trim())
  const max = parseInt(parts[1].trim())

  if (isNaN(min) || isNaN(max)) return []

  return Array.from({ length: Math.max(0, max - min + 1) }, (_, i) => min + i)
}

const parseAvailablePages = (annotation: model_Annotation): number[] => {
  if (!annotation.pages) return []

  return annotation.pages.split(',').flatMap((p) => expandRange(p))
}

export function PageNavigation() {
  const { annotation, state, setState, jumpToPage } = useAppState()
  const [isIndexCollapsed, setIsIndexCollapsed] = useLocalStorageState(
    'indexCollapsed',
    {
      defaultValue: false,
    },
  )
  const [isSearchCollapsed, setIsSearchCollapsed] = useLocalStorageState(
    'searchCollapsed',
    {
      defaultValue: true,
    },
  )

  const onPageNumChange = (page: number) => setState({ currentPage: page })
  const availablePages = useMemo(
    () => (annotation ? parseAvailablePages(annotation) : []),
    [annotation],
  )

  const currentIndex = availablePages.indexOf(state.currentPage)
  const isFirstPage = currentIndex === 0
  const isLastPage = currentIndex === availablePages.length - 1

  const onPrevPage = () => {
    if (currentIndex > 0) {
      jumpToPage(availablePages[currentIndex - 1])
    }
  }
  const onNextPage = () => {
    if (currentIndex < availablePages.length - 1) {
      jumpToPage(availablePages[currentIndex + 1])
    }
  }

  useEffect(() => {
    if (!availablePages.includes(state.currentPage)) {
      setState({ currentPage: availablePages[0] })
    }
  }, [availablePages, setState, state.currentPage])

  if (!annotation) {
    return
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <div className="flex w-full px-2 py-4 gap-4 items-center justify-center">
        <div className="flex gap-2">
          {!isFirstPage && (
            <button
              title="Previous page"
              className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
              onClick={onPrevPage}
            >
              ←
            </button>
          )}

          <div className="flex items-center gap-2">
            <label htmlFor="pageNum" className="text-xs opacity-80">
              Page
            </label>
            <Select
              value={
                availablePages.find((p) => p === state.currentPage)
                  ? {
                      value: state.currentPage,
                      label: String(state.currentPage),
                    }
                  : null
              }
              onChange={(option: { value: number; label: string } | null) =>
                onPageNumChange(option?.value || 1)
              }
              options={availablePages.map((p) => ({
                value: p,
                label: String(p),
              }))}
              placeholder="Select page..."
              styles={selectStyles<{ value: number; label: string }>()}
              isClearable
            />
          </div>

          {!isLastPage && (
            <button
              title="Next page"
              className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
              onClick={onNextPage}
            >
              →
            </button>
          )}
        </div>
      </div>
      <div className="flex flex-col flex-1 min-h-0">
        <div className="flex flex-col flex-[0_0_50%] min-h-0 border-t border-gray-300 overflow-hidden">
          <button
            title={isIndexCollapsed ? 'Expand index' : 'Collapse index'}
            aria-label={isIndexCollapsed ? 'Expand index' : 'Collapse index'}
            className="w-full flex items-center gap-2 px-3 py-4 text-left text-gray-500 hover:text-gray-700 transition-colors"
            onClick={() => setIsIndexCollapsed((prev) => !prev)}
          >
            <span className="text-sm">{isIndexCollapsed ? '▶' : '▼'}</span>
            <span className="font-semibold text-sm">Index</span>
          </button>
          <div className="flex-1 min-h-0 overflow-hidden">
            {!isIndexCollapsed && <IndexMenu />}
          </div>
        </div>
        <div className="flex flex-col flex-[0_0_50%] min-h-0 border-t border-gray-300">
          <button
            title={isSearchCollapsed ? 'Expand search' : 'Collapse search'}
            aria-label={isSearchCollapsed ? 'Expand search' : 'Collapse search'}
            className="w-full flex items-center gap-2 px-3 py-4 text-left text-gray-500 hover:text-gray-700 transition-colors"
            onClick={() => setIsSearchCollapsed((prev) => !prev)}
          >
            <span className="text-sm">{isSearchCollapsed ? '▶' : '▼'}</span>
            <span className="font-semibold text-sm">Search</span>
          </button>
          <div className="flex-1 min-h-0 overflow-hidden">
            {!isSearchCollapsed && <div>TODO</div>}
          </div>
        </div>
      </div>
    </div>
  )
}
