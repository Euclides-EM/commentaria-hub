import { useState } from 'react'

export type AnnotationFilter = 'transcribed' | 'ground_truth'

interface AnnotationFilterDropdownProps {
  filters: AnnotationFilter[]
  onToggleFilter: (filter: AnnotationFilter) => void
}

export function AnnotationFilterDropdown({
  filters,
  onToggleFilter,
}: AnnotationFilterDropdownProps) {
  const [isOpen, setIsOpen] = useState(false)

  const getLabel = () => {
    if (filters.length === 0) return 'All'
    if (filters.length === 1) {
      return filters[0] === 'transcribed' ? 'Transcribed' : 'Ground Truth'
    }
    return `${filters.length} filters`
  }

  return (
    <div className="relative" style={{ minWidth: '120px' }}>
      <button
        className="flex items-center justify-between w-full px-2 py-1 text-sm bg-white border border-gray-400 rounded-md hover:border-gray-500 focus:border-blue-500 focus:outline-none focus:ring-3 focus:ring-blue-100 transition-colors"
        style={{ height: '32px' }}
        onClick={() => setIsOpen(!isOpen)}
      >
        <span className="text-gray-700">{getLabel()}</span>
        <svg
          className={`w-4 h-4 text-gray-600 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute top-full left-0 mt-1 w-full bg-white border border-gray-300 rounded-md shadow-lg z-20">
            <label className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer text-sm">
              <input
                type="checkbox"
                checked={filters.includes('transcribed')}
                onChange={() => onToggleFilter('transcribed')}
                className="text-blue-600"
              />
              <span className="text-gray-700">Transcribed</span>
            </label>
            <label className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer text-sm">
              <input
                type="checkbox"
                checked={filters.includes('ground_truth')}
                onChange={() => onToggleFilter('ground_truth')}
                className="text-blue-600"
              />
              <span className="text-gray-700">Ground Truth</span>
            </label>
          </div>
        </>
      )}
    </div>
  )
}
