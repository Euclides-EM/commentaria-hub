import React, { useEffect, useMemo, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import Select from 'react-select'
import ImageZoom from 'react-image-zooom'
import { useAppState } from '../../../context/useAppState.ts'
import {
  useAnnotationCategories,
  useAnnotationSearch,
  useAnnotationTeisQuery,
} from '../../../queries/annotations.ts'
import {
  useDatasetFeaturesQuery,
  useDatasetImageKeysQuery,
} from '../../../queries/datasets.ts'
import { AnnotationNavigation } from '../contents/AnnotationNavigation.tsx'
import { TeiContentView } from '../contents/tei/TeiContentView.tsx'
import { TeiDisplayControls } from '../contents/tei/TeiDisplayControls.tsx'
import {
  getTeiHighlightCategories,
  getTeiSurfaceZones,
  getTeiTranslations,
  hasTeiCertaintyDegrees,
  parseTeisXml,
  type TeiHighlightConfig,
  type TeiSurfaceZone,
  type TeiViewMode,
} from '../contents/tei/tei.ts'
import {
  isVerbLike,
  matchTeiCategoryToFeature,
  toFeatureOptions,
  VIEW_LABEL_MAP,
} from '../contents/tei/teiPaneUtils.tsx'
import type { ResolvedTeiFeature } from '../contents/tei/TeiPane.types.ts'
import { expandRange } from '../../../utils/pages.ts'
import { selectStyles } from '../../../styles/selectStyles.ts'
import { RangeInput } from '../../core/RangeInput.tsx'
import { DEFAULT_IMAGE_ZOOM } from '../imageZoom.ts'
import type { annotation_SearchWithin } from '@hub-api'
import {
  ANNOTATION_SEARCH_CATEGORIES_KEY,
  ANNOTATION_SEARCH_TERM_KEY,
  ANNOTATION_SEARCH_WITHIN_KEY,
  getSearchResultPageOrKey,
} from '../contents/navigation/annotationSearchUtils.ts'
import { hasAnnotationPages } from '../../../utils/editions.ts'

type GalleryViewMode = 'images' | 'texts' | 'side-by-side'
type ViewModeOption = { value: GalleryViewMode; label: string }
type TeiViewModeOption = { value: TeiViewMode; label: string }
type HighlightMode = 'hide' | 'hover' | 'show'
type HighlightModeOption = { value: HighlightMode; label: string }

const BATCH_SIZE = 20

const VIEW_MODE_OPTIONS: ViewModeOption[] = [
  { value: 'images', label: 'Images' },
  { value: 'texts', label: 'Texts' },
  { value: 'side-by-side', label: 'Side by side' },
]

const HIGHLIGHT_MODE_OPTIONS: HighlightModeOption[] = [
  { value: 'hide', label: 'Hide' },
  { value: 'hover', label: 'On hover' },
  { value: 'show', label: 'Show' },
]

const annotationSearchWithinOptions: annotation_SearchWithin[] = [
  'categories',
  'transcription',
  'translation',
  'biblio_metadata',
]

type RenderedImageRect = {
  left: number
  top: number
  width: number
  height: number
  naturalWidth: number
  naturalHeight: number
}

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

type GalleryImageCardProps = {
  imageUrl: string
  imageZoom: number
  highlightMode: HighlightMode
  surfaceZones: TeiSurfaceZone[]
  half: boolean
  activeLineMatchIds: string[]
  onHoverLineMatchIds: (ids: string[]) => void
}

function GalleryImageCard({
  imageUrl,
  imageZoom,
  highlightMode,
  surfaceZones,
  half,
  activeLineMatchIds,
  onHoverLineMatchIds,
}: GalleryImageCardProps) {
  const viewportRef = useRef<HTMLDivElement | null>(null)
  const [isImageZoomEngaged, setIsImageZoomEngaged] = useState(false)
  const [imageDisplayBox, setImageDisplayBox] = useState<RenderedImageRect>({
    left: 0,
    top: 0,
    width: 0,
    height: 0,
    naturalWidth: 0,
    naturalHeight: 0,
  })

  useEffect(() => {
    const viewport = viewportRef.current
    if (!viewport) return
    const getImageElement = () =>
      viewport.querySelector('img#imageZoom') || viewport.querySelector('img')
    const isImageZoomedNow = () => {
      if (viewport.querySelector('.zoomed')) return true
      const image = getImageElement()
      if (!(image instanceof HTMLImageElement)) return false
      const style = window.getComputedStyle(image)
      if (style.cursor.includes('zoom-out')) return true
      if (style.transform && style.transform !== 'none') return true
      const imageRect = image.getBoundingClientRect()
      const vpRect = viewport.getBoundingClientRect()
      return (
        imageRect.width > vpRect.width + 1 ||
        imageRect.height > vpRect.height + 1
      )
    }
    const syncZoomEngaged = () => setIsImageZoomEngaged(isImageZoomedNow())
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
      const vpRect = viewport.getBoundingClientRect()
      const rect = getRenderedImageRect(image, vpRect)
      if (!rect.width || !rect.height) {
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
      setImageDisplayBox(rect)
    }
    const frame = window.requestAnimationFrame(recalc)
    const resizeObserver = new ResizeObserver(recalc)
    resizeObserver.observe(viewport)
    let observedImage: HTMLImageElement | null = null
    const attachImage = () => {
      const image = getImageElement()
      if (!(image instanceof HTMLImageElement) || observedImage === image)
        return
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
      if (observedImage) observedImage.removeEventListener('load', recalc)
      mutationObserver.disconnect()
      resizeObserver.disconnect()
      window.removeEventListener('resize', recalc)
      viewport.removeEventListener('pointermove', scheduleSyncZoomEngaged)
      viewport.removeEventListener('click', scheduleSyncZoomEngaged, true)
      setIsImageZoomEngaged(false)
    }
  }, [imageUrl, imageZoom])

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

  const visibleZones = useMemo(() => {
    if (!imageDisplayBox.width || !imageDisplayBox.height) return []
    return highlightableZones
      .map((zone) => {
        const useNaturalBounds =
          !zone.hasSurfaceBounds &&
          imageDisplayBox.naturalWidth > 0 &&
          imageDisplayBox.naturalHeight > 0
        const refUlx = useNaturalBounds ? 0 : zone.refUlx
        const refUly = useNaturalBounds ? 0 : zone.refUly
        const refLrx = useNaturalBounds
          ? imageDisplayBox.naturalWidth
          : zone.refLrx
        const refLry = useNaturalBounds
          ? imageDisplayBox.naturalHeight
          : zone.refLry
        const refWidth = refLrx - refUlx
        const refHeight = refLry - refUly
        return {
          ...zone,
          left:
            imageDisplayBox.left +
            ((zone.ulx - refUlx) / refWidth) * imageDisplayBox.width,
          top:
            imageDisplayBox.top +
            ((zone.uly - refUly) / refHeight) * imageDisplayBox.height,
          width: ((zone.lrx - zone.ulx) / refWidth) * imageDisplayBox.width,
          height: ((zone.lry - zone.uly) / refHeight) * imageDisplayBox.height,
        }
      })
      .filter(
        (z) =>
          z.width > 0 &&
          z.height > 0 &&
          z.left < imageDisplayBox.left + imageDisplayBox.width &&
          z.top < imageDisplayBox.top + imageDisplayBox.height &&
          z.left + z.width > imageDisplayBox.left &&
          z.top + z.height > imageDisplayBox.top,
      )
  }, [highlightableZones, imageDisplayBox])

  const activeMatchIdSet = useMemo(
    () => new Set(activeLineMatchIds),
    [activeLineMatchIds],
  )

  const showOverlay = highlightMode !== 'hide' && visibleZones.length > 0

  useEffect(() => {
    onHoverLineMatchIds([])
  }, [imageUrl, highlightMode, onHoverLineMatchIds, surfaceZones])

  const handleViewportPointerMove = (
    event: React.PointerEvent<HTMLDivElement>,
  ) => {
    if (highlightMode !== 'hover' || isImageZoomEngaged) {
      onHoverLineMatchIds([])
      return
    }
    const rect = event.currentTarget.getBoundingClientRect()
    const x = event.clientX - rect.left
    const y = event.clientY - rect.top
    let hovered: (typeof visibleZones)[number] | null = null
    for (let i = visibleZones.length - 1; i >= 0; i--) {
      const zone = visibleZones[i]
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
    onHoverLineMatchIds(
      hovered
        ? hovered.hoverMatchIds.length > 0
          ? hovered.hoverMatchIds
          : hovered.matchIds
        : [],
    )
  }

  return (
    <div
      className={`h-full overflow-hidden flex items-center justify-center ${half ? 'w-1/2' : 'w-full'}`}
    >
      <div
        ref={viewportRef}
        className="relative h-full w-full flex justify-center items-center"
        onPointerMove={handleViewportPointerMove}
        onPointerLeave={() => onHoverLineMatchIds([])}
      >
        <div className="relative z-0 h-full w-full flex items-center justify-center">
          <ImageZoom
            key={imageUrl}
            src={imageUrl}
            alt="Page image"
            zoom={String(imageZoom)}
            width="100%"
            height="100%"
            className="max-h-full max-w-full w-full h-full overflow-hidden [&>*]:mx-auto [&_img]:h-full [&_img]:w-full [&_img]:max-w-none [&_img]:max-h-none [&_img]:object-contain [&_img]:object-center"
          />
        </div>
        {showOverlay && (
          <div className="absolute inset-0 z-30 pointer-events-none">
            {visibleZones.map((zone) => {
              const isShown =
                highlightMode === 'show' ||
                (highlightMode === 'hover' &&
                  zone.matchIds.some((id) => activeMatchIdSet.has(id)))
              if (!isShown) return null
              return (
                <div
                  key={zone.id}
                  className={`absolute border-2 rounded-sm ${zone.zoneType === 'block' ? 'border-amber-400/60 bg-amber-200/10' : 'border-teal-400/60 bg-teal-200/10'}`}
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
  )
}

type GalleryCardBodyProps = {
  imageUrl: string | null
  imageZoom: number
  highlightMode: HighlightMode
  surfaceZones: TeiSurfaceZone[]
  sideBySide: boolean
  showText: boolean
  teiContent: string | undefined
  isFetchingTei: boolean
  minCert: number
  showCertaintyVisualization: boolean
  teiViewMode: TeiViewMode
  alignLines: boolean
  highlightConfig: TeiHighlightConfig | undefined
  visibleFeatureIdsKey: string
  showTeiLineHighlights: boolean
  searchResultHighlight: string | null
}

function GalleryCardBody({
  imageUrl,
  imageZoom,
  highlightMode,
  surfaceZones,
  sideBySide,
  showText,
  teiContent,
  isFetchingTei,
  minCert,
  showCertaintyVisualization,
  teiViewMode,
  alignLines,
  highlightConfig,
  visibleFeatureIdsKey,
  showTeiLineHighlights,
  searchResultHighlight,
}: GalleryCardBodyProps) {
  const [activeLineMatchIds, setActiveLineMatchIds] = useState<string[]>([])
  return (
    <>
      {imageUrl && (
        <GalleryImageCard
          imageUrl={imageUrl}
          imageZoom={imageZoom}
          highlightMode={highlightMode}
          surfaceZones={surfaceZones}
          half={sideBySide}
          activeLineMatchIds={activeLineMatchIds}
          onHoverLineMatchIds={setActiveLineMatchIds}
        />
      )}
      {showText && (
        <div
          className={`h-full overflow-auto ${sideBySide ? 'w-1/2 border-l border-gray-200' : 'w-full'}`}
        >
          {teiContent ? (
            <TeiContentView
              key={`${teiViewMode}:${visibleFeatureIdsKey}`}
              data={teiContent}
              minCert={minCert}
              searchResultHighlight={searchResultHighlight}
              showCertaintyVisualization={showCertaintyVisualization}
              viewMode={teiViewMode}
              viewLabel=""
              showViewLabel={false}
              alignLines={alignLines}
              centerRows={false}
              editable={false}
              canAddHighlight={false}
              noFrame
              highlightConfig={highlightConfig}
              activeLineMatchIds={
                showTeiLineHighlights ? activeLineMatchIds : []
              }
              onHoverLineMatchIds={setActiveLineMatchIds}
              onRequestAddHighlight={() => {}}
              onRequestRemoveHighlight={() => {}}
            />
          ) : isFetchingTei ? (
            <p className="text-xs text-gray-400 p-2">Loading…</p>
          ) : (
            <p className="text-xs text-gray-400 italic p-2">No text</p>
          )}
        </div>
      )}
    </>
  )
}

export function GalleryViewTab() {
  const {
    annotation,
    dataset,
    state: { annotationId, datasetId, currentPageOrKey },
    setState,
  } = useAppState()

  const hasPages = hasAnnotationPages(annotation)
  const shouldLoadImageKeys = !!annotation
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    datasetId,
    shouldLoadImageKeys,
    hasPages ? annotation!.pages!.split(',') : null,
  )

  const availablePages = useMemo(() => {
    if (!annotation) return []
    if (hasPages) {
      return [...new Set((annotation.pages || '').split(',').flatMap(expandRange))]
        .sort((a, b) => a.localeCompare(b, undefined, { numeric: true }))
    }
    return imageKeys.map((img) => img.key)
  }, [annotation, hasPages, imageKeys])
  const [searchTerm] = useLocalStorageState(ANNOTATION_SEARCH_TERM_KEY, {
    defaultValue: '',
    storageSync: false,
  })
  const [selectedCategories] = useLocalStorageState<string[] | null>(
    ANNOTATION_SEARCH_CATEGORIES_KEY,
    {
      defaultValue: null,
      storageSync: false,
    },
  )
  const [selectedSearchWithin] = useLocalStorageState<
    annotation_SearchWithin | annotation_SearchWithin[] | null
  >(ANNOTATION_SEARCH_WITHIN_KEY, {
    defaultValue: null,
    storageSync: false,
  })
  const { data: categories } = useAnnotationCategories(datasetId, annotationId)
  const normalizedSearch = searchTerm.trim()
  const hasCategories = (categories?.length ?? 0) > 0
  const activeSearchCategories = useMemo(() => {
    if (!categories || categories.length === 0) {
      return []
    }
    if (
      selectedCategories == null ||
      selectedCategories.length === categories.length
    ) {
      return categories
    }
    return selectedCategories.filter((category) =>
      categories.includes(category),
    )
  }, [categories, selectedCategories])
  const sortedSearchCategories = useMemo(
    () => [...activeSearchCategories].sort(),
    [activeSearchCategories],
  )
  const availableSearchWithinOptions = useMemo(() => {
    return annotationSearchWithinOptions.filter((option) => {
      if (option === 'categories') {
        return hasCategories
      }
      if (option === 'transcription') {
        return !hasCategories
      }
      if (option === 'biblio_metadata') {
        return !dataset?.edition_id
      }
      return true
    })
  }, [dataset?.edition_id, hasCategories])
  const normalizedSelectedSearchWithin = useMemo(() => {
    const currentSelection = Array.isArray(selectedSearchWithin)
      ? selectedSearchWithin
      : selectedSearchWithin
        ? [selectedSearchWithin]
        : []

    return currentSelection.find((option) =>
      availableSearchWithinOptions.includes(option),
    )
  }, [availableSearchWithinOptions, selectedSearchWithin])
  const activeSearchWithin = useMemo(
    () =>
      normalizedSelectedSearchWithin ? [normalizedSelectedSearchWithin] : [],
    [normalizedSelectedSearchWithin],
  )
  const gallerySearchQuery = useAnnotationSearch(
    datasetId,
    annotationId,
    normalizedSearch,
    sortedSearchCategories,
    activeSearchWithin,
  )
  const filteredPageSet = useMemo(() => {
    if (
      !normalizedSearch ||
      gallerySearchQuery.isLoading ||
      gallerySearchQuery.error
    ) {
      return null
    }
    return new Set(
      (gallerySearchQuery.data?.results ?? [])
        .map(getSearchResultPageOrKey)
        .filter((pageOrKey): pageOrKey is string => !!pageOrKey),
    )
  }, [
    gallerySearchQuery.data?.results,
    gallerySearchQuery.error,
    gallerySearchQuery.isLoading,
    normalizedSearch,
  ])
  const filteredAvailablePages = useMemo(() => {
    if (!filteredPageSet) {
      return availablePages
    }
    return availablePages.filter((page) => filteredPageSet.has(page))
  }, [availablePages, filteredPageSet])
  const searchHighlightsByPage = useMemo(() => {
    const highlights = new Map<string, string[]>()

    for (const result of gallerySearchQuery.data?.results ?? []) {
      const pageOrKey = getSearchResultPageOrKey(result)
      if (!pageOrKey || !result.content) {
        continue
      }
      const existing = highlights.get(pageOrKey) ?? []
      existing.push(result.content)
      highlights.set(pageOrKey, existing)
    }

    return new Map(
      Array.from(highlights.entries()).map(([pageOrKey, contents]) => [
        pageOrKey,
        contents.join(' '),
      ]),
    )
  }, [gallerySearchQuery.data?.results])
  const galleryRenderKey = useMemo(
    () =>
      [
        normalizedSearch,
        sortedSearchCategories.join('|'),
        activeSearchWithin.join('|'),
      ].join('::'),
    [activeSearchWithin, normalizedSearch, sortedSearchCategories],
  )

  const [viewMode, setViewMode] = useLocalStorageState<GalleryViewMode>(
    'galleryViewMode',
    { defaultValue: 'images', storageSync: false },
  )
  const [cardSize, setCardSize] = useLocalStorageState('galleryCardSize', {
    defaultValue: 280,
    storageSync: false,
  })

  const [imageZoom, setImageZoom] = useLocalStorageState('galleryImageZoom', {
    defaultValue: DEFAULT_IMAGE_ZOOM,
    storageSync: false,
  })

  const [highlightMode, setHighlightMode] = useLocalStorageState<HighlightMode>(
    'imagePaneHighlightMode',
    {
      defaultValue: 'hover',
      storageSync: false,
    },
  )

  const [showTeiLineHighlights, setShowTeiLineHighlights] =
    useLocalStorageState('showTeiLineHighlights', { defaultValue: true })
  const [alignLines, setAlignLines] = useLocalStorageState('alignTeiLines', {
    defaultValue: false,
  })
  const [minCert, setMinCert] = useLocalStorageState('minCert', {
    defaultValue: 0.8,
  })
  const [showCertaintyVisualization, setShowCertaintyVisualization] =
    useLocalStorageState('showTeiCertaintyVisualization', {
      defaultValue: false,
    })
  const [teiViewMode, setTeiViewMode] = useLocalStorageState<TeiViewMode>(
    'galleryTeiViewMode',
    { defaultValue: 'original', storageSync: false },
  )
  const [isFeatureSelectExpanded, setIsFeatureSelectExpanded] =
    useLocalStorageState('teiFeatureSelectExpanded', { defaultValue: false })

  const highlightStorageKey = datasetId
    ? `teiVisibleHighlightFeatures:${datasetId}`
    : 'teiVisibleHighlightFeatures'
  const [storedVisibleFeatureIds, setStoredVisibleFeatureIds] =
    useLocalStorageState<string[] | null>(highlightStorageKey, {
      defaultValue: null,
    })

  const showText = viewMode === 'texts' || viewMode === 'side-by-side'
  const showImage = viewMode === 'images' || viewMode === 'side-by-side'
  const ocred = !!annotation?.ocred

  const [visibleCount, setVisibleCount] = useState(BATCH_SIZE)
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    setVisibleCount(BATCH_SIZE)
  }, [datasetId, annotationId, filteredAvailablePages.length, normalizedSearch])

  const visiblePages = useMemo(
    () => filteredAvailablePages.slice(0, visibleCount),
    [filteredAvailablePages, visibleCount],
  )
  const hasMore = visibleCount < filteredAvailablePages.length
  const currentPageKey = String(currentPageOrKey)
  const currentPageIndex = useMemo(
    () => filteredAvailablePages.indexOf(currentPageKey),
    [filteredAvailablePages, currentPageKey],
  )

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel || !hasMore) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          setVisibleCount((prev) =>
            Math.min(prev + BATCH_SIZE, filteredAvailablePages.length),
          )
        }
      },
      { threshold: 0.1 },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [filteredAvailablePages.length, hasMore])

  useEffect(() => {
    if (currentPageIndex < 0 || currentPageIndex < visibleCount) return
    setVisibleCount((prev) =>
      Math.min(
        filteredAvailablePages.length,
        Math.max(prev + BATCH_SIZE, currentPageIndex + 1),
      ),
    )
  }, [currentPageIndex, filteredAvailablePages.length, visibleCount])

  const fetchedPageSetRef = useRef(new Set<string>())
  const pagesToFetchRef = useRef<string[]>([])
  const [teiByPage, setTeiByPage] = useState(new Map<string, string>())

  useEffect(() => {
    fetchedPageSetRef.current = new Set()
    setTeiByPage(new Map())
  }, [datasetId, annotationId])

  const pagesToFetch = useMemo(() => {
    if (!ocred) return []
    return visiblePages.filter((p) => !fetchedPageSetRef.current.has(p))
  }, [ocred, visiblePages])
  const pendingTeiPageSet = useMemo(() => new Set(pagesToFetch), [pagesToFetch])

  pagesToFetchRef.current = pagesToFetch

  const teisQuery = useAnnotationTeisQuery(
    datasetId,
    annotationId,
    pagesToFetch,
    pagesToFetch.length > 0,
  )

  useEffect(() => {
    if (!teisQuery.data) return
    const pages = pagesToFetchRef.current
    if (pages.length === 0) return
    const parsed = parseTeisXml(teisQuery.data)
    pages.forEach((p) => fetchedPageSetRef.current.add(p))
    setTeiByPage((prev) => {
      const next = new Map(prev)
      parsed.forEach((v, k) => next.set(k, v))
      return next
    })
  }, [teisQuery.data])

  const firstTei = useMemo(
    () => (teiByPage.size > 0 ? [...teiByPage.values()][0] : null),
    [teiByPage],
  )
  const teiTranslations = useMemo(
    () => (firstTei ? getTeiTranslations(firstTei) : []),
    [firstTei],
  )
  const availableTeiViewModes = useMemo<TeiViewMode[]>(
    () => ['original', ...teiTranslations.map((t) => t.id)],
    [teiTranslations],
  )
  const showMinCertControl = useMemo(
    () => (firstTei ? hasTeiCertaintyDegrees(firstTei) : false),
    [firstTei],
  )
  const effectiveTeiViewMode: TeiViewMode = availableTeiViewModes.includes(
    teiViewMode,
  )
    ? teiViewMode
    : 'original'

  const getViewModeLabel = (mode: TeiViewMode): string => {
    if (mode === 'original') return 'Original'
    const rawLabel = teiTranslations.find((t) => t.id === mode)?.label || mode
    return VIEW_LABEL_MAP[rawLabel] || rawLabel
  }

  const teiViewModeOptions = useMemo<TeiViewModeOption[]>(
    () =>
      availableTeiViewModes.map((mode) => ({
        value: mode,
        label: getViewModeLabel(mode),
      })),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [availableTeiViewModes],
  )

  const featuresQuery = useDatasetFeaturesQuery(datasetId, !!annotationId)
  const datasetFeatures = useMemo(
    () => featuresQuery.data ?? [],
    [featuresQuery.data],
  )
  const teiCategories = useMemo(() => {
    const byId = new Map<string, string>()
    teiByPage.forEach((tei) => {
      for (const category of getTeiHighlightCategories(tei)) {
        if (!byId.has(category.id)) {
          byId.set(category.id, category.label)
        }
      }
    })
    return [...byId.entries()].map(([id, label]) => ({ id, label }))
  }, [teiByPage])
  const resolvedTeiFeatures = useMemo<ResolvedTeiFeature[]>(() => {
    const byId = new Map<string, ResolvedTeiFeature>()
    for (const category of teiCategories) {
      const matched = matchTeiCategoryToFeature(
        category.id,
        category.label,
        datasetFeatures,
      )
      const featureId = matched?.id || category.id
      if (byId.has(featureId)) continue
      byId.set(featureId, {
        id: featureId,
        label: matched?.name?.trim() || category.label,
        description: matched?.description?.trim() || '',
        color: matched?.color || '#f2f2f2',
        isDefault: !!matched?.is_default,
        renderMode: isVerbLike(
          matched?.id,
          matched?.name,
          category.id,
          category.label,
        )
          ? 'outline'
          : 'fill',
      })
    }
    return [...byId.values()]
  }, [datasetFeatures, teiCategories])

  const allResolvedFeatures = useMemo<ResolvedTeiFeature[]>(() => {
    const byId = new Map<string, ResolvedTeiFeature>()
    for (const feature of datasetFeatures) {
      byId.set(feature.id!, {
        id: feature.id!,
        label: feature.name?.trim() || feature.id!,
        description: feature.description?.trim() || '',
        color: feature.color || '#f2f2f2',
        isDefault: !!feature.is_default,
        renderMode: isVerbLike(feature.id, feature.name) ? 'outline' : 'fill',
      })
    }
    for (const feature of resolvedTeiFeatures) {
      if (!byId.has(feature.id)) byId.set(feature.id, feature)
    }
    return [...byId.values()].sort((a, b) =>
      a.label.localeCompare(b.label, undefined, { sensitivity: 'base' }),
    )
  }, [datasetFeatures, resolvedTeiFeatures])

  const categoryToFeatureId = useMemo(() => {
    const next: Record<string, string> = {}
    for (const category of teiCategories) {
      const matched = matchTeiCategoryToFeature(
        category.id,
        category.label,
        datasetFeatures,
      )
      next[category.id] = matched?.id || category.id
    }
    return next
  }, [datasetFeatures, teiCategories])

  const visibleFeatureIds = useMemo(() => {
    const availableIds = allResolvedFeatures.map((f) => f.id)
    if (!availableIds.length) return []
    const availableSet = new Set(availableIds)
    const defaultIds = allResolvedFeatures
      .filter((f) => f.isDefault)
      .map((f) => f.id)
    const order = (ids: string[]) =>
      availableIds.filter((id) => ids.includes(id))
    if (storedVisibleFeatureIds === null) return order(defaultIds)
    const filtered = order(
      storedVisibleFeatureIds.filter((id) => availableSet.has(id)),
    )
    if (storedVisibleFeatureIds.length === 0) return []
    return filtered
  }, [allResolvedFeatures, storedVisibleFeatureIds])

  const highlightConfig = useMemo<TeiHighlightConfig | undefined>(() => {
    if (!allResolvedFeatures.length) return undefined
    const categoryConfigById = allResolvedFeatures.reduce<
      Record<
        string,
        {
          label: string
          color: string
          description: string
          renderMode: 'fill' | 'outline'
        }
      >
    >((acc, feature) => {
      acc[feature.id] = {
        label: feature.label,
        color: feature.color,
        description: feature.description,
        renderMode: feature.renderMode,
      }
      return acc
    }, {})
    return {
      selectedCategoryIds: [...visibleFeatureIds],
      categoryConfigById,
      categoryToFeatureId,
    }
  }, [allResolvedFeatures, categoryToFeatureId, visibleFeatureIds])

  const allFeatureOptions = useMemo(
    () => toFeatureOptions(allResolvedFeatures),
    [allResolvedFeatures],
  )
  const selectedFeatureOptions = useMemo(
    () => allFeatureOptions.filter((o) => visibleFeatureIds.includes(o.value)),
    [allFeatureOptions, visibleFeatureIds],
  )
  const visibleFeatureIdsKey = visibleFeatureIds.join('|')

  const surfaceZonesByPage = useMemo(() => {
    const result = new Map<string, TeiSurfaceZone[]>()
    teiByPage.forEach((tei, page) => {
      result.set(page, getTeiSurfaceZones(tei))
    })
    return result
  }, [teiByPage])
  const hasSurfaceZones = surfaceZonesByPage.size > 0

  const cardRefs = useRef<Record<string, HTMLDivElement | null>>({})
  const lastScrolledPageRef = useRef('')

  const handleOpenInPageView = (pageKey: string) => {
    setState({
      annotationTab: 'text',
      currentPageOrKey: String(pageKey),
    })
  }

  useEffect(() => {
    const key = currentPageKey
    const el = cardRefs.current[key]
    if (!el) return
    if (key === lastScrolledPageRef.current) return
    lastScrolledPageRef.current = key
    el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }, [currentPageKey, visiblePages])

  const getImageUrl = (pageOrKey: string): string | null => {
    const num = Number(pageOrKey)
    let normalizedKey: string
    if (!Number.isNaN(num) && Number.isInteger(num)) {
      normalizedKey = `page-${String(num).padStart(4, '0')}.png`
    } else {
      const matched = imageKeys.find((img) => img.key === pageOrKey)
      normalizedKey = matched?.filename || ''
    }
    if (!normalizedKey) return null
    return `${import.meta.env.VITE_BACKEND_URL}/store/data/${datasetId}/imgs/${normalizedKey}`
  }

  const cardWidth = viewMode === 'side-by-side' ? cardSize * 2 : cardSize
  const selectedViewModeOption =
    VIEW_MODE_OPTIONS.find((o) => o.value === viewMode) ?? VIEW_MODE_OPTIONS[0]
  const selectedTeiViewModeOption =
    teiViewModeOptions.find((o) => o.value === effectiveTeiViewMode) ??
    teiViewModeOptions[0]
  const selectedHighlightModeOption =
    HIGHLIGHT_MODE_OPTIONS.find((o) => o.value === highlightMode) ??
    HIGHLIGHT_MODE_OPTIONS[1]

  return (
    <div className="h-full flex overflow-hidden">
      <AnnotationNavigation />
      <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
        <div className="border-b border-gray-200 bg-gray-50 shrink-0 flex flex-col">
          <div className="px-3 py-2 flex items-center gap-3 flex-wrap">
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-gray-600">View</span>
              <div className="w-36">
                <Select<ViewModeOption, false>
                  value={selectedViewModeOption}
                  onChange={(option) => option && setViewMode(option.value)}
                  options={VIEW_MODE_OPTIONS}
                  isClearable={false}
                  styles={selectStyles<ViewModeOption>()}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                />
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-gray-600">Zoom</span>
              <button
                className="w-6 h-6 flex items-center justify-center border border-gray-300 rounded bg-white text-sm leading-none disabled:opacity-40 disabled:cursor-not-allowed enabled:hover:bg-gray-100"
                onClick={() => setCardSize(Math.max(120, cardSize - 40))}
                title="Smaller cards"
                disabled={cardSize <= 120}
              >
                −
              </button>
              <button
                className="w-6 h-6 flex items-center justify-center border border-gray-300 rounded bg-white text-sm leading-none disabled:opacity-40 disabled:cursor-not-allowed enabled:hover:bg-gray-100"
                onClick={() => setCardSize(Math.min(800, cardSize + 40))}
                title="Larger cards"
                disabled={cardSize >= 800}
              >
                +
              </button>
            </div>
          </div>
          {showImage && (
            <div className="px-3 py-2 flex items-center gap-3 flex-wrap border-t border-gray-200">
              <RangeInput
                label="Image zoom"
                value={imageZoom}
                min={105}
                max={1000}
                step={5}
                onChange={(value) => setImageZoom(Math.round(value))}
                className="bg-transparent border-gray-300"
              />
              <button
                className="h-7 px-2 flex items-center justify-center border border-gray-300 rounded bg-white text-xs font-medium disabled:opacity-40 disabled:cursor-not-allowed enabled:hover:bg-gray-100"
                onClick={() => setImageZoom(DEFAULT_IMAGE_ZOOM)}
                disabled={imageZoom === DEFAULT_IMAGE_ZOOM}
                title="Reset image zoom"
              >
                Reset zoom
              </button>
              {hasSurfaceZones && (
                <div className="flex items-center gap-1.5 text-xs font-medium text-gray-700">
                  <span>Highlights</span>
                  <div className="w-28">
                    <Select<HighlightModeOption, false>
                      value={selectedHighlightModeOption}
                      onChange={(option) =>
                        setHighlightMode(option?.value || 'hover')
                      }
                      options={HIGHLIGHT_MODE_OPTIONS}
                      isClearable={false}
                      styles={selectStyles<HighlightModeOption>()}
                      menuPortalTarget={document.body}
                      menuPosition="fixed"
                    />
                  </div>
                </div>
              )}
            </div>
          )}
          {showText && (
            <div className="px-3 py-2 flex items-center gap-3 flex-wrap border-t border-gray-200">
              {availableTeiViewModes.length > 1 && (
                <div className="flex items-center gap-1.5">
                  <span className="text-xs font-medium text-gray-600">
                    Text view
                  </span>
                  <div className="w-36">
                    <Select<TeiViewModeOption, false>
                      value={selectedTeiViewModeOption}
                      onChange={(option) =>
                        option && setTeiViewMode(option.value)
                      }
                      options={teiViewModeOptions}
                      isClearable={false}
                      styles={selectStyles<TeiViewModeOption>()}
                      menuPortalTarget={document.body}
                      menuPosition="fixed"
                    />
                  </div>
                </div>
              )}
              <TeiDisplayControls
                showMinCertControl={showMinCertControl}
                minCert={minCert}
                setMinCert={setMinCert}
                showTeiLineHighlights={showTeiLineHighlights}
                setShowTeiLineHighlights={setShowTeiLineHighlights}
                alignLines={alignLines}
                setAlignLines={setAlignLines}
                showCertaintyVisualization={showCertaintyVisualization}
                setShowCertaintyVisualization={setShowCertaintyVisualization}
                allFeatureOptions={allFeatureOptions}
                selectedFeatureOptions={selectedFeatureOptions}
                isFeatureSelectExpanded={isFeatureSelectExpanded}
                setIsFeatureSelectExpanded={setIsFeatureSelectExpanded}
                setVisibleFeatureIds={setStoredVisibleFeatureIds}
                onResetVisibleFeatureIds={() =>
                  setStoredVisibleFeatureIds(null)
                }
                isFeaturesLoading={featuresQuery.isLoading}
              />
              {teisQuery.isFetching && (
                <span className="text-xs text-gray-400">Loading texts…</span>
              )}
            </div>
          )}
        </div>

        <div className="flex-1 min-h-0 overflow-auto p-3">
          {normalizedSearch && gallerySearchQuery.isLoading && (
            <div className="text-xs text-gray-400 mb-3">Filtering gallery…</div>
          )}
          {normalizedSearch &&
            !gallerySearchQuery.isLoading &&
            !gallerySearchQuery.error && (
              <div className="text-xs text-gray-500 mb-3">
                Showing {filteredAvailablePages.length} matching out of{' '}
                {availablePages.length} total items
              </div>
            )}
          <div className="flex flex-wrap gap-3">
            {visiblePages.map((page) => (
              <div
                key={page}
                ref={(el) => {
                  cardRefs.current[page] = el
                }}
                className="border border-gray-300 rounded-xl overflow-hidden flex flex-col bg-white shrink-0"
                style={{ width: cardWidth, height: cardSize }}
              >
                <div className="px-2 py-1 border-b border-gray-200 bg-gray-50 flex items-center justify-between shrink-0">
                  <span className="text-xs font-semibold">{page}</span>
                  <button
                    className="w-5 h-5 flex items-center justify-center border border-gray-300 rounded bg-white hover:bg-gray-100 text-xs leading-none"
                    title="Open in Page View"
                    onClick={() => handleOpenInPageView(page)}
                  >
                    ⤢
                  </button>
                </div>
                <div className="flex-1 min-h-0 flex overflow-hidden">
                  <GalleryCardBody
                    key={`${page}:${galleryRenderKey}`}
                    imageUrl={showImage ? getImageUrl(page) : null}
                    imageZoom={imageZoom}
                    highlightMode={highlightMode}
                    surfaceZones={surfaceZonesByPage.get(page) ?? []}
                    sideBySide={viewMode === 'side-by-side'}
                    showText={showText}
                    teiContent={showText ? teiByPage.get(page) : undefined}
                    isFetchingTei={pendingTeiPageSet.has(page)}
                    minCert={minCert}
                    showCertaintyVisualization={showCertaintyVisualization}
                    teiViewMode={effectiveTeiViewMode}
                    alignLines={alignLines}
                    highlightConfig={highlightConfig}
                    visibleFeatureIdsKey={visibleFeatureIdsKey}
                    showTeiLineHighlights={showTeiLineHighlights}
                    searchResultHighlight={
                      searchHighlightsByPage.get(page) ?? null
                    }
                  />
                </div>
              </div>
            ))}
          </div>
          {normalizedSearch &&
            !gallerySearchQuery.isLoading &&
            !gallerySearchQuery.error &&
            filteredAvailablePages.length === 0 && (
              <div className="text-xs text-gray-400 italic">
                No gallery items match the current search.
              </div>
            )}
          {hasMore && <div ref={sentinelRef} className="h-8 w-full mt-1" />}
        </div>
      </div>
    </div>
  )
}
