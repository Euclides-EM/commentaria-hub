import {
  type PointerEvent as ReactPointerEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
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
  const [isImageZoomEngaged, setIsImageZoomEngaged] = useState(false)
  const viewportRef = useRef<HTMLDivElement | null>(null)
  const lastHoverIdsKeyRef = useRef('')
  const [imageDisplayBox, setImageDisplayBox] = useState({
    left: 0,
    top: 0,
    width: 0,
    height: 0,
    naturalWidth: 0,
    naturalHeight: 0,
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
    const isImageZoomedNow = () => {
      if (viewport.querySelector('.zoomed')) {
        return true
      }
      const image = getImageElement()
      if (!(image instanceof HTMLImageElement)) {
        return false
      }
      const style = window.getComputedStyle(image)
      if (style.cursor.includes('zoom-out')) {
        return true
      }
      if (style.transform && style.transform !== 'none') {
        return true
      }
      const imageRect = image.getBoundingClientRect()
      const viewportRect = viewport.getBoundingClientRect()
      return (
        imageRect.width > viewportRect.width + 1 ||
        imageRect.height > viewportRect.height + 1
      )
    }
    const syncZoomEngaged = () => {
      setIsImageZoomEngaged(isImageZoomedNow())
    }
    const scheduleSyncZoomEngaged = () =>
      window.requestAnimationFrame(syncZoomEngaged)

    const recalc = () => {
      const image = getImageElement()
      syncZoomEngaged()
      if (!(image instanceof HTMLImageElement)) {
        setImageDisplayBox({
          left: 0,
          top: 0,
          width: 0,
          height: 0,
          naturalWidth: 0,
          naturalHeight: 0,
        })
        return
      }
      const viewportRect = viewport.getBoundingClientRect()
      const imageRect = image.getBoundingClientRect()
      const width = imageRect.width
      const height = imageRect.height
      if (!width || !height) {
        setImageDisplayBox({
          left: 0,
          top: 0,
          width: 0,
          height: 0,
          naturalWidth: 0,
          naturalHeight: 0,
        })
        return
      }
      const left = imageRect.left - viewportRect.left
      const top = imageRect.top - viewportRect.top
      setImageDisplayBox({
        left,
        top,
        width,
        height,
        naturalWidth: image.naturalWidth,
        naturalHeight: image.naturalHeight,
      })
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
      syncZoomEngaged()
    })
    mutationObserver.observe(viewport, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['src', 'style', 'class'],
    })
    attachImage()
    window.addEventListener('resize', recalc)
    viewport.addEventListener('pointermove', scheduleSyncZoomEngaged)
    viewport.addEventListener('click', scheduleSyncZoomEngaged, true)
    return () => {
      window.cancelAnimationFrame(frame)
      if (observedImage) {
        observedImage.removeEventListener('load', recalc)
      }
      mutationObserver.disconnect()
      resizeObserver.disconnect()
      window.removeEventListener('resize', recalc)
      viewport.removeEventListener('pointermove', scheduleSyncZoomEngaged)
      viewport.removeEventListener('click', scheduleSyncZoomEngaged, true)
      setIsImageZoomEngaged(false)
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
        const useNaturalBounds =
          !zone.hasSurfaceBounds &&
          imageDisplayBox.naturalWidth > 0 &&
          imageDisplayBox.naturalHeight > 0
        const referenceUlx = useNaturalBounds ? 0 : zone.refUlx
        const referenceUly = useNaturalBounds ? 0 : zone.refUly
        const referenceLrx = useNaturalBounds
          ? imageDisplayBox.naturalWidth
          : zone.refLrx
        const referenceLry = useNaturalBounds
          ? imageDisplayBox.naturalHeight
          : zone.refLry
        const referenceWidth = referenceLrx - referenceUlx
        const referenceHeight = referenceLry - referenceUly
        const left =
          imageDisplayBox.left +
          ((zone.ulx - referenceUlx) / referenceWidth) * imageDisplayBox.width
        const top =
          imageDisplayBox.top +
          ((zone.uly - referenceUly) / referenceHeight) * imageDisplayBox.height
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
    imageDisplayBox.naturalHeight,
    imageDisplayBox.naturalWidth,
    imageDisplayBox.top,
    imageDisplayBox.width,
    surfaceZones,
  ])

  useEffect(() => {
    if (isImageZoomEngaged) {
      onHoverLineMatchIds?.([])
      lastHoverIdsKeyRef.current = ''
    }
  }, [isImageZoomEngaged, onHoverLineMatchIds])

  const emitHoverIds = (ids: string[]) => {
    const key = ids.join('|')
    if (key === lastHoverIdsKeyRef.current) {
      return
    }
    lastHoverIdsKeyRef.current = key
    onHoverLineMatchIds?.(ids)
  }

  const handleViewportPointerMove = (
    event: ReactPointerEvent<HTMLDivElement>,
  ) => {
    if (!enableHoverSync || !onHoverLineMatchIds) {
      return
    }
    if (isImageZoomEngaged) {
      emitHoverIds([])
      return
    }
    const rect = event.currentTarget.getBoundingClientRect()
    const x = event.clientX - rect.left
    const y = event.clientY - rect.top
    let hovered: (typeof visibleZones)[number] | null = null
    for (let index = visibleZones.length - 1; index >= 0; index--) {
      const zone = visibleZones[index]
      if (
        x >= zone.left &&
        x <= zone.left + zone.width &&
        y >= zone.top &&
        y <= zone.top + zone.height
      ) {
        hovered = zone
        break
      }
    }
    if (!hovered) {
      emitHoverIds([])
      return
    }
    emitHoverIds(
      hovered.hoverMatchIds.length > 0
        ? hovered.hoverMatchIds
        : hovered.matchIds,
    )
  }

  const handleViewportPointerLeave = () => {
    if (!enableHoverSync || !onHoverLineMatchIds) {
      return
    }
    emitHoverIds([])
  }

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
          <div
            ref={viewportRef}
            className="relative h-full w-full"
            onPointerMove={handleViewportPointerMove}
            onPointerLeave={handleViewportPointerLeave}
          >
            <ImageZoom
              key={imageUrl}
              src={imageUrl}
              alt="Page image"
              zoom={String(zoom)}
              width="100%"
              height="100%"
              className="max-h-full max-w-full w-full h-full overflow-hidden [&_img]:h-full [&_img]:w-full [&_img]:max-w-none [&_img]:max-h-none [&_img]:object-contain"
            />
            {enableHoverSync &&
              !isImageZoomEngaged &&
              visibleZones.length > 0 && (
                <div className="absolute inset-0 z-20 pointer-events-none">
                  {visibleZones.map((zone) => (
                    <div
                      key={zone.id}
                      className={`absolute border-2 rounded-sm ${
                        zone.isActive
                          ? zone.zoneType === 'block'
                            ? 'border-amber-500 bg-amber-300/20'
                            : 'border-teal-500 bg-teal-300/20'
                          : 'border-transparent bg-transparent'
                      }`}
                      style={{
                        left: `${zone.zoneType === 'line' ? zone.left - 1 : zone.left}px`,
                        top: `${zone.zoneType === 'line' ? zone.top - 1 : zone.top}px`,
                        width: `${zone.zoneType === 'line' ? zone.width + 2 : zone.width}px`,
                        height: `${zone.zoneType === 'line' ? zone.height + 2 : zone.height}px`,
                      }}
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
