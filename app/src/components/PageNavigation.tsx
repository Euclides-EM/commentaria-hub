import { useAppState } from '../context/AppStateContext.tsx'
import type { model_Annotation } from '../api'
import { selectStyles } from '../styles/selectStyles.ts'
import Select from 'react-select'
import { useEffect, useMemo } from 'react'

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
    <div className="flex gap-4 items-center">
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
                ? { value: state.currentPage, label: String(state.currentPage) }
                : null
            }
            onChange={(option) => onPageNumChange(option?.value || 1)}
            options={availablePages.map((p) => ({
              value: p,
              label: String(p),
            }))}
            placeholder="Select page..."
            styles={selectStyles}
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
  )
}
