import React, { useEffect, useMemo, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import Select from 'react-select'
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
import { parsePageEntries } from '../../../utils/pages.ts'
import { selectStyles } from '../../../styles/selectStyles.ts'
import { RangeInput } from '../../core/RangeInput.tsx'
import { DEFAULT_IMAGE_ZOOM } from '../imageZoom.ts'
import { MultiSelectDropdown } from '../../core/MultiSelectDropdown.tsx'
import type { annotation_SearchWithin } from '@hub-api'
import {
  ANNOTATION_SEARCH_CATEGORIES_KEY,
  ANNOTATION_SEARCH_TERM_KEY,
  ANNOTATION_SEARCH_WITHIN_KEY,
  getSearchResultPageOrKey,
} from '../contents/navigation/annotationSearchUtils.ts'
import { findMatchingImage } from '../../../utils/editions.ts'
import {
  buildDatasetImageUrl,
  type DatasetImageVariant,
} from '../../../utils/imageUrls.ts'
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

type GalleryViewMode = 'images' | 'texts' | 'side-by-side'
type ViewModeOption = { value: GalleryViewMode; label: string }
type TeiViewModeOption = { value: TeiViewMode; label: string }
type HighlightMode = 'hide' | 'hover' | 'show'
type HighlightModeOption = { value: HighlightMode; label: string }

const BATCH_SIZE = 20
const DEFAULT_CARD_SIZE = 280
const PREVIEW_VARIANT_CARD_SIZE_THRESHOLD = 400
const PREVIEW_VARIANT_ZOOM_THRESHOLD = 300
const ORIGINAL_VARIANT_ZOOM_THRESHOLD = 600

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

type GalleryImageCardProps = {
  baseVariant: DatasetImageVariant
  getVariantUrl: (variant: DatasetImageVariant) => string
  imageZoom: number
  highlightMode: HighlightMode
  highlightZoneFilters: HighlightZoneFilter[]
  surfaceZones: TeiSurfaceZone[]
  half: boolean
  activeLineMatchIds: string[]
  onHoverLineMatchIds: (ids: string[]) => void
}

function GalleryImageCard({
  baseVariant,
  getVariantUrl,
  imageZoom,
  highlightMode,
  highlightZoneFilters,
  surfaceZones,
  half,
  activeLineMatchIds,
  onHoverLineMatchIds,
}: GalleryImageCardProps) {
  const zoomKey = getVariantUrl(baseVariant)
  const [zoomedVariant, setZoomedVariant] =
    useState<DatasetImageVariant | null>(null)
  const [baseNaturalSize, setBaseNaturalSize] = useState<{
    width: number
    height: number
  } | null>(null)
  const [prevZoomKey, setPrevZoomKey] = useState(zoomKey)

  if (zoomKey !== prevZoomKey) {
    setPrevZoomKey(zoomKey)
    setZoomedVariant(null)
    setBaseNaturalSize(null)
  }

  const effectiveVariant = zoomedVariant ?? baseVariant
  const imageUrl = getVariantUrl(effectiveVariant)

  const {
    containerRef,
    imgRef,
    isZoomed,
    zoomTransform,
    imageDisplayBox,
    handleContainerClick: hookHandleContainerClick,
    updateCursor,
    getLocalCursor,
  } = useZoomableImage(imageUrl, imageZoom, zoomKey)

  const zoneDisplayBox = useMemo(() => {
    if (!zoomedVariant || !baseNaturalSize || imageDisplayBox.width === 0) {
      return imageDisplayBox
    }
    return {
      ...imageDisplayBox,
      naturalWidth: baseNaturalSize.width,
      naturalHeight: baseNaturalSize.height,
    }
  }, [imageDisplayBox, zoomedVariant, baseNaturalSize])

  const handleContainerClick = (event: React.MouseEvent<HTMLDivElement>) => {
    if (!isZoomed) {
      setBaseNaturalSize({
        width: imageDisplayBox.naturalWidth,
        height: imageDisplayBox.naturalHeight,
      })
      const upgraded: DatasetImageVariant =
        imageZoom >= ORIGINAL_VARIANT_ZOOM_THRESHOLD
          ? 'original'
          : imageZoom >= PREVIEW_VARIANT_ZOOM_THRESHOLD
            ? 'preview'
            : baseVariant
      setZoomedVariant(upgraded)
    } else {
      setZoomedVariant(null)
      setBaseNaturalSize(null)
    }
    hookHandleContainerClick(event)
  }

  const activeMatchIdSet = useMemo(
    () => new Set(activeLineMatchIds),
    [activeLineMatchIds],
  )
  const highlightableZones = useMemo(
    () => filterValidZones(surfaceZones, highlightZoneFilters),
    [highlightZoneFilters, surfaceZones],
  )
  const visibleZones = useMemo(
    () =>
      computeVisibleZones(highlightableZones, zoneDisplayBox, activeMatchIdSet),
    [activeMatchIdSet, highlightableZones, zoneDisplayBox],
  )

  const showOverlay = highlightMode !== 'hide' && visibleZones.length > 0

  useEffect(() => {
    onHoverLineMatchIds([])
  }, [imageUrl, highlightMode, onHoverLineMatchIds, surfaceZones])

  const handleViewportPointerMove = (
    event: React.PointerEvent<HTMLDivElement>,
  ) => {
    const rect = event.currentTarget.getBoundingClientRect()
    const cx = event.clientX - rect.left
    const cy = event.clientY - rect.top

    updateCursor(cx, cy)

    if (highlightMode !== 'hover') {
      onHoverLineMatchIds([])
      return
    }

    const { localX, localY } = getLocalCursor(cx, cy)
    let hovered: (typeof visibleZones)[number] | null = null
    for (let i = visibleZones.length - 1; i >= 0; i--) {
      const zone = visibleZones[i]
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
    onHoverLineMatchIds(
      hovered
        ? hovered.hoverMatchIds.length > 0
          ? hovered.hoverMatchIds
          : hovered.matchIds
        : [],
    )
  }

  return (
    <div className={`h-full overflow-hidden ${half ? 'w-1/2' : 'w-full'}`}>
      <div
        ref={containerRef}
        className={`relative h-full w-full overflow-hidden select-none ${isZoomed ? 'cursor-zoom-out' : 'cursor-zoom-in'}`}
        onClick={handleContainerClick}
        onPointerMove={handleViewportPointerMove}
        onPointerLeave={() => onHoverLineMatchIds([])}
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
          {showOverlay && (
            <div className="absolute inset-0 z-30 pointer-events-none">
              {visibleZones.map((zone) => {
                const isShown = highlightMode === 'show' || zone.isActive
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
    </div>
  )
}

type GalleryCardBodyProps = {
  baseVariant: DatasetImageVariant
  getVariantUrl: ((variant: DatasetImageVariant) => string) | null
  imageZoom: number
  highlightMode: HighlightMode
  highlightZoneFilters: HighlightZoneFilter[]
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
  baseVariant,
  getVariantUrl,
  imageZoom,
  highlightMode,
  highlightZoneFilters,
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
      {getVariantUrl && (
        <GalleryImageCard
          baseVariant={baseVariant}
          getVariantUrl={getVariantUrl}
          imageZoom={imageZoom}
          highlightMode={highlightMode}
          highlightZoneFilters={highlightZoneFilters}
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

  const annotationPageEntries = useMemo(
    () => (annotation ? parsePageEntries(annotation.pages || '') : []),
    [annotation],
  )
  const shouldLoadImageKeys = !!annotation
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    datasetId,
    shouldLoadImageKeys,
    annotationPageEntries.length > 0 ? annotationPageEntries : null,
  )

  const availablePages = useMemo(() => {
    if (!annotation) return []
    if (annotationPageEntries.length > 0) {
      return [...new Set(annotationPageEntries)].sort((a, b) =>
        a.localeCompare(b, undefined, { numeric: true }),
      )
    }
    return imageKeys.map((img) => img.key)
  }, [annotation, annotationPageEntries, imageKeys])
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
    defaultValue: DEFAULT_CARD_SIZE,
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
  const [highlightZoneFilters, setHighlightZoneFilters] = useLocalStorageState<
    HighlightZoneFilter[]
  >(HIGHLIGHT_ZONE_FILTER_STORAGE_KEY, {
    defaultValue: DEFAULT_HIGHLIGHT_ZONE_FILTERS,
    storageSync: false,
  })

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
  const segmented = !!annotation?.segmented

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

  const imageVariant =
    cardSize >= PREVIEW_VARIANT_CARD_SIZE_THRESHOLD ? 'preview' : 'thumb'

  const fetchedPageSetRef = useRef(new Set<string>())
  const pagesToFetchRef = useRef<string[]>([])
  const [teiByPage, setTeiByPage] = useState(new Map<string, string>())

  useEffect(() => {
    fetchedPageSetRef.current = new Set()
    setTeiByPage(new Map())
  }, [datasetId, annotationId, imageVariant])

  const pagesToFetch = useMemo(() => {
    if (!segmented) {
      return []
    }
    return visiblePages.filter((p) => !fetchedPageSetRef.current.has(p))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(visiblePages), segmented])
  const pendingTeiPageSet = useMemo(() => new Set(pagesToFetch), [pagesToFetch])

  pagesToFetchRef.current = pagesToFetch

  const teisQuery = useAnnotationTeisQuery(
    datasetId,
    annotationId,
    pagesToFetch,
    pagesToFetch.length > 0,
    imageVariant,
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

  const makeGetVariantUrl = (pageOrKey: string) => {
    const num = Number(pageOrKey)
    let normalizedKey: string
    if (!Number.isNaN(num) && Number.isInteger(num)) {
      normalizedKey = `page-${String(num).padStart(4, '0')}.png`
    } else {
      const matched = findMatchingImage(pageOrKey, imageKeys)
      normalizedKey = matched?.filename || ''
    }
    if (!normalizedKey) return null
    return (variant: DatasetImageVariant) =>
      buildDatasetImageUrl(datasetId, normalizedKey, variant)
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
              <button
                className="h-6 px-2 flex items-center justify-center border border-gray-300 rounded bg-white text-xs font-medium disabled:opacity-40 disabled:cursor-not-allowed enabled:hover:bg-gray-100"
                onClick={() => setCardSize(DEFAULT_CARD_SIZE)}
                disabled={cardSize === DEFAULT_CARD_SIZE}
                title="Reset card size"
              >
                Reset
              </button>
            </div>
          </div>
          {showImage && (
            <div className="px-3 py-2 flex items-center gap-3 flex-wrap border-t border-gray-200">
              {hasSurfaceZones && (
                <div className="flex items-center gap-3 text-xs font-medium text-gray-700 flex-wrap">
                  <div className="flex items-center gap-1.5">
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
                  <div className="flex items-center gap-1.5">
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
              <RangeInput
                label="Image zoom"
                value={imageZoom}
                min={105}
                max={1000}
                step={5}
                onChange={(value) => setImageZoom(Math.round(value))}
                className="bg-transparent border-gray-300"
              />
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
                    baseVariant={imageVariant}
                    getVariantUrl={showImage ? makeGetVariantUrl(page) : null}
                    imageZoom={imageZoom}
                    highlightMode={highlightMode}
                    highlightZoneFilters={highlightZoneFilters}
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
