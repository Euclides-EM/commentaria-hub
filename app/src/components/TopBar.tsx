import { DatasetSelector } from './DatasetSelector'
import { AnnotationSelector } from './AnnotationSelector'

interface TopBarProps {
  selectedDatasetId: string
  onDatasetChange: (id: string) => void
  selectedAnnotationId: string
  onAnnotationChange: (id: string) => void
  onlyTranscribed: boolean
  onOnlyTranscribedChange: (checked: boolean) => void
  pageNum: number
  onPageNumChange: (page: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  onLoad: () => void
  reqSummary: string
}

export function TopBar({
  selectedDatasetId,
  onDatasetChange,
  selectedAnnotationId,
  onAnnotationChange,
  onlyTranscribed,
  onOnlyTranscribedChange,
  pageNum,
  onPageNumChange,
  onPrevPage,
  onNextPage,
  onLoad,
  reqSummary,
}: TopBarProps) {
  return (
    <div className="p-3 border-b border-gray-200 bg-white">
      <div className="grid grid-cols-[1fr_1fr_0.7fr_0.5fr_auto] gap-2 items-end">
        <DatasetSelector
          selectedDatasetId={selectedDatasetId}
          onDatasetChange={onDatasetChange}
        />

        <AnnotationSelector
          datasetId={selectedDatasetId}
          onlyTranscribed={onlyTranscribed}
          selectedAnnotationId={selectedAnnotationId}
          onAnnotationChange={onAnnotationChange}
        />

        <div>
          <label className="block text-xs opacity-80 mb-1 ml-0.5">&nbsp;</label>
          <label className="flex items-center gap-1.5 text-xs opacity-90 select-none">
            <input
              type="checkbox"
              checked={onlyTranscribed}
              onChange={(e) => onOnlyTranscribedChange(e.target.checked)}
            />
            <span>Only transcribed annotations</span>
          </label>
        </div>

        <div>
          <label
            htmlFor="pageNum"
            className="block text-xs opacity-80 mb-1 ml-0.5"
          >
            Page
          </label>
          <input
            id="pageNum"
            type="number"
            min="0"
            className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
            value={pageNum}
            onChange={(e) => onPageNumChange(parseInt(e.target.value) || 0)}
          />
        </div>

        <div className="flex gap-2 items-end">
          <button
            className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
            onClick={onPrevPage}
          >
            ←
          </button>
          <button
            className="px-2.5 py-1.5 bg-black text-white border border-black rounded-lg hover:bg-gray-800 font-semibold text-xs"
            onClick={onLoad}
          >
            Load
          </button>
          <button
            className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
            onClick={onNextPage}
          >
            →
          </button>
        </div>
      </div>

      <div className="mt-2">
        <div className="text-xs opacity-75 font-mono">{reqSummary}</div>
      </div>
    </div>
  )
}
