import { useAppState } from '../../context/useAppState'
import type { model_Annotation } from '../../api'
import { selectStyles } from '../../styles/selectStyles.ts'
import Select from 'react-select'
import { useEffect, useMemo, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { IndexMenu } from './IndexMenu.tsx'
import { AnnotationSearchMenu } from './AnnotationSearchMenu'

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
  const [splitRatio, setSplitRatio] = useLocalStorageState(
    'indexSearchSplitRatio',
    {
      defaultValue: 0.5,
    },
  )
  const [isResizing, setIsResizing] = useState(false)
  const splitRef = useRef<HTMLDivElement | null>(null)

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

  useEffect(() => {
    if (!isResizing) {
      return
    }

    const onPointerMove = (event: PointerEvent) => {
      const container = splitRef.current
      if (!container) {
        return
      }
      const rect = container.getBoundingClientRect()
      const raw = (event.clientY - rect.top) / rect.height
      const clamped = Math.min(0.8, Math.max(0.2, raw))
      setSplitRatio(clamped)
    }

    const onPointerUp = () => {
      setIsResizing(false)
    }

    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
    return () => {
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerup', onPointerUp)
    }
  }, [isResizing, setSplitRatio])

  if (!annotation) {
    return
  }

  return (
    <div className="flex flex-col flex-1 min-h-0 mr-1">
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
      {annotation.ocred && (
        <div className="flex flex-col flex-1 min-h-0" ref={splitRef}>
          <div
            className="flex flex-col min-h-0 border-t border-gray-300 overflow-hidden flex-none"
            style={{ flexBasis: `calc(${splitRatio * 100}% - 4px)` }}
          >
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
          <div
            role="separator"
            aria-label="Resize index and search"
            className="h-2 cursor-row-resize bg-gray-100 hover:bg-gray-200 transition-colors flex items-center justify-center"
            onPointerDown={(event) => {
              event.preventDefault()
              setIsResizing(true)
            }}
          >
            <div className="h-0.5 w-10 rounded-full bg-gray-300" />
          </div>
          <div
            className="flex flex-col min-h-0 flex-none"
            style={{ flexBasis: `calc(${(1 - splitRatio) * 100}% - 4px)` }}
          >
            <button
              title={isSearchCollapsed ? 'Expand search' : 'Collapse search'}
              aria-label={
                isSearchCollapsed ? 'Expand search' : 'Collapse search'
              }
              className="w-full flex items-center gap-2 px-3 py-4 text-left text-gray-500 hover:text-gray-700 transition-colors"
              onClick={() => setIsSearchCollapsed((prev) => !prev)}
            >
              <span className="text-sm">{isSearchCollapsed ? '▶' : '▼'}</span>
              <span className="font-semibold text-sm">Search</span>
            </button>
            <div className="flex-1 min-h-0 overflow-hidden">
              {!isSearchCollapsed && <AnnotationSearchMenu />}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
