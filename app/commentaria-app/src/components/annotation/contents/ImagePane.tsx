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
import {
  findMatchingImage,
  hasAnnotationPages,
  TITLE_PAGES_DATASET_ID,
} from '../../../utils/editions.ts'
import { buildDatasetImageUrl } from '../../../utils/imageUrls.ts'
import { useAuthStore } from '../../../store/authStore.ts'
import { Button } from '../../core/Button.tsx'
import { ReplaceImageModal } from '../../modal/ReplaceImageModal.tsx'
import { ApiError } from '@hub-api'
import { selectStyles } from '../../../styles/selectStyles.ts'
import type { StylesConfig } from 'react-select'
import { DEFAULT_IMAGE_ZOOM } from '../imageZoom.ts'
import { MultiSelectDropdown } from '../../core/MultiSelectDropdown.tsx'
import {
  DEFAULT_HIGHLIGHT_ZONE_FILTERS,
  getHighlightZoneFilterLabel,
  getHighlightZoneFilterPickerLabel,
  HIGHLIGHT_ZONE_FILTER_OPTIONS,
  HIGHLIGHT_ZONE_FILTER_STORAGE_KEY,
  type HighlightZoneFilter,
} from '../highlightControls.ts'
import { computeVisibleZones, filterValidZones } from '../imageZoneUtils.ts'
import { useZoomableImage } from '../useZoomableImage.ts'

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
  const [isReplaceModalOpen, setIsReplaceModalOpen] = useState(false)
  const [replaceError, setReplaceError] = useState<string | null>(null)
  const [imageVersion, setImageVersion] = useState(0)
  const [highlightMode, setHighlightMode] = useLocalStorageState<HighlightMode>(
    'imagePaneHighlightMode',
    { defaultValue: 'hover', storageSync: false },
  )
  const [highlightZoneFilters, setHighlightZoneFilters] = useLocalStorageState<
    HighlightZoneFilter[]
  >(HIGHLIGHT_ZONE_FILTER_STORAGE_KEY, {
    defaultValue: DEFAULT_HIGHLIGHT_ZONE_FILTERS,
    storageSync: false,
  })
  const fileInputRef = useRef<HTMLInputElement | null>(null)
  const lastHoverIdsKeyRef = useRef('')
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const replaceImageMutation = useReplaceDatasetImageMutation()
  const hasPages = hasAnnotationPages(annotation)
  const isKeyNavigation = !!annotation && !hasPages
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    datasetId,
    isKeyNavigation,
  )
  const matchedImage = findMatchingImage(String(currentPageOrKey), imageKeys)
  const currentImageName = matchedImage?.key || String(currentPageOrKey)
  const hasSurfaceZones = surfaceZones.length > 0
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

  const imageUrl = buildDatasetImageUrl(
    datasetId,
    normalizedKey,
    'preview',
    imageVersion,
  )
  const originalImageUrl = buildDatasetImageUrl(
    datasetId,
    normalizedKey,
    'original',
    imageVersion,
  )

  const {
    containerRef,
    imgRef,
    isZoomed,
    zoomTransform,
    imageDisplayBox,
    handleContainerClick,
    updateCursor,
    getLocalCursor,
  } = useZoomableImage(imageUrl, zoom)

  const activeMatchIdSet = useMemo(
    () => new Set(activeLineMatchIds),
    [activeLineMatchIds],
  )
  const highlightableZones = useMemo(
    () => filterValidZones(surfaceZones, highlightZoneFilters),
    [highlightZoneFilters, surfaceZones],
  )
  const hasHighlightableZones = highlightableZones.length > 0
  const visibleZones = useMemo(
    () =>
      computeVisibleZones(
        highlightableZones,
        imageDisplayBox,
        activeMatchIdSet,
      ),
    [activeMatchIdSet, highlightableZones, imageDisplayBox],
  )

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
    const rect = event.currentTarget.getBoundingClientRect()
    const cx = event.clientX - rect.left
    const cy = event.clientY - rect.top

    updateCursor(cx, cy)

    if (!isHighlightInteractionEnabled || !onHoverLineMatchIds) {
      return
    }

    const { localX, localY } = getLocalCursor(cx, cy)
    let hovered: (typeof visibleZones)[number] | null = null
    for (let index = visibleZones.length - 1; index >= 0; index--) {
      const zone = visibleZones[index]
      if (
        localX >= zone.left &&
        localX <= zone.left + zone.width &&
        localY >= zone.top &&
        localY <= zone.top + zone.height
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
          onClick={() =>
            window.open(originalImageUrl, '_blank', 'noopener,noreferrer')
          }
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
        <div className="order-3 basis-full min-w-0 flex items-center gap-3 flex-wrap">
          {hasSurfaceZones && (
            <div className="flex items-center gap-3 text-xs font-medium text-gray-700 shrink-0 flex-wrap">
              <div className="flex items-center gap-1.5 shrink-0">
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
              <div className="flex items-center gap-1.5 shrink-0">
                <MultiSelectDropdown<HighlightZoneFilter>
                  allItems={HIGHLIGHT_ZONE_FILTER_OPTIONS}
                  selectedItems={highlightZoneFilters}
                  setSelectedItems={(items) =>
                    setHighlightZoneFilters(items ?? [])
                  }
                  itemsLabel="highlights"
                  getItemLabel={getHighlightZoneFilterLabel}
                  getPickerLabel={getHighlightZoneFilterPickerLabel}
                  showBulkActions={false}
                  minWidth="120px"
                />
              </div>
            </div>
          )}
          <div className="min-w-[220px] flex-1">
            <RangeInput
              label="Zoom control"
              value={zoom}
              min={105}
              max={1000}
              step={5}
              onChange={(value) => setZoom(Math.round(value))}
              className="bg-transparent border-gray-300 min-w-0 w-full"
            />
          </div>
        </div>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden p-2">
        <div className="h-full w-full overflow-hidden flex items-center justify-center">
          <div
            ref={containerRef}
            className={`relative h-full w-full overflow-hidden select-none ${isZoomed ? 'cursor-zoom-out' : 'cursor-zoom-in'}`}
            onClick={handleContainerClick}
            onPointerMove={handleViewportPointerMove}
            onPointerLeave={handleViewportPointerLeave}
          >
            <div
              className="absolute inset-0"
              style={{ transform: zoomTransform, transformOrigin: '0 0' }}
            >
              <img
                ref={imgRef}
                key={imageUrl}
                src={imageUrl}
                alt="Page image"
                className="w-full h-full object-contain"
              />
              {isHighlightInteractionEnabled && visibleZones.length > 0 && (
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
