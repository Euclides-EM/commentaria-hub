import { useAppState } from '../../../../context/useAppState.ts'
import type { annotation_Annotation } from '@hub-api'
import { selectStyles } from '../../../../styles/selectStyles.ts'
import Select from 'react-select'
import { useEffect, useMemo, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { IndexMenu } from './IndexMenu.tsx'
import { AnnotationSearchMenu } from './AnnotationSearchMenu.tsx'
import { useDatasetImageKeysQuery } from '../../../../queries/datasets.ts'
import {
  formatEditionLabel,
  TITLE_PAGES_DATASET_ID,
} from '../../../../utils/editions.ts'
import { expandRange } from '../../../../utils/pages.ts'
import { useQuery } from '@tanstack/react-query'
import { listAllEditions } from '../../../../queries/editions.ts'

const parseAvailablePages = (annotation: annotation_Annotation): string[] => {
  if (!annotation.pages) {
    return []
  }

  return annotation.pages.split(',').flatMap((p) => expandRange(p))
}

const getDefaultPageOrKey = (availablePages: string[]): string => {
  if (!availablePages.length) return ''
  if (availablePages[0] !== '1') {
    return availablePages[0]
  }
  return availablePages[Math.floor(availablePages.length / 2)]
}

export function PageNavigation() {
  const { annotation, state, setState, jumpToPage } = useAppState()
  const isKeyNavigation =
    !!annotation &&
    (!annotation.pages || annotation.dataset_id === TITLE_PAGES_DATASET_ID)
  const showIndexPane = !!annotation?.segmented
  const showSearchPane =
    !!annotation && (!annotation.pages || !!annotation.ocred)
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    state.datasetId,
    isKeyNavigation,
    annotation?.pages ? annotation.pages.split(',') : null,
  )
  const editionsQuery = useQuery({
    queryKey: ['editions', 'all', 'items'],
    queryFn: async () => await listAllEditions(),
    enabled: isKeyNavigation,
    refetchOnWindowFocus: false,
  })
  const currentValue = String(state.currentPageOrKey)
  const [isIndexCollapsed, setIsIndexCollapsed] = useLocalStorageState(
    'indexCollapsed',
    {
      defaultValue: false,
      storageSync: false,
    },
  )
  const [isSearchCollapsed, setIsSearchCollapsed] = useLocalStorageState(
    'searchCollapsed',
    {
      defaultValue: true,
      storageSync: false,
    },
  )
  const [splitRatio, setSplitRatio] = useLocalStorageState(
    'indexSearchSplitRatio',
    {
      defaultValue: 0.5,
      storageSync: false,
    },
  )
  const [isResizing, setIsResizing] = useState(false)
  const splitRef = useRef<HTMLDivElement | null>(null)

  const onPageNumChange = (value: string) =>
    setState({ currentPageOrKey: value })
  const availableOptions = useMemo(() => {
    if (!annotation) {
      return []
    }
    if (annotation.pages && annotation.dataset_id !== TITLE_PAGES_DATASET_ID) {
      const pages = parseAvailablePages(annotation)
      return [...new Set(pages)]
        .sort((a, b) => a.localeCompare(b, undefined, { numeric: true }))
        .map((page) => ({
          value: page,
          label: page,
        }))
    }
    return imageKeys.map((image) => ({
      value: image.key,
      label: image.key,
    }))
  }, [annotation, imageKeys])

  const availablePages = useMemo(
    () => availableOptions.map((option) => option.value),
    [availableOptions],
  )

  const currentOption = useMemo(
    () =>
      availableOptions.find((option) => option.value === currentValue) ||
      availableOptions.find((option) => option.label === currentValue) ||
      null,
    [availableOptions, currentValue],
  )

  const currentOptionValue = currentOption?.value || currentValue
  const editionDetailsByKey = useMemo(() => {
    const map = new Map<string, string>()
    for (const item of editionsQuery.data ?? []) {
      if (!item.key) continue
      map.set(item.key, formatEditionLabel(item))
    }
    return map
  }, [editionsQuery.data])
  const currentEditionDetails =
    editionDetailsByKey.get(currentOptionValue) || ''
  const currentIndex = availablePages.indexOf(currentOptionValue)
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
    if (
      currentOption &&
      currentOption.value !== currentValue &&
      availablePages.includes(currentOption.value)
    ) {
      setState({ currentPageOrKey: currentOption.value })
      return
    }

    if (availablePages.length > 0 && !availablePages.includes(currentValue)) {
      setState({ currentPageOrKey: getDefaultPageOrKey(availablePages) })
    }
  }, [availablePages, currentOption, currentValue, setState])

  useEffect(() => {
    if (showSearchPane && !showIndexPane && isSearchCollapsed) {
      setIsSearchCollapsed(false)
    }
  }, [isSearchCollapsed, setIsSearchCollapsed, showIndexPane, showSearchPane])

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
          <button
            title="Previous page"
            className={`px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white font-semibold text-xs ${isFirstPage ? 'invisible pointer-events-none' : 'hover:bg-gray-50'}`}
            onClick={onPrevPage}
            disabled={isFirstPage}
            aria-hidden={isFirstPage}
            tabIndex={isFirstPage ? -1 : 0}
          >
            ←
          </button>

          <div className="flex items-center gap-2">
            <label htmlFor="pageNum" className="text-xs opacity-80">
              {isKeyNavigation ? 'Key' : 'Page'}
            </label>
            <Select
              value={
                currentOption
                  ? {
                      value: currentOption.value,
                      label: currentOption.label,
                    }
                  : null
              }
              onChange={(option: { value: string; label: string } | null) =>
                onPageNumChange(
                  option?.value ?? getDefaultPageOrKey(availablePages) ?? '1',
                )
              }
              options={availableOptions}
              placeholder={isKeyNavigation ? 'Select key...' : 'Select page...'}
              styles={selectStyles<{ value: string; label: string }>()}
              menuPortalTarget={document.body}
              menuPosition="fixed"
              isClearable
            />
          </div>

          <button
            title="Next page"
            className={`px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white font-semibold text-xs ${isLastPage ? 'invisible pointer-events-none' : 'hover:bg-gray-50'}`}
            onClick={onNextPage}
            disabled={isLastPage}
            aria-hidden={isLastPage}
            tabIndex={isLastPage ? -1 : 0}
          >
            →
          </button>
        </div>
      </div>
      {currentEditionDetails && (
        <div className="px-3 pb-4 -mt-2 text-xs text-gray-500 leading-5 text-center">
          {currentEditionDetails}
        </div>
      )}
      {showIndexPane && showSearchPane && (
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
              {!isIndexCollapsed && (
                <IndexMenu
                  disableHighlight={state.annotationTab === 'gallery'}
                />
              )}
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
      {showIndexPane && !showSearchPane && (
        <div className="flex flex-col flex-1 min-h-0 border-t border-gray-300">
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
            {!isIndexCollapsed && (
              <IndexMenu disableHighlight={state.annotationTab === 'gallery'} />
            )}
          </div>
        </div>
      )}
      {!showIndexPane && showSearchPane && (
        <div className="flex flex-col flex-1 min-h-0 border-t border-gray-300">
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
            {!isSearchCollapsed && <AnnotationSearchMenu />}
          </div>
        </div>
      )}
    </div>
  )
}
