import { useEffect, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { useAppState } from '../../../../context/useAppState.ts'
import {
  getTeiParagraphSelection,
  teiToHtml,
  type TeiHighlightConfig,
  type TeiViewMode,
} from './tei.ts'
import type {
  SelectionDraft,
  TeiTooltipItem,
  TeiTooltipState,
} from './TeiPane.types.ts'
import { parseLineMatchIds } from './teiPaneUtils.tsx'

const TEI_HIGHLIGHT_SELECTOR = '[data-tei-highlight="true"]'
const TEI_LINE_MATCH_SELECTOR = '[data-tei-line-match-ids]'

const getTooltipItems = (element: Element | null): TeiTooltipItem[] => {
  if (!element) return []
  const highlighted = element.closest(TEI_HIGHLIGHT_SELECTOR)
  if (!(highlighted instanceof HTMLElement)) {
    return []
  }
  const payload = highlighted.dataset.teiHighlightTooltip || ''
  if (!payload) {
    return []
  }
  try {
    const parsed = JSON.parse(decodeURIComponent(payload))
    if (!Array.isArray(parsed)) {
      return []
    }
    return parsed
      .filter((item) => item && typeof item === 'object')
      .map((item) => {
        const value = item as Record<string, unknown>
        return {
          id: String(value.id || ''),
          featureId: String(value.featureId || ''),
          categoryId: String(value.categoryId || ''),
          label: String(value.label || ''),
          description: String(value.description || ''),
          surface: String(value.surface || ''),
          normalized: String(value.normalized || ''),
          institution: String(value.institution || ''),
          ancientPersona: String(value.ancientPersona || ''),
          paragraphIndex: Number(value.paragraphIndex || 0),
          start: Number(value.start || 0),
          end: Number(value.end || 0),
          fromAnchorId: String(value.fromAnchorId || ''),
          toAnchorId: String(value.toAnchorId || ''),
          color: String(value.color || '#f2f2f2'),
        }
      })
      .filter((item) => item.id && item.label)
  } catch {
    return []
  }
}

const getTooltipPosition = (element: Element) => {
  const TOOLTIP_MAX_WIDTH = 384
  const TOOLTIP_ESTIMATED_HEIGHT = 240
  const VIEWPORT_MARGIN = 12
  const rect = element.getBoundingClientRect()
  let x = rect.left + rect.width / 2 + 12
  let y = rect.top + 14

  if (x + TOOLTIP_MAX_WIDTH + VIEWPORT_MARGIN > window.innerWidth) {
    x = window.innerWidth - TOOLTIP_MAX_WIDTH - VIEWPORT_MARGIN
  }
  if (x < VIEWPORT_MARGIN) {
    x = VIEWPORT_MARGIN
  }

  if (y + TOOLTIP_ESTIMATED_HEIGHT + VIEWPORT_MARGIN > window.innerHeight) {
    y = rect.bottom - TOOLTIP_ESTIMATED_HEIGHT - 14
  }
  if (y < VIEWPORT_MARGIN) {
    y = VIEWPORT_MARGIN
  }

  return {
    x,
    y,
  }
}

const getTooltipItemsKey = (items: TeiTooltipItem[]) =>
  items
    .map((item) =>
      [item.id, item.featureId, item.paragraphIndex, item.start, item.end].join(
        ':',
      ),
    )
    .sort()
    .join('|')

const getSelectionDraft = (
  root: HTMLElement,
  selection: Selection,
): SelectionDraft | null => {
  if (!selection.rangeCount || selection.isCollapsed) {
    return null
  }
  const range = selection.getRangeAt(0)
  if (
    !root.contains(range.startContainer) ||
    !root.contains(range.endContainer)
  ) {
    return null
  }

  const startParent =
    range.startContainer instanceof Element
      ? range.startContainer
      : range.startContainer.parentElement
  const endParent =
    range.endContainer instanceof Element
      ? range.endContainer
      : range.endContainer.parentElement

  const startParagraph = startParent?.closest('p[data-tei-paragraph-index]')
  const endParagraph = endParent?.closest('p[data-tei-paragraph-index]')

  if (!startParagraph || !endParagraph || startParagraph !== endParagraph) {
    return null
  }

  const paragraphIndex = Number.parseInt(
    startParagraph.getAttribute('data-tei-paragraph-index') || '',
    10,
  )
  if (Number.isNaN(paragraphIndex)) {
    return null
  }

  const paragraphSelection = getTeiParagraphSelection(startParagraph, range)
  if (!paragraphSelection) {
    return null
  }

  const rect = range.getBoundingClientRect()
  return {
    paragraphIndex,
    start: paragraphSelection.start,
    end: paragraphSelection.end,
    surface: paragraphSelection.surface,
    x: rect.left + rect.width / 2 + 8,
    y: rect.top - 34,
  }
}

type TeiContentViewProps = {
  data: string
  minCert: number
  showCertaintyVisualization: boolean
  viewMode: TeiViewMode
  viewLabel: string
  showViewLabel: boolean
  alignLines: boolean
  centerRows: boolean
  highlightConfig?: TeiHighlightConfig
  editable: boolean
  canAddHighlight: boolean
  noFrame?: boolean
  activeLineMatchIds: string[]
  onHoverLineMatchIds: (ids: string[]) => void
  onRequestAddHighlight: (selection: SelectionDraft) => void
  onRequestRemoveHighlight: (item: TeiTooltipItem) => void
}

export const TeiContentView = ({
  minCert,
  showCertaintyVisualization,
  data,
  viewMode,
  viewLabel,
  showViewLabel,
  alignLines,
  centerRows,
  highlightConfig,
  editable,
  canAddHighlight,
  noFrame,
  activeLineMatchIds,
  onHoverLineMatchIds,
  onRequestAddHighlight,
  onRequestRemoveHighlight,
}: TeiContentViewProps) => {
  const { searchResultHighlight } = useAppState()
  const [tooltipState, setTooltipState] = useState<TeiTooltipState | null>(null)
  const [selectionState, setSelectionState] = useState<SelectionDraft | null>(
    null,
  )
  const hideTooltipTimerRef = useRef<number | null>(null)
  const tooltipHoveredRef = useRef(false)
  const rootRef = useRef<HTMLDivElement | null>(null)

  const clearHideTooltipTimer = () => {
    if (hideTooltipTimerRef.current == null) return
    window.clearTimeout(hideTooltipTimerRef.current)
    hideTooltipTimerRef.current = null
  }

  const scheduleHideTooltip = () => {
    if (tooltipHoveredRef.current) return
    clearHideTooltipTimer()
    hideTooltipTimerRef.current = window.setTimeout(() => {
      if (tooltipHoveredRef.current) return
      setTooltipState(null)
      hideTooltipTimerRef.current = null
    }, 180)
  }

  useEffect(() => () => clearHideTooltipTimer(), [])
  useEffect(() => {
    if (!tooltipState) {
      tooltipHoveredRef.current = false
    }
  }, [tooltipState])

  const html = useMemo(
    () =>
      teiToHtml(
        data,
        minCert,
        searchResultHighlight,
        '@',
        viewMode,
        alignLines,
        showCertaintyVisualization,
        highlightConfig,
      ),
    [
      alignLines,
      data,
      highlightConfig,
      minCert,
      searchResultHighlight,
      showCertaintyVisualization,
      viewMode,
    ],
  )

  useEffect(() => {
    const root = rootRef.current
    if (!root) {
      return
    }
    const activeSet = new Set(activeLineMatchIds)
    const lineElements = root.querySelectorAll<HTMLElement>(
      TEI_LINE_MATCH_SELECTOR,
    )
    lineElements.forEach((lineElement) => {
      const ids = parseLineMatchIds(lineElement.dataset.teiLineMatchIds)
      const isActive = ids.length > 0 && ids.some((id) => activeSet.has(id))
      if (isActive) {
        lineElement.setAttribute('data-tei-corresp-hovered', 'true')
      } else {
        lineElement.removeAttribute('data-tei-corresp-hovered')
      }
    })
  }, [activeLineMatchIds, html])

  const tooltip =
    tooltipState &&
    createPortal(
      <div
        style={{
          position: 'fixed',
          left: tooltipState.x,
          top: tooltipState.y,
          zIndex: 12000,
          pointerEvents: 'auto',
          backgroundColor: 'white',
          color: 'black',
          padding: '0.75rem',
          borderRadius: '0.35rem',
          fontSize: '0.75rem',
          lineHeight: 1.3,
          maxWidth: '24rem',
          maxHeight: 'min(60vh, 420px)',
          overflowY: 'auto',
          border: '1px solid #d9d9d9',
          boxShadow: '0 6px 16px rgba(0, 0, 0, 0.2)',
        }}
        onMouseEnter={() => {
          tooltipHoveredRef.current = true
          clearHideTooltipTimer()
        }}
        onMouseLeave={() => {
          tooltipHoveredRef.current = false
          setTooltipState(null)
        }}
      >
        <div className="flex flex-col gap-2">
          {tooltipState.items.map((item) => (
            <div
              key={`${item.id}:${item.featureId}:${item.start}:${item.end}`}
              className="flex flex-col gap-1"
            >
              <div className="flex items-center gap-1.5">
                <span
                  className="inline-block rounded px-1.5 py-0.5 font-semibold"
                  style={{ backgroundColor: item.color }}
                >
                  {item.label}
                </span>
              </div>
              {item.description && (
                <div className="text-gray-700 mt-0.5">{item.description}</div>
              )}
              {item.surface && (
                <div className="mt-0.5 italic text-teal-800 bg-teal-50/70 rounded px-1.5 py-0.5">
                  "{item.surface}"
                </div>
              )}
              {item.normalized && (
                <div className="mt-0.5 text-teal-700">
                  Normalized: {item.normalized}
                </div>
              )}
              {(item.institution || item.ancientPersona) && (
                <div className="flex flex-wrap gap-1 mt-1">
                  {item.institution && (
                    <span className="inline-block rounded border border-gray-300 bg-gray-100 px-1.5 py-0.5 text-[11px] text-gray-700">
                      Institution: {item.institution}
                    </span>
                  )}
                  {item.ancientPersona && (
                    <span className="inline-block rounded border border-gray-300 bg-gray-100 px-1.5 py-0.5 text-[11px] text-gray-700">
                      Ancient persona: {item.ancientPersona}
                    </span>
                  )}
                </div>
              )}
              {editable && viewMode === 'original' && (
                <div className="mt-3 flex items-center gap-1.5">
                  <button
                    type="button"
                    className="px-1.5 py-0.5 rounded border border-red-300 text-red-700 hover:bg-red-50"
                    onClick={() => {
                      onRequestRemoveHighlight(item)
                      setTooltipState(null)
                    }}
                  >
                    Remove
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>,
      document.body,
    )

  const selectionTooltip =
    selectionState &&
    !tooltipState &&
    editable &&
    canAddHighlight &&
    viewMode === 'original' &&
    createPortal(
      <div
        style={{
          position: 'fixed',
          left: selectionState.x,
          top: selectionState.y,
          zIndex: 12001,
          pointerEvents: 'auto',
        }}
        onMouseDown={(event) => {
          event.stopPropagation()
        }}
        onMouseUp={(event) => {
          event.stopPropagation()
        }}
      >
        <button
          type="button"
          className="px-2 py-1 rounded border border-teal-300 bg-white text-teal-700 text-xs font-semibold shadow-sm hover:bg-teal-50"
          onClick={() => {
            onRequestAddHighlight(selectionState)
            setSelectionState(null)
            const selection = window.getSelection()
            selection?.removeAllRanges()
          }}
        >
          Highlight
        </button>
      </div>,
      document.body,
    )

  return (
    <div
      ref={rootRef}
      className="relative"
      onMouseDown={() => {
        setSelectionState(null)
      }}
      onMouseMove={(event) => {
        if (selectionState) {
          onHoverLineMatchIds([])
          clearHideTooltipTimer()
          setTooltipState(null)
          return
        }

        const elements = document.elementsFromPoint(
          event.clientX,
          event.clientY,
        )
        const lineElement = elements.find((el) => {
          if (!(el instanceof Element)) {
            return false
          }
          return (
            el.matches(TEI_LINE_MATCH_SELECTOR) ||
            !!el.closest(TEI_LINE_MATCH_SELECTOR)
          )
        }) as Element | undefined
        const lineMatchIds = lineElement
          ? parseLineMatchIds(
              (lineElement.matches(TEI_LINE_MATCH_SELECTOR)
                ? lineElement
                : lineElement.closest(TEI_LINE_MATCH_SELECTOR)
              )?.getAttribute('data-tei-line-match-ids'),
            )
          : []
        onHoverLineMatchIds(lineMatchIds)

        if (tooltipHoveredRef.current) {
          return
        }
        clearHideTooltipTimer()
        const hitElement =
          elements.find((el) => el.matches?.(TEI_HIGHLIGHT_SELECTOR)) ||
          elements.find((el) => el.closest?.(TEI_HIGHLIGHT_SELECTOR)) ||
          (event.target instanceof Element ? event.target : null)
        const items = getTooltipItems(hitElement)
        if (!items.length || !hitElement) {
          if (tooltipHoveredRef.current) {
            clearHideTooltipTimer()
            return
          }
          scheduleHideTooltip()
          return
        }
        const position = getTooltipPosition(hitElement)
        setSelectionState(null)
        setTooltipState((previous) => {
          if (!previous) {
            return { x: position.x, y: position.y, items }
          }
          const previousKey = getTooltipItemsKey(previous.items)
          const nextKey = getTooltipItemsKey(items)
          if (previousKey !== nextKey) {
            return { x: position.x, y: position.y, items }
          }
          return { x: position.x, y: position.y, items }
        })
      }}
      onMouseLeave={() => {
        onHoverLineMatchIds([])
        if (tooltipHoveredRef.current) return
        scheduleHideTooltip()
      }}
      onMouseUp={() => {
        if (!editable || !canAddHighlight || viewMode !== 'original') {
          setSelectionState(null)
          return
        }
        if (tooltipState) {
          setSelectionState(null)
          return
        }
        const root = rootRef.current
        if (!root) {
          setSelectionState(null)
          return
        }
        const selection = window.getSelection()
        if (!selection) {
          setSelectionState(null)
          return
        }
        const nextSelection = getSelectionDraft(root, selection)
        setSelectionState(nextSelection)
      }}
    >
      {showViewLabel && (
        <div className="absolute top-2 right-2 z-10 rounded bg-white/90 border border-gray-300 px-1.5 py-0.5 text-[10px] font-medium text-gray-700">
          {viewLabel}
        </div>
      )}
      <div
        className={`text-xs leading-relaxed p-2 ${noFrame ? '' : 'border border-gray-300 rounded-xl bg-gray-50'} ${showViewLabel ? 'pt-7' : ''} [&_p]:mb-2 [&_p:last-child]:mb-0 ${centerRows ? '[&_p]:text-center' : ''} [&_[data-tei-selected='true']]:bg-yellow-200/70 [&_[data-tei-selected='true']]:text-gray-900 [&_[data-tei-selected='true']]:rounded-sm [&_[data-tei-selected='true']]:px-0.5 [&_[data-tei-corresp-hovered='true']]:bg-teal-100/70 [&_[data-tei-corresp-hovered='true']]:outline [&_[data-tei-corresp-hovered='true']]:outline-1 [&_[data-tei-corresp-hovered='true']]:outline-teal-500/70 [&_[data-tei-corresp-hovered='true']]:rounded-sm`}
        style={{ whiteSpace: 'normal' }}
        dangerouslySetInnerHTML={{ __html: html }}
      />
      {tooltip}
      {selectionTooltip}
    </div>
  )
}
