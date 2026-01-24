import { useAppState } from '../context/useAppState'
import { OpenAPI } from '../api'
import ImageZoom from 'react-image-zooom'

export function ImagePane() {
  const {
    state: { datasetId, currentPage },
  } = useAppState()
  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Page {currentPage} Facsimile</div>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden flex items-center justify-center p-0">
        <ImageZoom
          src={`${OpenAPI.BASE}/datasets/${datasetId}/images/${currentPage}`}
          alt="Page image"
          zoom="250"
          width="100%"
          height="auto"
        />
      </div>
    </section>
  )
}
