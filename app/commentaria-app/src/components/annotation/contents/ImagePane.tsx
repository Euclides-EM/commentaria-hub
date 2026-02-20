import { useState } from 'react'
import { useAppState } from '../../../context/useAppState.ts'
import ImageZoom from 'react-image-zooom'
import { RangeInput } from '../../core/RangeInput.tsx'

export function ImagePane() {
  const {
    state: { datasetId, currentPageOrKey },
  } = useAppState()
  const [zoom, setZoom] = useState(250)
  const normalizedKey = (() => {
    const num = Number(currentPageOrKey)

    if (!Number.isNaN(num) && Number.isInteger(num)) {
      return `page-${String(num).padStart(4, "0")}.png`
    }

    const key = String(currentPageOrKey)
    return key.endsWith(".png") ? key : `${key}.png`
  })()

  const imageUrl = `${import.meta.env.VITE_BACKEND_URL}/store/data/${datasetId}/imgs/${normalizedKey}`
  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 bg-gray-50 flex items-center flex-wrap gap-2.5">
        <div className="text-sm font-semibold grow min-w-0">
          Page {currentPageOrKey} Facsimile
        </div>
        <button
          type="button"
          onClick={() => window.open(imageUrl, '_blank', 'noopener,noreferrer')}
          className="h-7 w-7 shrink-0 rounded-md bg-white border border-gray-200 text-gray-600 hover:text-gray-800 hover:bg-white shadow-sm flex items-center justify-center text-sm"
          title="Open image in new tab"
          aria-label="Open image in new tab"
        >
          ⤢
        </button>
        <RangeInput
          label="Zoom control"
          value={zoom}
          min={105}
          max={1000}
          step={5}
          onChange={(value) => setZoom(Math.round(value))}
          className="bg-transparent border-gray-300 order-3 basis-full min-w-0"
        />
      </div>

      <div className="flex-1 min-h-0 flex items-center justify-center p-2">
        <div className="h-full w-full max-h-full max-w-full flex items-center justify-center">
          <ImageZoom
            src={imageUrl}
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
