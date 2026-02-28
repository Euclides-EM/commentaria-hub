import { useState } from 'react'
import { useAppState } from '../../../context/useAppState.ts'
import ImageZoom from 'react-image-zooom'
import { RangeInput } from '../../core/RangeInput.tsx'
import { useDatasetImageKeysQuery } from '../../../queries/datasets.ts'

type ImagePaneProps = {
  showResizeHandle?: boolean
  onResizeStart?: () => void
}

export function ImagePane({
  showResizeHandle = false,
  onResizeStart,
}: ImagePaneProps) {
  const {
    annotation,
    state: { datasetId, currentPageOrKey },
  } = useAppState()
  const [zoom, setZoom] = useState(250)
  const isKeyNavigation = !!annotation && !annotation.pages
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    datasetId,
    isKeyNavigation,
  )
  const currentImageName =
    imageKeys.find((image) => image.key === String(currentPageOrKey))?.key ||
    String(currentPageOrKey)
  const normalizedKey = (() => {
    const num = Number(currentPageOrKey)

    if (!Number.isNaN(num) && Number.isInteger(num)) {
      return `page-${String(num).padStart(4, '0')}.png`
    }

    const key = String(currentPageOrKey)
    const matchedImage = imageKeys.find((image) => image.key === key)
    if (matchedImage?.filename) {
      return matchedImage.filename
    }

    if (/\.(png|jpe?g|webp|gif|bmp|tiff?)$/i.test(key)) {
      return key
    }

    return `${key}.png`
  })()

  const imageUrl = `${import.meta.env.VITE_BACKEND_URL}/store/data/${datasetId}/imgs/${normalizedKey}`

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 h-full bg-white relative">
      <div className="px-2.5 py-2 border-b border-gray-200 bg-gray-50 flex items-center flex-wrap gap-2.5">
        <div className="text-sm font-semibold grow min-w-0">
          {annotation?.pages
            ? `Page ${currentPageOrKey} Facsimile`
            : currentImageName}
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

      <div className="flex-1 min-h-0 overflow-hidden flex items-center justify-center p-2">
        <div className="h-full w-full max-h-full max-w-full overflow-hidden flex items-center justify-center">
          <ImageZoom
            src={imageUrl}
            alt="Page image"
            zoom={String(zoom)}
            width="100%"
            height="100%"
            className="max-h-full max-w-full w-full h-full overflow-hidden [&_img]:h-full [&_img]:w-full [&_img]:max-w-none [&_img]:max-h-none [&_img]:object-cover"
          />
        </div>
      </div>
      {showResizeHandle && (
        <div
          role="separator"
          aria-label="Resize image pane"
          className="absolute top-0 right-0 h-full w-2 cursor-col-resize flex items-center justify-center bg-gray-100 hover:bg-gray-200 transition-colors"
          onPointerDown={(event) => {
            event.preventDefault()
            onResizeStart?.()
          }}
        >
          <div className="h-10 w-0.5 rounded-full bg-gray-300" />
        </div>
      )}
    </section>
  )
}
