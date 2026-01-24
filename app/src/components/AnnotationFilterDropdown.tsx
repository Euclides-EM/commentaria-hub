import { useState } from 'react'
import type { annotationrule_PipelineStage } from '../api'
import { getStageDisplayName } from '../utils/rules.ts'

interface AnnotationFilterDropdownProps {
  allStages: annotationrule_PipelineStage[]
  selectedStages: annotationrule_PipelineStage[]
  onToggleStage: (filter: annotationrule_PipelineStage) => void
}

export function AnnotationFilterDropdown({
  allStages,
  selectedStages,
  onToggleStage,
}: AnnotationFilterDropdownProps) {
  const [isOpen, setIsOpen] = useState(false)

  const getLabel = () => {
    if (
      selectedStages.length === 0 ||
      selectedStages.length === allStages.length
    )
      return 'All stages'
    if (selectedStages.length === 1) {
      return getStageDisplayName(selectedStages[0])
    }
    return `${selectedStages.length} stages`
  }

  return (
    <div className="relative" style={{ minWidth: '160px' }}>
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
          <div className="absolute top-full left-0 mt-1 w-max bg-white border border-gray-300 rounded-md shadow-lg z-20">
            {allStages.map((stage) => (
              <label
                key={stage}
                className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer text-sm"
              >
                <input
                  type="checkbox"
                  checked={
                    selectedStages.length === 0 ||
                    selectedStages.includes(stage)
                  }
                  onChange={() => onToggleStage(stage)}
                  className="text-blue-600"
                />
                <span className="text-gray-700">
                  {getStageDisplayName(stage)}
                </span>
              </label>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
