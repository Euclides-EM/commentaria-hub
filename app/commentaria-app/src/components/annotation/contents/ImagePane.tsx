import { useEffect, useMemo, useRef, useState } from 'react'
import { useAppState } from '../../../context/useAppState.ts'
import { RangeInput } from '../../core/RangeInput.tsx'
import { useDatasetImageKeysQuery } from '../../../queries/datasets.ts'
import type { TeiSurfaceZone } from './tei/tei.ts'
import ImageZoom from 'react-image-zooom'

type ImagePaneProps = {
  showResizeHandle?: boolean
  onResizeStart?: () => void
  surfaceZones?: TeiSurfaceZone[]
  activeLineMatchIds?: string[]
  enableHoverSync?: boolean
  onHoverLineMatchIds?: (ids: string[]) => void
}

export function ImagePane({
  showResizeHandle = false,
  onResizeStart,
  surfaceZones = [],
  activeLineMatchIds = [],
  enableHoverSync = true,
  onHoverLineMatchIds,
}: ImagePaneProps) {
  const {
    annotation,
    state: { datasetId, currentPageOrKey },
  } = useAppState()
  const [zoom, setZoom] = useState(250)
  const viewportRef = useRef<HTMLDivElement | null>(null)
  const [imageDisplayBox, setImageDisplayBox] = useState({
    left: 0,
    top: 0,
    width: 0,
    height: 0,
  })
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

  useEffect(() => {
    const viewport = viewportRef.current
    if (!viewport) {
      return
    }

    const getImageElement = () =>
      viewport.querySelector('img#imageZoom') || viewport.querySelector('img')

    const recalc = () => {
      const image = getImageElement()
      if (!(image instanceof HTMLImageElement)) {
        setImageDisplayBox({ left: 0, top: 0, width: 0, height: 0 })
        return
      }
      const viewportRect = viewport.getBoundingClientRect()
      const imageRect = image.getBoundingClientRect()
      const width = imageRect.width
      const height = imageRect.height
      if (!width || !height) {
        setImageDisplayBox({ left: 0, top: 0, width: 0, height: 0 })
        return
      }
      const left = imageRect.left - viewportRect.left
      const top = imageRect.top - viewportRect.top
      setImageDisplayBox({ left, top, width, height })
    }

    const frame = window.requestAnimationFrame(recalc)
    const resizeObserver = new ResizeObserver(recalc)
    resizeObserver.observe(viewport)
    let observedImage: HTMLImageElement | null = null
    const attachImage = () => {
      const image = getImageElement()
      if (!(image instanceof HTMLImageElement)) {
        return
      }
      if (observedImage === image) {
        return
      }
      if (observedImage) {
        resizeObserver.unobserve(observedImage)
        observedImage.removeEventListener('load', recalc)
      }
      observedImage = image
      resizeObserver.observe(image)
      image.addEventListener('load', recalc)
      recalc()
    }
    const mutationObserver = new MutationObserver(() => {
      attachImage()
      recalc()
    })
    mutationObserver.observe(viewport, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['src', 'style', 'class'],
    })
    attachImage()
    window.addEventListener('resize', recalc)
    return () => {
      window.cancelAnimationFrame(frame)
      if (observedImage) {
        observedImage.removeEventListener('load', recalc)
      }
      mutationObserver.disconnect()
      resizeObserver.disconnect()
      window.removeEventListener('resize', recalc)
    }
  }, [imageUrl, zoom])

  const activeMatchIdSet = useMemo(
    () => new Set(activeLineMatchIds),
    [activeLineMatchIds],
  )
  const visibleZones = useMemo(() => {
    if (!imageDisplayBox.width || !imageDisplayBox.height) {
      return []
    }
    return surfaceZones
      .filter(
        (zone) =>
          Number.isFinite(zone.ulx) &&
          Number.isFinite(zone.uly) &&
          Number.isFinite(zone.lrx) &&
          Number.isFinite(zone.lry) &&
          Number.isFinite(zone.refUlx) &&
          Number.isFinite(zone.refUly) &&
          Number.isFinite(zone.refLrx) &&
          Number.isFinite(zone.refLry) &&
          zone.lrx > zone.ulx &&
          zone.lry > zone.uly &&
          zone.refLrx > zone.refUlx &&
          zone.refLry > zone.refUly,
      )
      .map((zone) => {
        const referenceWidth = zone.refLrx - zone.refUlx
        const referenceHeight = zone.refLry - zone.refUly
        const left =
          imageDisplayBox.left +
          ((zone.ulx - zone.refUlx) / referenceWidth) * imageDisplayBox.width
        const top =
          imageDisplayBox.top +
          ((zone.uly - zone.refUly) / referenceHeight) * imageDisplayBox.height
        const width =
          ((zone.lrx - zone.ulx) / referenceWidth) * imageDisplayBox.width
        const height =
          ((zone.lry - zone.uly) / referenceHeight) * imageDisplayBox.height
        const isActive = zone.matchIds.some((id) => activeMatchIdSet.has(id))
        return {
          ...zone,
          left,
          top,
          width,
          height,
          isActive,
        }
      })
      .filter(
        (zone) =>
          zone.width > 0 &&
          zone.height > 0 &&
          zone.left < imageDisplayBox.left + imageDisplayBox.width &&
          zone.top < imageDisplayBox.top + imageDisplayBox.height &&
          zone.left + zone.width > imageDisplayBox.left &&
          zone.top + zone.height > imageDisplayBox.top,
      )
  }, [
    activeMatchIdSet,
    imageDisplayBox.height,
    imageDisplayBox.left,
    imageDisplayBox.top,
    imageDisplayBox.width,
    surfaceZones,
  ])

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

      <div className="flex-1 min-h-0 overflow-auto p-2">
        <div className="h-full w-full max-h-full max-w-full overflow-hidden flex items-center justify-center">
          <div ref={viewportRef} className="relative h-full w-full">
            <ImageZoom
              key={imageUrl}
              src={imageUrl}
              alt="Page image"
              zoom={String(zoom)}
              width="100%"
              height="100%"
              className="max-h-full max-w-full w-full h-full overflow-hidden [&_img]:h-full [&_img]:w-full [&_img]:max-w-none [&_img]:max-h-none [&_img]:object-contain"
            />
            {enableHoverSync && visibleZones.length > 0 && (
              <div className="absolute inset-0 z-20 pointer-events-none">
                {visibleZones.map((zone) => (
                  <button
                    key={zone.id}
                    type="button"
                    tabIndex={-1}
                    aria-label={`Zone ${zone.id}`}
                    className={`absolute border-2 rounded-sm ${
                      zone.isActive
                        ? zone.zoneType === 'block'
                          ? 'border-amber-500 bg-amber-300/20'
                          : 'border-teal-500 bg-teal-300/20'
                        : 'border-transparent bg-transparent'
                    } pointer-events-auto`}
                    style={{
                      left: `${zone.zoneType === 'line' ? zone.left - 1 : zone.left}px`,
                      top: `${zone.zoneType === 'line' ? zone.top - 1 : zone.top}px`,
                      width: `${zone.zoneType === 'line' ? zone.width + 2 : zone.width}px`,
                      height: `${zone.zoneType === 'line' ? zone.height + 2 : zone.height}px`,
                    }}
                    onMouseEnter={() =>
                      onHoverLineMatchIds?.(
                        zone.hoverMatchIds.length > 0
                          ? zone.hoverMatchIds
                          : zone.matchIds,
                      )
                    }
                    onMouseLeave={() => onHoverLineMatchIds?.([])}
                  />
                ))}
              </div>
            )}
          </div>
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
