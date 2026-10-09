import { useEffect, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { AnnotationContentsTab } from '../annotation/contents/AnnotationContentsTab.tsx'
import { CompareSideProvider } from '../../context/CompareSideProvider.tsx'
import { useAppState } from '../../context/useAppState.ts'
import {
  COMPARE_SIDE_KEYS,
  type ComparePane,
  type CompareSide,
} from '../../context/compareSides.ts'
import { CompareSideSelectors } from './CompareSideSelectors.tsx'

const DISPLAY_STORAGE_KEYS = [
  'sidebarWidth',
  'sidebarCollapsed',
  'navigationFloating',
  'indexCollapsed',
  'searchCollapsed',
  'indexSearchSplitRatio',
  'imagePaneWidth',
  'imagePaneHeight',
  'imagePaneZoom',
  'teiPaneHeight',
]

const MIN_SPLIT_RATIO = 0.2
const MAX_SPLIT_RATIO = 0.8

const SideSeparator = ({
  onResizeStart,
  onReset,
}: {
  onResizeStart?: () => void
  onReset?: () => void
}) => (
  <div
    role="separator"
    aria-orientation="vertical"
    aria-label={onResizeStart ? 'Resize compare sides' : undefined}
    title={onResizeStart ? 'Drag to resize · double-click to reset' : undefined}
    onDoubleClick={onReset}
    className={`shrink-0 flex justify-center px-1 ${onResizeStart ? 'cursor-col-resize group' : ''}`}
    onPointerDown={
      onResizeStart
        ? (event) => {
            event.preventDefault()
            onResizeStart()
          }
        : undefined
    }
  >
    <div className="w-1.5 h-full bg-gray-800 rounded-full group-hover:bg-teal-600 transition-colors" />
  </div>
)

const SelectionPlaceholder = () => (
  <div className="h-full flex items-center justify-center p-6 text-center font-medium text-gray-600">
    Select a dataset and an annotation using the controls above.
  </div>
)

interface CompareSidePanesProps {
  visiblePanes: ComparePane[] | null
}

function CompareSideAnnotation({ visiblePanes }: CompareSidePanesProps) {
  const { state } = useAppState()

  if (!state.annotationId) {
    return <SelectionPlaceholder />
  }

  return (
    <AnnotationContentsTab
      compareLayout
      showImage={!visiblePanes || visiblePanes.includes('scan')}
      showContents={!visiblePanes || visiblePanes.includes('contents')}
    />
  )
}

function CompareSideContents({
  side,
  visiblePanes,
}: CompareSidePanesProps & { side: CompareSide }) {
  const { state } = useAppState()

  if (!state[COMPARE_SIDE_KEYS[side].datasetId]) {
    return <SelectionPlaceholder />
  }

  return (
    <CompareSideProvider side={side}>
      <CompareSideAnnotation visiblePanes={visiblePanes} />
    </CompareSideProvider>
  )
}

export function CompareView() {
  const [leftPanes, setLeftPanes] = useLocalStorageState<ComparePane[] | null>(
    'compare.left.visiblePanes',
    { defaultValue: null },
  )
  const [rightPanes, setRightPanes] = useLocalStorageState<
    ComparePane[] | null
  >('compare.right.visiblePanes', { defaultValue: null })
  const [splitRatio, setSplitRatio] = useLocalStorageState(
    'compare.splitRatio',
    { defaultValue: 0.5 },
  )
  const [isResizingSplit, setIsResizingSplit] = useState(false)
  const [resetCounts, setResetCounts] = useState<Record<CompareSide, number>>({
    left: 0,
    right: 0,
  })
  const splitContainerRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!isResizingSplit) return

    const onPointerMove = (event: PointerEvent) => {
      const container = splitContainerRef.current
      if (!container) return
      const rect = container.getBoundingClientRect()
      const ratio = (event.clientX - rect.left) / rect.width
      setSplitRatio(Math.min(MAX_SPLIT_RATIO, Math.max(MIN_SPLIT_RATIO, ratio)))
    }

    const onPointerUp = () => {
      setIsResizingSplit(false)
    }

    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
    return () => {
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerup', onPointerUp)
    }
  }, [isResizingSplit, setSplitRatio])

  const resetDisplay = (side: CompareSide) => {
    DISPLAY_STORAGE_KEYS.forEach((key) => {
      try {
        localStorage.removeItem(`compare.${side}.${key}`)
      } catch {
        return
      }
    })
    if (side === 'left') {
      setLeftPanes(null)
    } else {
      setRightPanes(null)
    }
    setResetCounts((counts) => ({ ...counts, [side]: counts[side] + 1 }))
  }

  const leftSideStyle = { width: `${splitRatio * 100}%` }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex w-full p-3 border-b border-gray-200 bg-white">
        <div
          className="shrink-0 min-w-0 flex justify-center"
          style={leftSideStyle}
        >
          <CompareSideSelectors
            side="left"
            selectedPanes={leftPanes}
            setSelectedPanes={setLeftPanes}
            onResetDisplay={() => resetDisplay('left')}
          />
        </div>
        <SideSeparator />
        <div className="flex-1 min-w-0 flex justify-center">
          <CompareSideSelectors
            side="right"
            selectedPanes={rightPanes}
            setSelectedPanes={setRightPanes}
            onResetDisplay={() => resetDisplay('right')}
          />
        </div>
      </div>
      <div
        ref={splitContainerRef}
        className={`flex-1 min-h-0 m-3 flex overflow-hidden ${isResizingSplit ? 'select-none' : ''}`}
      >
        <div
          className="shrink-0 min-w-0 h-full overflow-hidden"
          style={leftSideStyle}
        >
          <CompareSideContents
            key={resetCounts.left}
            side="left"
            visiblePanes={leftPanes}
          />
        </div>
        <SideSeparator
          onResizeStart={() => setIsResizingSplit(true)}
          onReset={() => setSplitRatio(0.5)}
        />
        <div className="flex-1 min-w-0 h-full overflow-hidden">
          <CompareSideContents
            key={resetCounts.right}
            side="right"
            visiblePanes={rightPanes}
          />
        </div>
      </div>
    </div>
  )
}
