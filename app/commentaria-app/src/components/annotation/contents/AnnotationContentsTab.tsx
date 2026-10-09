import { ImagePane } from './ImagePane.tsx'
import { TeiPane } from './tei/TeiPane.tsx'
import type { TeiSurfaceZone } from './tei/tei.ts'
import { AnnotationNavigation } from './AnnotationNavigation.tsx'
import { useAppState } from '../../../context/useAppState.ts'
import { useEffect, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { useScopedStorageKey } from '../../../context/storageScope.ts'
import { HeightResizable } from '../../core/HeightResizable.tsx'

const MIN_STACKED_CONTENTS_WIDTH = 360
const DEFAULT_COMPARE_PANE_HEIGHT = 800
const COMPARE_CONTENT_CHROME_WIDTH = 36

const normalizeMatchIds = (ids: string[]) =>
  [...new Set(ids.map((id) => id.trim()).filter(Boolean))].sort()

interface AnnotationContentsTabProps {
  showImage?: boolean
  showContents?: boolean
  compareLayout?: boolean
}

export function AnnotationContentsTab({
  showImage = true,
  showContents = true,
  compareLayout = false,
}: AnnotationContentsTabProps) {
  const {
    state: { annotationId, currentPageOrKey, datasetId },
  } = useAppState()
  const [storedImagePaneWidth, setImagePaneWidth] = useLocalStorageState<
    number | null
  >(useScopedStorageKey('imagePaneWidth'), {
    defaultValue: compareLayout ? null : 560,
    storageSync: false,
  })
  const [isResizingImagePane, setIsResizingImagePane] = useState(false)
  const [activeLineMatchIds, setActiveLineMatchIds] = useState<string[]>([])
  const [surfaceZones, setSurfaceZones] = useState<TeiSurfaceZone[]>([])
  const [allZoneCategories, setAllZoneCategories] = useState<string[]>([])
  const contentRef = useRef<HTMLDivElement | null>(null)
  const [contentWidth, setContentWidth] = useState(0)
  const [navigationFloating, setNavigationFloating] = useLocalStorageState(
    useScopedStorageKey('navigationFloating'),
    { defaultValue: compareLayout, storageSync: false },
  )
  const [navigationFloatingOpen, setNavigationFloatingOpen] = useState(false)
  const imagePaneHeightKey = useScopedStorageKey('imagePaneHeight')
  const teiPaneHeightKey = useScopedStorageKey('teiPaneHeight')

  const handleHoverLineMatchIds = (ids: string[]) => {
    const normalized = normalizeMatchIds(ids)
    setActiveLineMatchIds((previous) =>
      normalizeMatchIds(previous).join('|') === normalized.join('|')
        ? previous
        : normalized,
    )
  }

  useEffect(() => {
    const clampWidth = (value: number) => {
      const container = contentRef.current
      if (!container) return value
      const maxWidth = Math.round(container.clientWidth * 0.75)
      return Math.min(maxWidth, Math.max(320, value))
    }

    if (isResizingImagePane) {
      const onPointerMove = (event: PointerEvent) => {
        const container = contentRef.current
        if (!container) return
        const rect = container.getBoundingClientRect()
        const raw = event.clientX - rect.left
        setImagePaneWidth(clampWidth(raw))
      }

      const onPointerUp = () => {
        setIsResizingImagePane(false)
      }

      window.addEventListener('pointermove', onPointerMove)
      window.addEventListener('pointerup', onPointerUp)
      return () => {
        window.removeEventListener('pointermove', onPointerMove)
        window.removeEventListener('pointerup', onPointerUp)
      }
    }

    const onResize = () => {
      setImagePaneWidth((current) =>
        current == null ? current : clampWidth(current),
      )
    }

    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
    }
  }, [isResizingImagePane, setImagePaneWidth])

  useEffect(() => {
    const container = contentRef.current
    if (!compareLayout || !container) return
    const observer = new ResizeObserver(() => {
      setContentWidth(container.clientWidth)
    })
    observer.observe(container)
    return () => observer.disconnect()
  }, [compareLayout])

  const imagePaneWidth =
    storedImagePaneWidth ??
    (contentWidth > 0
      ? Math.round((contentWidth - COMPARE_CONTENT_CHROME_WIDTH) / 2)
      : 560)
  const isStacked =
    compareLayout &&
    showImage &&
    showContents &&
    contentWidth > 0 &&
    contentWidth < imagePaneWidth + MIN_STACKED_CONTENTS_WIDTH
  const hasFixedImageWidth = showContents && !isStacked
  const imagePaneStyle = hasFixedImageWidth
    ? {
        width: `${imagePaneWidth}px`,
        minWidth: `${imagePaneWidth}px`,
        maxWidth: `${imagePaneWidth}px`,
      }
    : undefined

  if (compareLayout) {
    return (
      <div
        className="h-full overflow-y-auto overflow-x-hidden"
        style={{ scrollbarGutter: 'stable' }}
      >
        <div
          className={`flex ${showImage || showContents ? '' : 'min-h-full'}`}
        >
          {!navigationFloating && (
            <div className="flex shrink-0 pb-2 [&>aside]:[contain:size]">
              <AnnotationNavigation
                floating={false}
                onFloatingChange={setNavigationFloating}
                floatingOpen={navigationFloatingOpen}
                onFloatingOpenChange={setNavigationFloatingOpen}
              />
            </div>
          )}
          <div
            ref={contentRef}
            className={`relative flex-1 min-w-0 flex gap-3 px-3 box-border ${isStacked ? 'flex-col' : 'items-start'}`}
          >
            {navigationFloating && (
              <AnnotationNavigation
                floating
                onFloatingChange={setNavigationFloating}
                floatingOpen={navigationFloatingOpen}
                onFloatingOpenChange={setNavigationFloatingOpen}
              />
            )}
            {showImage && (
              <HeightResizable
                storageKey={imagePaneHeightKey}
                defaultHeight={DEFAULT_COMPARE_PANE_HEIGHT}
                className={hasFixedImageWidth ? 'shrink-0' : 'w-full min-w-0'}
                style={imagePaneStyle}
              >
                <ImagePane
                  key={`${datasetId}:${annotationId}:${String(currentPageOrKey)}`}
                  showResizeHandle={hasFixedImageWidth}
                  onResizeStart={() => setIsResizingImagePane(true)}
                  surfaceZones={surfaceZones}
                  allZoneCategories={allZoneCategories}
                  activeLineMatchIds={activeLineMatchIds}
                  onHoverLineMatchIds={handleHoverLineMatchIds}
                  titleInset={navigationFloating}
                />
              </HeightResizable>
            )}
            {(showImage || showContents) && (
              <HeightResizable
                storageKey={teiPaneHeightKey}
                defaultHeight={DEFAULT_COMPARE_PANE_HEIGHT}
                className={`min-w-0 ${isStacked ? 'w-full' : 'flex-1'}`}
                style={showContents ? undefined : { display: 'none' }}
              >
                <TeiPane
                  key={`${datasetId}:${annotationId}:${String(currentPageOrKey)}`}
                  activeLineMatchIds={activeLineMatchIds}
                  onHoverLineMatchIds={handleHoverLineMatchIds}
                  titleInset={navigationFloating && !showImage}
                  onSurfaceZonesChange={setSurfaceZones}
                  onAllZoneCategoriesChange={setAllZoneCategories}
                />
              </HeightResizable>
            )}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="h-full flex overflow-hidden">
      {!navigationFloating && (
        <AnnotationNavigation
          floating={false}
          onFloatingChange={setNavigationFloating}
          floatingOpen={navigationFloatingOpen}
          onFloatingOpenChange={setNavigationFloatingOpen}
        />
      )}
      <div
        ref={contentRef}
        className="relative flex-1 min-h-0 flex gap-3 px-3 box-border overflow-hidden"
      >
        {navigationFloating && (
          <AnnotationNavigation
            floating
            onFloatingChange={setNavigationFloating}
            floatingOpen={navigationFloatingOpen}
            onFloatingOpenChange={setNavigationFloatingOpen}
          />
        )}
        {showImage && (
          <div
            className={`relative min-h-0 h-full empty:hidden ${showContents ? 'shrink-0' : 'flex-1 min-w-0'}`}
            style={imagePaneStyle}
          >
            <ImagePane
              key={`${datasetId}:${annotationId}:${String(currentPageOrKey)}`}
              showResizeHandle={showContents}
              onResizeStart={() => setIsResizingImagePane(true)}
              surfaceZones={surfaceZones}
              allZoneCategories={allZoneCategories}
              activeLineMatchIds={activeLineMatchIds}
              onHoverLineMatchIds={handleHoverLineMatchIds}
              titleInset={navigationFloating}
            />
          </div>
        )}
        {(showImage || showContents) && (
          <div
            className={`flex-1 min-w-0 min-h-0 h-full ${showContents ? '' : 'hidden'}`}
          >
            <TeiPane
              key={`${datasetId}:${annotationId}:${String(currentPageOrKey)}`}
              activeLineMatchIds={activeLineMatchIds}
              onHoverLineMatchIds={handleHoverLineMatchIds}
              titleInset={navigationFloating && !showImage}
              onSurfaceZonesChange={setSurfaceZones}
              onAllZoneCategoriesChange={setAllZoneCategories}
            />
          </div>
        )}
      </div>
    </div>
  )
}
