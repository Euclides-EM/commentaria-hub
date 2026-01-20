interface TopBarProps {
  pageNum: number
  onPageNumChange: (page: number) => void
  onPrevPage: () => void
  onNextPage: () => void
  onLoad: () => void
  reqSummary: string
}

export function TopBar({
  pageNum,
  onPageNumChange,
  onPrevPage,
  onNextPage,
  onLoad,
  reqSummary,
}: TopBarProps) {
  return (
    <div className="p-3 border-b border-gray-200 bg-white">
      <div className="flex gap-4 items-center">
        <div className="flex items-center gap-2">
          <label htmlFor="pageNum" className="text-xs opacity-80">
            Page
          </label>
          <input
            id="pageNum"
            type="number"
            min="0"
            className="w-20 border border-gray-300 rounded-lg px-2.5 py-1.5 font-mono text-xs box-border"
            value={pageNum}
            onChange={(e) => onPageNumChange(parseInt(e.target.value) || 0)}
          />
        </div>

        <div className="flex gap-2">
          <button
            className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
            onClick={onPrevPage}
          >
            ←
          </button>
          <button
            className="px-3 py-1.5 bg-black text-white border border-black rounded-lg hover:bg-gray-800 font-semibold text-xs"
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

        <div className="flex-1">
          <div className="text-xs opacity-75 font-mono">{reqSummary}</div>
        </div>
      </div>
    </div>
  )
}
