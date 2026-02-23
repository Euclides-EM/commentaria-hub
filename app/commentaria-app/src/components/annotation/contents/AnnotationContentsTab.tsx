import { ImagePane } from './ImagePane.tsx'
import { TeiPane } from './tei/TeiPane.tsx'
import { AnnotationNavigation } from './AnnotationNavigation.tsx'
import { useAppState } from '../../../context/useAppState.ts'
import { useEffect, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'

export function AnnotationContentsTab() {
  const { annotation, dataset } = useAppState()
  const showTeiPane = !!annotation?.ocred || !!dataset?.edition_id
  const [imagePaneWidth, setImagePaneWidth] = useLocalStorageState(
    'imagePaneWidth',
    { defaultValue: 560 },
  )
  const [isResizingImagePane, setIsResizingImagePane] = useState(false)
  const contentRef = useRef<HTMLDivElement | null>(null)

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
      setImagePaneWidth((current) => clampWidth(current))
    }

    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
    }
  }, [isResizingImagePane, setImagePaneWidth])

  return (
    <div className="h-full flex overflow-hidden">
      <AnnotationNavigation />
      <div
        ref={contentRef}
        className="flex-1 min-h-0 flex gap-3 p-3 box-border overflow-hidden"
      >
        <div
          className={
            showTeiPane
              ? 'relative min-h-0 h-full shrink-0'
              : 'min-h-0 h-full flex-1'
          }
          style={
            showTeiPane
              ? {
                  width: `${imagePaneWidth}px`,
                  minWidth: `${imagePaneWidth}px`,
                  maxWidth: `${imagePaneWidth}px`,
                }
              : undefined
          }
        >
          <ImagePane
            showResizeHandle={showTeiPane}
            onResizeStart={() => setIsResizingImagePane(true)}
          />
        </div>
        {showTeiPane && (
          <div className="flex-1 min-w-0 min-h-0 h-full">
            <TeiPane />
          </div>
        )}
      </div>
    </div>
  )
}
