import { useAppState } from '../context/AppStateContext.tsx'
import type { model_Annotation } from '../api'
import { selectStyles } from '../styles/selectStyles.ts'
import Select from 'react-select'

const expandRange = (range: string): number[] => {
  const parts = range.trim().split('-')
  if (parts.length !== 2) {
    return [parseInt(range)]
  }
  const min = parseInt(parts[0].trim())
  const max = parseInt(parts[1].trim())
  return [min, max] // TODO
}

const parseAvailablePages = (annotation: model_Annotation): number[] => {
  return (annotation.pages || '').trim().split(',').flatMap(expandRange)
}

export function PageNavigation() {
  const { annotation, state, setState, jumpToPage } = useAppState()

  const onPageNumChange = (page: number) => setState({ currentPage: page })
  const onPrevPage = () => jumpToPage(state.currentPage - 1)
  const onNextPage = () => jumpToPage(state.currentPage + 1)
  if (!annotation) {
    return
  }
  const availablePages = parseAvailablePages(annotation)

  if (!availablePages.includes(state.currentPage)) {
    setState({ currentPage: availablePages[0] })
  }

  return (
    <div className="p-3 border-b border-gray-200 bg-white">
      <div className="flex gap-4 items-center">
        <div className="flex gap-2">
          <button
            className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
            onClick={onPrevPage}
          >
            ←
          </button>

          <div className="flex items-center gap-2">
            <label htmlFor="pageNum" className="text-xs opacity-80">
              Page
            </label>
            <Select
              value={state.currentPage}
              onChange={(option) => onPageNumChange(option || 1)}
              options={availablePages.map((p) => ({
                value: p,
                label: String(p),
              }))}
              placeholder="Select page..."
              styles={selectStyles}
              isClearable
            />
          </div>

          <button
            className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
            onClick={onNextPage}
          >
            →
          </button>
        </div>
      </div>
    </div>
  )
}
