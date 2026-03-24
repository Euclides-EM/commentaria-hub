import {
  type ChangeEvent,
  type PointerEvent as ReactPointerEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import Select from 'react-select'
import useLocalStorageState from 'use-local-storage-state'
import { useAppState } from '../../../context/useAppState.ts'
import { RangeInput } from '../../core/RangeInput.tsx'
import {
  useDatasetImageKeysQuery,
  useReplaceDatasetImageMutation,
} from '../../../queries/datasets.ts'
import type { TeiSurfaceZone } from './tei/tei.ts'
import ImageZoom from 'react-image-zooom'
import {
  findMatchingImage,
  hasAnnotationPages,
  TITLE_PAGES_DATASET_ID,
} from '../../../utils/editions.ts'
import { useAuthStore } from '../../../store/authStore.ts'
import { Button } from '../../core/Button.tsx'
import { ReplaceImageModal } from '../../modal/ReplaceImageModal.tsx'
import { ApiError } from '@hub-api'
import { selectStyles } from '../../../styles/selectStyles.ts'
import type { StylesConfig } from 'react-select'
import { DEFAULT_IMAGE_ZOOM } from '../imageZoom.ts'

type RenderedImageRect = {
  left: number
  top: number
  width: number
  height: number
  naturalWidth: number
  naturalHeight: number
}

type HighlightMode = 'hide' | 'hover' | 'show'
type HighlightModeOption = { value: HighlightMode; label: string }

const HIGHLIGHT_MODE_OPTIONS: HighlightModeOption[] = [
  { value: 'hide', label: 'Hide' },
  { value: 'hover', label: 'On hover' },
  { value: 'show', label: 'Show' },
]

const baseCompactSelectStyles = selectStyles<HighlightModeOption>({
  controlWidth: '100%',
})

const compactSelectStyles: StylesConfig<HighlightModeOption, false> = {
  ...baseCompactSelectStyles,
}

compactSelectStyles.control = (base, state) => ({
  ...baseCompactSelectStyles.control?.(base, state),
  minHeight: 26,
  height: 26,
})

compactSelectStyles.valueContainer = (base, props) => ({
  ...baseCompactSelectStyles.valueContainer?.(base, props),
  height: '24px',
  padding: '0 6px',
})

compactSelectStyles.indicatorsContainer = (base, props) => ({
  ...baseCompactSelectStyles.indicatorsContainer?.(base, props),
  height: '24px',
})

compactSelectStyles.dropdownIndicator = (base, props) => ({
  ...baseCompactSelectStyles.dropdownIndicator?.(base, props),
  padding: '2px 4px',
})

const getRenderedImageRect = (
  image: HTMLImageElement,
  viewportRect: DOMRect,
): RenderedImageRect => {
  const imageRect = image.getBoundingClientRect()
  const naturalWidth = image.naturalWidth
  const naturalHeight = image.naturalHeight

  if (
    naturalWidth > 0 &&
    naturalHeight > 0 &&
    imageRect.width > 0 &&
    imageRect.height > 0
  ) {
    const containerRatio = imageRect.width / imageRect.height
    const imageRatio = naturalWidth / naturalHeight
    let renderedWidth = imageRect.width
    let renderedHeight = imageRect.height
    let offsetLeft = 0
    let offsetTop = 0

    if (imageRatio > containerRatio) {
      renderedHeight = imageRect.width / imageRatio
      offsetTop = (imageRect.height - renderedHeight) / 2
    } else {
      renderedWidth = imageRect.height * imageRatio
      offsetLeft = (imageRect.width - renderedWidth) / 2
    }

    return {
      left: imageRect.left - viewportRect.left + offsetLeft,
      top: imageRect.top - viewportRect.top + offsetTop,
      width: renderedWidth,
      height: renderedHeight,
      naturalWidth,
      naturalHeight,
    }
  }

  return {
    left: imageRect.left - viewportRect.left,
    top: imageRect.top - viewportRect.top,
    width: imageRect.width,
    height: imageRect.height,
    naturalWidth,
    naturalHeight,
  }
}

type ImagePaneProps = {
  showResizeHandle?: boolean
  onResizeStart?: () => void
  surfaceZones?: TeiSurfaceZone[]
  activeLineMatchIds?: string[]
  onHoverLineMatchIds?: (ids: string[]) => void
}

export function ImagePane({
  showResizeHandle = false,
  onResizeStart,
  surfaceZones = [],
  activeLineMatchIds = [],
  onHoverLineMatchIds,
}: ImagePaneProps) {
  const {
    annotation,
    state: { datasetId, currentPageOrKey },
  } = useAppState()
  const [zoom, setZoom] = useState(DEFAULT_IMAGE_ZOOM)
  const [isImageZoomEngaged, setIsImageZoomEngaged] = useState(false)
  const [isReplaceModalOpen, setIsReplaceModalOpen] = useState(false)
  const [replaceError, setReplaceError] = useState<string | null>(null)
  const [imageVersion, setImageVersion] = useState(0)
  const [highlightMode, setHighlightMode] = useLocalStorageState<HighlightMode>(
    'imagePaneHighlightMode',
    { defaultValue: 'hover', storageSync: false },
  )
  const viewportRef = useRef<HTMLDivElement | null>(null)
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const lastHoverIdsKeyRef = useRef('')
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const replaceImageMutation = useReplaceDatasetImageMutation()
  const [imageDisplayBox, setImageDisplayBox] = useState({
    left: 0,
    top: 0,
    width: 0,
    height: 0,
    naturalWidth: 0,
    naturalHeight: 0,
  })
  const hasPages = hasAnnotationPages(annotation)
  const isKeyNavigation = !!annotation && !hasPages
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    datasetId,
    isKeyNavigation,
  )
  const matchedImage = findMatchingImage(String(currentPageOrKey), imageKeys)
  const currentImageName = matchedImage?.key || String(currentPageOrKey)
  const highlightableZones = useMemo(
    () =>
      surfaceZones.filter(
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
      ),
    [surfaceZones],
  )
  const hasHighlightableZones = highlightableZones.length > 0
  const selectedHighlightMode =
    HIGHLIGHT_MODE_OPTIONS.find((option) => option.value === highlightMode) ||
    HIGHLIGHT_MODE_OPTIONS[1]
  const isHighlightInteractionEnabled = highlightMode !== 'hide'
  const normalizedKey = (() => {
    const num = Number(currentPageOrKey)

    if (!Number.isNaN(num) && Number.isInteger(num)) {
      return `page-${String(num).padStart(4, '0')}.png`
    }

    return matchedImage?.filename || ''
  })()

  const imageUrl = `${import.meta.env.VITE_BACKEND_URL}/store/data/${datasetId}/imgs/${normalizedKey}?v=${imageVersion}`

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
      const renderedRect = getRenderedImageRect(image, viewportRect)
      const width = renderedRect.width
      const height = renderedRect.height
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
      setImageDisplayBox({
        left: renderedRect.left,
        top: renderedRect.top,
        width,
        height,
        naturalWidth: renderedRect.naturalWidth,
        naturalHeight: renderedRect.naturalHeight,
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
    return highlightableZones
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
    highlightableZones,
    imageDisplayBox.height,
    imageDisplayBox.left,
    imageDisplayBox.naturalHeight,
    imageDisplayBox.naturalWidth,
    imageDisplayBox.top,
    imageDisplayBox.width,
  ])

  useEffect(() => {
    if (isImageZoomEngaged || !isHighlightInteractionEnabled) {
      onHoverLineMatchIds?.([])
      lastHoverIdsKeyRef.current = ''
    }
  }, [isHighlightInteractionEnabled, isImageZoomEngaged, onHoverLineMatchIds])

  useEffect(() => {
    if (!isHighlightInteractionEnabled || !hasHighlightableZones) {
      onHoverLineMatchIds?.([])
      lastHoverIdsKeyRef.current = ''
    }
  }, [
    hasHighlightableZones,
    isHighlightInteractionEnabled,
    onHoverLineMatchIds,
  ])

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
    if (!isHighlightInteractionEnabled || !onHoverLineMatchIds) {
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
    if (!isHighlightInteractionEnabled || !onHoverLineMatchIds) {
      return
    }
    emitHoverIds([])
  }

  const handleReplaceConfirm = () => {
    setReplaceError(null)
    setIsReplaceModalOpen(false)
    fileInputRef.current?.click()
  }

  const handleReplaceCancel = () => {
    if (replaceImageMutation.isPending) {
      return
    }
    setReplaceError(null)
    setIsReplaceModalOpen(false)
  }

  const handleFileChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = ''

    if (!file || !datasetId || !currentPageOrKey) {
      return
    }

    try {
      setReplaceError(null)
      await replaceImageMutation.mutateAsync({
        datasetId,
        key: String(currentPageOrKey),
        type: datasetId === TITLE_PAGES_DATASET_ID ? 'tp' : 'facsimile',
        file,
      })
      setImageVersion((current) => current + 1)
    } catch (error) {
      setReplaceError(
        error instanceof ApiError
          ? typeof error.body === 'string'
            ? error.body
            : JSON.stringify(error.body)
          : error instanceof Error
            ? error.message
            : 'Failed to replace image.',
      )
      setIsReplaceModalOpen(true)
    }
  }

  if (!normalizedKey) {
    return
  }

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 h-full bg-white relative">
      <div className="px-2.5 py-2 border-b border-gray-200 bg-gray-50 flex items-center flex-wrap gap-2.5">
        <div className="text-sm font-semibold grow min-w-0">
          {hasPages ? `Page ${currentPageOrKey} Facsimile` : currentImageName}
        </div>
        <Button
          type="button"
          onClick={() => window.open(imageUrl, '_blank', 'noopener,noreferrer')}
          className="h-7 w-7 shrink-0 rounded-md flex items-center justify-center text-sm"
          title="Open image in new tab"
          aria-label="Open image in new tab"
        >
          ⤢
        </Button>
        {isAuthenticated && (
          <Button
            type="button"
            onClick={() => {
              setReplaceError(null)
              setIsReplaceModalOpen(true)
            }}
            className="px-2 py-1 text-xs"
            disabled={replaceImageMutation.isPending}
          >
            Replace image
          </Button>
        )}
        <div className="order-3 basis-full min-w-0 flex items-center gap-3">
          {hasHighlightableZones && (
            <div className="flex items-center gap-1.5 text-xs font-medium text-gray-700 shrink-0">
              <span>Highlights</span>
              <div className="w-24">
                <Select<HighlightModeOption, false>
                  value={selectedHighlightMode}
                  onChange={(option) =>
                    setHighlightMode(option?.value || 'hover')
                  }
                  options={HIGHLIGHT_MODE_OPTIONS}
                  isClearable={false}
                  styles={compactSelectStyles}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                />
              </div>
            </div>
          )}
          <RangeInput
            label="Zoom control"
            value={zoom}
            min={105}
            max={1000}
            step={5}
            onChange={(value) => setZoom(Math.round(value))}
            className="bg-transparent border-gray-300 min-w-0 flex-1"
          />
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-2">
        <div className="h-full w-full max-h-full max-w-full overflow-hidden flex items-center justify-center">
          <div
            ref={viewportRef}
            className="relative h-full w-full flex justify-center items-center"
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
            {isHighlightInteractionEnabled &&
              !isImageZoomEngaged &&
              visibleZones.length > 0 && (
                <div className="absolute inset-0 z-20 pointer-events-none">
                  {visibleZones.map((zone) => {
                    const isShown = highlightMode === 'show' || zone.isActive
                    if (!isShown) {
                      return null
                    }
                    const className = zone.isActive
                      ? zone.zoneType === 'block'
                        ? 'border-amber-500 bg-amber-300/20'
                        : 'border-teal-500 bg-teal-300/20'
                      : zone.zoneType === 'block'
                        ? 'border-amber-400/60 bg-amber-200/10'
                        : 'border-teal-400/60 bg-teal-200/10'
                    return (
                      <div
                        key={zone.id}
                        className={`absolute border-2 rounded-sm ${className}`}
                        style={{
                          left: `${zone.zoneType === 'line' ? zone.left - 1 : zone.left}px`,
                          top: `${zone.zoneType === 'line' ? zone.top - 1 : zone.top}px`,
                          width: `${zone.zoneType === 'line' ? zone.width + 2 : zone.width}px`,
                          height: `${zone.zoneType === 'line' ? zone.height + 2 : zone.height}px`,
                        }}
                      />
                    )
                  })}
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
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="sr-only"
        onChange={(event) => {
          void handleFileChange(event)
        }}
        disabled={replaceImageMutation.isPending}
      />
      <ReplaceImageModal
        isOpen={isReplaceModalOpen}
        body={`Replace ${
          hasPages ? `page ${currentPageOrKey}` : currentImageName
        } with a new image? This will overwrite the current image for this page.`}
        isReplacing={replaceImageMutation.isPending}
        error={replaceError}
        onCancel={handleReplaceCancel}
        onConfirm={handleReplaceConfirm}
      />
    </section>
  )
}
