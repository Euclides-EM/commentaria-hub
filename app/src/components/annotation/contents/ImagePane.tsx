import { useState } from 'react'
import { useAppState } from '../../../context/useAppState.ts'
import { OpenAPI } from '../../../api'
import ImageZoom from 'react-image-zooom'
import { RangeInput } from '../../core/RangeInput.tsx'

export function ImagePane() {
  const {
    state: { datasetId, currentPage },
  } = useAppState()
  const [zoom, setZoom] = useState(250)
  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200  bg-gray-50 flex items-center justify-between gap-2.5">
        <div className="flex items-center gap-3 flex-wrap">
          <div className="text-sm font-semibold">
            Page {currentPage} Facsimile
          </div>
          <RangeInput
            label="Zoom control"
            value={zoom}
            min={105}
            max={1000}
            step={5}
            onChange={(value) => setZoom(Math.round(value))}
            className="bg-transparent border-gray-300"
          />
        </div>
        <button
          type="button"
          onClick={() =>
            window.open(
              `${OpenAPI.BASE}/datasets/${datasetId}/images/${currentPage}`,
              '_blank',
              'noopener,noreferrer',
            )
          }
          className="h-7 w-7 rounded-md bg-white border border-gray-200 text-gray-600 hover:text-gray-800 hover:bg-white shadow-sm flex items-center justify-center text-sm"
          title="Open image in new tab"
          aria-label="Open image in new tab"
        >
          ⤢
        </button>
      </div>

      <div className="flex-1 min-h-0 flex items-center justify-center p-2">
        <div className="h-full w-full max-h-full max-w-full flex items-center justify-center">
          <ImageZoom
            src={`${OpenAPI.BASE}/datasets/${datasetId}/images/${currentPage}`}
            alt="Page image"
            zoom={String(zoom)}
            width="100%"
            height="100%"
            className="max-h-full max-w-full w-full h-full [&_img]:max-h-full [&_img]:max-w-full [&_img]:h-full [&_img]:w-full [&_img]:object-contain"
          />
        </div>
      </div>
    </section>
  )
}
