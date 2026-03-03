import {
  getTeiZoneToServerTextBlockId,
  getTeiOriginalEditableLines,
  getTeiHighlightCategories,
  getTeiSurfaceZones,
  getTeiTranslations,
  hasTeiCertaintyDegrees,
  type TeiHighlightConfig,
  type TeiOriginalEditableLine,
  type TeiManualHighlight,
  type TeiSurfaceZone,
  teiToHtml,
  type TeiTranslation,
  type TeiViewMode,
} from './tei.ts'
import { useAppState } from '../../../../context/useAppState.ts'
import { useEffect, useMemo, useRef, useState } from 'react'
import {
  annotationTeiQueryKey,
  editionTeiQueryKey,
  useAnnotationTeiQuery,
  useEditionTeiQuery,
} from '../../../../queries/annotations.ts'
import { useDatasetFeaturesQuery } from '../../../../queries/datasets.ts'
import useLocalStorageState from 'use-local-storage-state'
import { RangeInput } from '../../../core/RangeInput.tsx'
import Select, { type StylesConfig } from 'react-select'
import { selectStyles } from '../../../../styles/selectStyles.ts'
import { MultiSelectDropdown } from '../../../core/MultiSelectDropdown.tsx'
import { createPortal } from 'react-dom'
import {
  AnnotationsApplyRulesService,
  type annotationrule_TextBlockCorrections,
  type feature_Feature,
  type feature_Result,
  type feature_ResultValue,
  FeatureResultsService,
} from '@hub-api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../../../../store/authStore.ts'

const VIEW_LABEL_MAP: Record<string, string> = {
  en: 'English',
}

const normalizeTeiViewModes = (
  viewModes: TeiViewMode[],
  allowedViewModes: TeiViewMode[],
): TeiViewMode[] => {
  if (!allowedViewModes.length) {
    return ['original']
  }
  const allowed = new Set(allowedViewModes)
  const next = viewModes.filter((mode) => allowed.has(mode))
  if (!next.length) {
    return allowedViewModes
  }
  return next
}

const normalizeMatchKey = (value: string | null | undefined) =>
  (value || '').toLowerCase().replace(/[^a-z0-9]+/g, '')

type ResolvedTeiFeature = {
  id: string
  label: string
  description: string
  color: string
  isDefault: boolean
  renderMode: 'fill' | 'outline'
}

const matchTeiCategoryToFeature = (
  categoryId: string,
  categoryLabel: string,
  features: feature_Feature[],
) => {
  const categoryCandidates = [
    categoryId,
    categoryId.replace(/^cat_/, ''),
    categoryLabel,
  ]
    .map((value) => normalizeMatchKey(value))
    .filter(Boolean)

  for (const feature of features) {
    const featureCandidates = [
      normalizeMatchKey(feature.id),
      normalizeMatchKey(feature.name),
      normalizeMatchKey((feature as { key?: string }).key),
      normalizeMatchKey((feature as { slug?: string }).slug),
    ].filter(Boolean)

    if (
      categoryCandidates.some((candidate) =>
        featureCandidates.includes(candidate),
      )
    ) {
      return feature
    }
  }

  return null
}

const isVerbLike = (...values: Array<string | null | undefined>) =>
  values.some((value) => normalizeMatchKey(value).includes('verb'))

type TeiTooltipItem = {
  id: string
  featureId: string
  categoryId: string
  label: string
  description: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
  paragraphIndex: number
  start: number
  end: number
  fromAnchorId: string
  toAnchorId: string
  color: string
}

type TeiTooltipState = {
  x: number
  y: number
  items: TeiTooltipItem[]
}

type SelectionDraft = {
  paragraphIndex: number
  start: number
  end: number
  surface: string
  x: number
  y: number
}

type DraftHighlight = {
  localId: string
  sourceId: string
  paragraphIndex: number
  start: number
  end: number
  featureId: string
  categoryId: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
  fromAnchorId: string
  toAnchorId: string
}

type FeatureModalState = {
  selection: SelectionDraft
}

const TEI_HIGHLIGHT_SELECTOR = '[data-tei-highlight="true"]'
const TEI_LINE_MATCH_SELECTOR = '[data-tei-line-match-ids]'

const parseLineMatchIds = (value: string | null | undefined) =>
  [
    ...new Set(
      (value || '')
        .split(/\s+/)
        .map((entry) => entry.trim())
        .filter(Boolean),
    ),
  ].sort()

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

const getOffsetInParagraph = (
  paragraph: Element,
  node: Node,
  offset: number,
): number | null => {
  try {
    const range = document.createRange()
    range.setStart(paragraph, 0)
    range.setEnd(node, offset)
    return range.toString().length
  } catch {
    return null
  }
}

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

  const startOffset = getOffsetInParagraph(
    startParagraph,
    range.startContainer,
    range.startOffset,
  )
  const endOffset = getOffsetInParagraph(
    startParagraph,
    range.endContainer,
    range.endOffset,
  )
  if (startOffset == null || endOffset == null) {
    return null
  }

  const start = Math.min(startOffset, endOffset)
  const end = Math.max(startOffset, endOffset)
  if (end <= start) {
    return null
  }

  const paragraphText = startParagraph.textContent || ''
  const surface = paragraphText.slice(start, end)
  if (!surface.trim()) {
    return null
  }

  const rect = range.getBoundingClientRect()
  return {
    paragraphIndex,
    start,
    end,
    surface,
    x: rect.left + rect.width / 2 + 8,
    y: rect.top - 34,
  }
}

type TeiProps = {
  data: string
  minCert: number
  viewMode: TeiViewMode
  viewLabel: string
  showViewLabel: boolean
  alignLines: boolean
  highlightConfig?: TeiHighlightConfig
  editable: boolean
  canAddHighlight: boolean
  activeLineMatchIds: string[]
  enableHoverSync: boolean
  onHoverLineMatchIds: (ids: string[]) => void
  onRequestAddHighlight: (selection: SelectionDraft) => void
  onRequestRemoveHighlight: (item: TeiTooltipItem) => void
}

const Tei = ({
  minCert,
  data,
  viewMode,
  viewLabel,
  showViewLabel,
  alignLines,
  highlightConfig,
  editable,
  canAddHighlight,
  activeLineMatchIds,
  enableHoverSync,
  onHoverLineMatchIds,
  onRequestAddHighlight,
  onRequestRemoveHighlight,
}: TeiProps) => {
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
        highlightConfig,
      ),
    [
      alignLines,
      data,
      highlightConfig,
      minCert,
      searchResultHighlight,
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
      const isActive =
        enableHoverSync && ids.length > 0 && ids.some((id) => activeSet.has(id))
      if (isActive) {
        lineElement.setAttribute('data-tei-corresp-hovered', 'true')
      } else {
        lineElement.removeAttribute('data-tei-corresp-hovered')
      }
    })
  }, [activeLineMatchIds, enableHoverSync, html])

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
        onHoverLineMatchIds(enableHoverSync ? lineMatchIds : [])

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
        if (tooltipState) {
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
        className={`text-xs leading-relaxed border border-gray-300 rounded-xl bg-gray-50 p-2 ${showViewLabel ? 'pt-7' : ''} [&_p]:mb-2 [&_p:last-child]:mb-0 [&_[data-tei-selected='true']]:bg-yellow-200/70 [&_[data-tei-selected='true']]:text-gray-900 [&_[data-tei-selected='true']]:rounded-sm [&_[data-tei-selected='true']]:px-0.5 [&_[data-tei-corresp-hovered='true']]:bg-teal-100/70 [&_[data-tei-corresp-hovered='true']]:outline [&_[data-tei-corresp-hovered='true']]:outline-1 [&_[data-tei-corresp-hovered='true']]:outline-teal-500/70 [&_[data-tei-corresp-hovered='true']]:rounded-sm`}
        style={{ whiteSpace: 'normal' }}
        dangerouslySetInnerHTML={{ __html: html }}
      />
      {tooltip}
      {selectionTooltip}
    </div>
  )
}

type SourceOption = {
  value: 'annotation' | 'edition'
  label: string
}

type FeatureOption = {
  value: string
  label: string
  color: string
  description: string
}

type EditableOriginalLine = TeiOriginalEditableLine & {
  originalText: string
}

type OriginalViewEditorProps = {
  lines: EditableOriginalLine[]
  showViewLabel: boolean
  onChangeLine: (lineId: string, text: string) => void
}

const OriginalViewEditor = ({
  lines,
  showViewLabel,
  onChangeLine,
}: OriginalViewEditorProps) => (
  <div className="relative">
    {showViewLabel && (
      <div className="absolute top-2 right-2 z-10 rounded bg-white/90 border border-gray-300 px-1.5 py-0.5 text-[10px] font-medium text-gray-700">
        Original
      </div>
    )}
    <div
      className={`text-xs leading-relaxed border border-gray-300 rounded-xl bg-gray-50 p-2 ${showViewLabel ? 'pt-7' : ''} flex flex-col gap-1.5`}
    >
      {lines.map((line) => (
        <input
          key={line.id}
          type="text"
          value={line.text}
          onChange={(event) => onChangeLine(line.id, event.target.value)}
          className="w-full border border-gray-300 rounded-md px-2 py-1 bg-white text-xs text-gray-800 focus:outline-none focus:ring-2 focus:ring-teal-200 focus:border-teal-400"
          spellCheck={false}
        />
      ))}
    </div>
  </div>
)

const formatFeatureOptionLabel = (
  option: FeatureOption,
  context: 'menu' | 'value',
) =>
  context === 'value' ? (
    <span>{option.label}</span>
  ) : (
    <div className="flex items-center gap-2">
      <span
        className="w-2.5 h-2.5 rounded-full shrink-0"
        style={{ backgroundColor: option.color }}
      />
      <div>{option.label}</div>
    </div>
  )

const featureSelectStyles: StylesConfig<FeatureOption, true> = {
  control: (base, state) => ({
    ...base,
    minHeight: 32,
    border: `1px solid ${state.isFocused ? '#14b8a6' : '#9ca3af'}`,
    borderRadius: 6,
    boxShadow: state.isFocused ? '0 0 0 3px rgba(20, 184, 166, 0.15)' : 'none',
  }),
  valueContainer: (base) => ({
    ...base,
    padding: '0 6px',
  }),
  menuPortal: (base) => ({ ...base, zIndex: 1000 }),
  option: (base, state) => ({
    ...base,
    backgroundColor: state.isFocused ? '#f3f4f6' : 'white',
    color: '#374151',
  }),
  multiValue: (base, state) => ({
    ...base,
    backgroundColor: state.data.color || '#f2f2f2',
    borderRadius: 4,
    padding: '0 2px',
  }),
  multiValueLabel: (base) => ({
    ...base,
    color: '#111827',
    fontWeight: 600,
    fontSize: '11px',
    lineHeight: 1.1,
    padding: '2px 4px',
    maxWidth: 240,
  }),
  multiValueRemove: (base) => ({
    ...base,
    color: '#111827',
    padding: '0 3px',
    ':hover': { backgroundColor: 'rgba(255,255,255,0.6)', color: '#111827' },
  }),
}

const featureModalStyles: StylesConfig<FeatureOption, false> = {
  ...selectStyles<FeatureOption>(),
  menuPortal: (base) => ({ ...base, zIndex: 13000 }),
}

const sameStringArray = (left: string[] | null, right: string[]) => {
  if (!left) return false
  if (left.length !== right.length) return false
  return left.every((value, index) => value === right[index])
}

const toResultValues = (highlights: DraftHighlight[]): feature_ResultValue[] =>
  highlights
    .map((highlight) => {
      const properties: Record<string, string> = {}
      if (highlight.normalized) {
        properties.normalized = highlight.normalized
      }
      if (highlight.institution) {
        properties.institution = highlight.institution
      }
      if (highlight.ancientPersona) {
        properties.ancient_persona = highlight.ancientPersona
      }
      return {
        surface: highlight.surface,
        properties,
      }
    })
    .sort((left, right) => {
      const leftKey = `${left.properties?.paragraph_index || ''}:${left.properties?.start || ''}:${left.properties?.end || ''}:${left.surface || ''}`
      const rightKey = `${right.properties?.paragraph_index || ''}:${right.properties?.start || ''}:${right.properties?.end || ''}:${right.surface || ''}`
      return leftKey.localeCompare(rightKey)
    })

const getComparableValues = (values: feature_ResultValue[]) =>
  values
    .map((value) => ({
      surface: value.surface || '',
      properties: Object.entries(value.properties || {})
        .sort(([left], [right]) => left.localeCompare(right))
        .map(([key, nextValue]) => `${key}:${nextValue}`)
        .join('|'),
    }))
    .sort((left, right) =>
      `${left.properties}:${left.surface}`.localeCompare(
        `${right.properties}:${right.surface}`,
      ),
    )

const groupByFeature = (highlights: DraftHighlight[]) => {
  const grouped: Record<string, DraftHighlight[]> = {}
  for (const highlight of highlights) {
    grouped[highlight.featureId] = grouped[highlight.featureId] || []
    grouped[highlight.featureId].push(highlight)
  }
  return grouped
}

const debugTeiHighlights = (...args: unknown[]) => {
  if (!import.meta.env.DEV) {
    return
  }
  console.log('[TeiPane highlights]', ...args)
}

const toDraftHighlightsFromResults = (
  results: feature_Result[],
): DraftHighlight[] => {
  const out: DraftHighlight[] = []
  for (const result of results) {
    const featureId = result.feature_id || ''
    if (!featureId) {
      continue
    }
    const values = result.values || []
    for (let index = 0; index < values.length; index++) {
      const value = values[index]
      const properties = value.properties || {}
      const paragraphIndex = Number.parseInt(
        properties.paragraph_index || '',
        10,
      )
      const start = Number.parseInt(properties.start || '', 10)
      const end = Number.parseInt(properties.end || '', 10)
      if (
        Number.isNaN(paragraphIndex) ||
        Number.isNaN(start) ||
        Number.isNaN(end) ||
        end <= start
      ) {
        continue
      }
      const sourceId =
        properties.highlight_id ||
        `${result.id || featureId}:${paragraphIndex}:${start}:${end}:${index}`
      const localId = `${sourceId}:${featureId}`
      out.push({
        localId,
        sourceId,
        paragraphIndex,
        start,
        end,
        featureId,
        categoryId: properties.category_id || featureId,
        surface: value.surface || '',
        normalized: properties.normalized || '',
        institution: properties.institution || '',
        ancientPersona: properties.ancient_persona || '',
        fromAnchorId: properties.from_anchor_id || '',
        toAnchorId: properties.to_anchor_id || '',
      })
    }
  }
  return out
}

type TeiPaneProps = {
  activeLineMatchIds: string[]
  enableHoverSync: boolean
  onHoverLineMatchIds: (ids: string[]) => void
  onEnableHoverSyncChange: (enabled: boolean) => void
  onSurfaceZonesChange: (zones: TeiSurfaceZone[]) => void
}

export function TeiPane({
  activeLineMatchIds,
  enableHoverSync,
  onHoverLineMatchIds,
  onEnableHoverSyncChange,
  onSurfaceZonesChange,
}: TeiPaneProps) {
  const {
    annotation,
    dataset,
    state: { datasetId, annotationId, currentPageOrKey },
  } = useAppState()
  const queryClient = useQueryClient()
  const isAuthenticated = !!useAuthStore((store) => store.token)

  const [showTeiSource, setShowTeiSource] = useLocalStorageState(
    'showTeiSource',
    { defaultValue: false, storageSync: false },
  )
  const [minCert, setMinCert] = useLocalStorageState('minCert', {
    defaultValue: 0.8,
    storageSync: false,
  })
  const [alignLines, setAlignLines] = useLocalStorageState('alignTeiLines', {
    defaultValue: false,
    storageSync: false,
  })
  const [isFeatureSelectExpanded, setIsFeatureSelectExpanded] =
    useLocalStorageState('teiFeatureSelectExpanded', {
      defaultValue: false,
      storageSync: false,
    })
  const [featureModalState, setFeatureModalState] =
    useState<FeatureModalState | null>(null)
  const [modalFeatureId, setModalFeatureId] = useState<string>('')
  const [saveError, setSaveError] = useState<string | null>(null)
  const [isTextEditMode, setIsTextEditMode] = useState(false)
  const [textEditError, setTextEditError] = useState<string | null>(null)
  const [editableOriginalLines, setEditableOriginalLines] = useState<
    EditableOriginalLine[]
  >([])
  const [baselineHighlights, setBaselineHighlights] = useState<
    DraftHighlight[]
  >([])
  const [draftHighlights, setDraftHighlights] = useState<DraftHighlight[]>([])
  const [removedTeiHighlightIds, setRemovedTeiHighlightIds] = useState<
    string[]
  >([])
  const [forcedChangedFeatureIds, setForcedChangedFeatureIds] = useState<
    string[]
  >([])
  const ocred = !!annotation?.ocred
  const editionId = dataset?.edition_id

  const candidateSources = useMemo(() => {
    const sources: Array<'annotation' | 'edition'> = []
    if (ocred) {
      sources.push('annotation')
    }
    if (editionId) {
      sources.push('edition')
    }
    return sources
  }, [editionId, ocred])

  const [storedTeiSource, setStoredTeiSource] = useLocalStorageState<
    'annotation' | 'edition'
  >('teiSource', { defaultValue: 'annotation', storageSync: false })
  const preferredTeiSource = candidateSources.includes(storedTeiSource)
    ? storedTeiSource
    : candidateSources[0]

  const editionTeiQuery = useEditionTeiQuery(
    editionId,
    currentPageOrKey,
    !!editionId,
  )
  const annotationTeiQuery = useAnnotationTeiQuery(
    datasetId,
    annotationId,
    currentPageOrKey,
    ocred,
  )
  const featuresQuery = useDatasetFeaturesQuery(datasetId, !!datasetId)
  const featureResultsQuery = useQuery({
    queryKey: ['featureResults', datasetId, annotationId, currentPageOrKey],
    queryFn: () =>
      FeatureResultsService.getDatasetsAnnotationsResults({
        dataSetId: datasetId,
        id: annotationId,
        keys: String(currentPageOrKey),
      }),
    enabled: !!datasetId && !!annotationId,
  })

  const availableSources = useMemo(() => {
    const sources: Array<'annotation' | 'edition'> = []
    if (ocred && annotationTeiQuery.isSuccess && annotationTeiQuery.data) {
      sources.push('annotation')
    }
    if (editionId && editionTeiQuery.isSuccess && editionTeiQuery.data) {
      sources.push('edition')
    }
    return sources
  }, [
    editionId,
    ocred,
    annotationTeiQuery.data,
    annotationTeiQuery.isSuccess,
    editionTeiQuery.data,
    editionTeiQuery.isSuccess,
  ])

  const effectiveTeiSource = availableSources.includes(storedTeiSource)
    ? storedTeiSource
    : availableSources[0] || preferredTeiSource

  useEffect(() => {
    if (
      availableSources.length > 0 &&
      effectiveTeiSource &&
      effectiveTeiSource !== storedTeiSource
    ) {
      setStoredTeiSource(effectiveTeiSource)
    }
  }, [
    availableSources,
    effectiveTeiSource,
    setStoredTeiSource,
    storedTeiSource,
  ])

  const data =
    effectiveTeiSource === 'edition'
      ? editionTeiQuery.data
      : effectiveTeiSource === 'annotation'
        ? annotationTeiQuery.data
        : undefined

  const teiContents = data ?? null
  const baseEditableOriginalLines = useMemo(
    () =>
      teiContents
        ? getTeiOriginalEditableLines(teiContents).map((line, index) => ({
            ...line,
            id: `${line.id}:${String(index)}`,
            originalText: line.text,
          }))
        : [],
    [teiContents],
  )
  const zoneToServerTextBlockId = useMemo(
    () => (teiContents ? getTeiZoneToServerTextBlockId(teiContents) : {}),
    [teiContents],
  )
  const [teiViewModes, setTeiViewModes] = useLocalStorageState<TeiViewMode[]>(
    'teiViewModes',
    { defaultValue: [] },
  )

  const teiTranslations = useMemo<TeiTranslation[]>(
    () => (teiContents ? getTeiTranslations(teiContents) : []),
    [teiContents],
  )
  const showMinCertControl = useMemo(
    () => (teiContents ? hasTeiCertaintyDegrees(teiContents) : false),
    [teiContents],
  )

  const availableViewModes = useMemo<TeiViewMode[]>(
    () => ['original', ...teiTranslations.map((translation) => translation.id)],
    [teiTranslations],
  )

  useEffect(() => {
    if (teiContents == null) {
      return
    }
    const next = normalizeTeiViewModes(teiViewModes, availableViewModes)
    if (
      next.length === teiViewModes.length &&
      next.every((mode, index) => mode === teiViewModes[index])
    ) {
      return
    }
    setTeiViewModes(next)
  }, [availableViewModes, setTeiViewModes, teiContents, teiViewModes])

  const orderedSelectedViewModes = useMemo(() => {
    const selected = new Set(
      normalizeTeiViewModes(teiViewModes, availableViewModes),
    )
    return availableViewModes.filter((mode) => selected.has(mode))
  }, [availableViewModes, teiViewModes])

  const activeHoveredLineMatchIds = enableHoverSync ? activeLineMatchIds : []

  useEffect(() => {
    setEditableOriginalLines(baseEditableOriginalLines)
  }, [baseEditableOriginalLines])

  useEffect(() => {
    setIsTextEditMode(false)
    setTextEditError(null)
  }, [annotationId, currentPageOrKey, datasetId, effectiveTeiSource])

  useEffect(() => {
    onSurfaceZonesChange(teiContents ? getTeiSurfaceZones(teiContents) : [])
  }, [currentPageOrKey, effectiveTeiSource, onSurfaceZonesChange, teiContents])

  useEffect(
    () => () => {
      onSurfaceZonesChange([])
    },
    [onSurfaceZonesChange],
  )

  const teiCategories = useMemo(
    () => (teiContents ? getTeiHighlightCategories(teiContents) : []),
    [teiContents],
  )

  const datasetFeatures = useMemo(
    () => featuresQuery.data ?? [],
    [featuresQuery.data],
  )

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

  const resolvedTeiFeatures = useMemo<ResolvedTeiFeature[]>(() => {
    const byId = new Map<string, ResolvedTeiFeature>()
    for (const category of teiCategories) {
      const matched = matchTeiCategoryToFeature(
        category.id,
        category.label,
        datasetFeatures,
      )
      const featureId = matched?.id || category.id
      if (byId.has(featureId)) {
        continue
      }
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
    return [...byId.values()].sort((left, right) =>
      left.label.localeCompare(right.label, undefined, { sensitivity: 'base' }),
    )
  }, [datasetFeatures, teiCategories])

  const currentFeatureResults = useMemo(
    () => featureResultsQuery.data || [],
    [featureResultsQuery.data],
  )
  const baselineResultsByFeature = useMemo(() => {
    const map: Record<string, feature_Result> = {}
    for (const result of currentFeatureResults) {
      if (!result.feature_id) {
        continue
      }
      map[result.feature_id] = result
    }
    return map
  }, [currentFeatureResults])

  useEffect(() => {
    const next = toDraftHighlightsFromResults(currentFeatureResults)
    let cancelled = false
    queueMicrotask(() => {
      if (cancelled) {
        return
      }
      setBaselineHighlights(next)
      setDraftHighlights(next)
      setRemovedTeiHighlightIds([])
      setForcedChangedFeatureIds([])
      setFeatureModalState(null)
      setModalFeatureId('')
      setSaveError(null)
    })
    return () => {
      cancelled = true
    }
  }, [
    annotationId,
    currentPageOrKey,
    currentFeatureResults,
    datasetId,
    effectiveTeiSource,
  ])

  const allResolvedFeatures = useMemo<ResolvedTeiFeature[]>(() => {
    const byId = new Map<string, ResolvedTeiFeature>()
    for (const feature of datasetFeatures) {
      if (!feature.id) continue
      byId.set(feature.id, {
        id: feature.id,
        label: feature.name?.trim() || feature.id,
        description: feature.description?.trim() || '',
        color: feature.color || '#f2f2f2',
        isDefault: !!feature.is_default,
        renderMode: isVerbLike(feature.id, feature.name) ? 'outline' : 'fill',
      })
    }
    for (const feature of resolvedTeiFeatures) {
      if (!byId.has(feature.id)) {
        byId.set(feature.id, feature)
      }
    }
    return [...byId.values()].sort((left, right) =>
      left.label.localeCompare(right.label, undefined, { sensitivity: 'base' }),
    )
  }, [datasetFeatures, resolvedTeiFeatures])

  const highlightStorageKey = datasetId
    ? `teiVisibleHighlightFeatures:${datasetId}`
    : 'teiVisibleHighlightFeatures'
  const [storedVisibleFeatureIds, setStoredVisibleFeatureIds] =
    useLocalStorageState<string[] | null>(highlightStorageKey, {
      defaultValue: null,
    })

  const visibleFeatureIds = useMemo(() => {
    const availableIds = allResolvedFeatures.map((feature) => feature.id)
    if (!availableIds.length) {
      return []
    }

    const availableSet = new Set(availableIds)
    const defaultIds = allResolvedFeatures
      .filter((feature) => feature.isDefault)
      .map((feature) => feature.id)

    const order = (ids: string[]) =>
      availableIds.filter((id) => ids.includes(id))

    if (storedVisibleFeatureIds === null) {
      return order(defaultIds)
    }

    const filtered = order(
      storedVisibleFeatureIds.filter((id) => availableSet.has(id)),
    )

    if (storedVisibleFeatureIds.length === 0) {
      return []
    }

    return filtered
  }, [allResolvedFeatures, storedVisibleFeatureIds])

  useEffect(() => {
    if (!allResolvedFeatures.length) {
      if (storedVisibleFeatureIds !== null) {
        setStoredVisibleFeatureIds(null)
      }
      return
    }

    if (!sameStringArray(storedVisibleFeatureIds, visibleFeatureIds)) {
      setStoredVisibleFeatureIds(visibleFeatureIds)
    }
  }, [
    allResolvedFeatures,
    setStoredVisibleFeatureIds,
    storedVisibleFeatureIds,
    visibleFeatureIds,
  ])

  const highlightConfig = useMemo<TeiHighlightConfig | undefined>(() => {
    if (!allResolvedFeatures.length) {
      return undefined
    }

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

    const manualHighlights: TeiManualHighlight[] = draftHighlights
      .filter((highlight) => highlight.sourceId.startsWith('manual'))
      .map((highlight) => ({
        id: highlight.localId,
        paragraphIndex: highlight.paragraphIndex,
        start: highlight.start,
        end: highlight.end,
        featureId: highlight.featureId,
        surface: highlight.surface,
        normalized: highlight.normalized,
        institution: highlight.institution,
        ancientPersona: highlight.ancientPersona,
      }))

    const draftFeatureIds = [
      ...new Set(draftHighlights.map((h) => h.featureId)),
    ]
    const selectedCategoryIds = [
      ...new Set([...visibleFeatureIds, ...draftFeatureIds]),
    ]
    const draftLocalIds = new Set(
      draftHighlights.map((highlight) => highlight.localId),
    )
    const hiddenFromDraft = baselineHighlights
      .filter(
        (highlight) =>
          !highlight.sourceId.startsWith('manual') &&
          !draftLocalIds.has(highlight.localId),
      )
      .map((highlight) => highlight.sourceId)
    const hiddenTeiHighlightIds = [
      ...new Set([...hiddenFromDraft, ...removedTeiHighlightIds]),
    ]

    return {
      selectedCategoryIds,
      categoryConfigById,
      categoryToFeatureId,
      manualHighlights,
      hiddenTeiHighlightIds,
    }
  }, [
    allResolvedFeatures,
    baselineHighlights,
    categoryToFeatureId,
    draftHighlights,
    removedTeiHighlightIds,
    visibleFeatureIds,
  ])

  const removeHighlightFromTooltip = (item: TeiTooltipItem) => {
    if (isTextEditMode) {
      return
    }
    debugTeiHighlights('remove-click', {
      id: item.id,
      featureId: item.featureId,
      paragraphIndex: item.paragraphIndex,
      start: item.start,
      end: item.end,
      surface: item.surface,
    })
    setRemovedTeiHighlightIds((previous) =>
      previous.includes(item.id) ? previous : [...previous, item.id],
    )
    setDraftHighlights((previous) => {
      const isMatch = (highlight: DraftHighlight) =>
        highlight.localId === item.id ||
        highlight.sourceId === item.id ||
        (highlight.featureId === item.featureId &&
          highlight.paragraphIndex === item.paragraphIndex &&
          highlight.start === item.start &&
          highlight.end === item.end) ||
        (highlight.featureId === item.featureId &&
          highlight.paragraphIndex === item.paragraphIndex &&
          highlight.surface.trim() !== '' &&
          item.surface.trim() !== '' &&
          highlight.surface.trim() === item.surface.trim()) ||
        (highlight.paragraphIndex === item.paragraphIndex &&
          highlight.start === item.start &&
          highlight.end === item.end) ||
        (highlight.featureId === item.featureId &&
          highlight.paragraphIndex === item.paragraphIndex &&
          Math.max(highlight.start, item.start) <
            Math.min(highlight.end, item.end))

      const matchedIndexes: number[] = []
      for (let index = 0; index < previous.length; index++) {
        if (isMatch(previous[index])) {
          matchedIndexes.push(index)
        }
      }
      debugTeiHighlights('remove-match', {
        draftCountBefore: previous.length,
        matchedIndexes,
      })

      if (matchedIndexes.length > 0) {
        const next = previous.filter(
          (_, index) => !matchedIndexes.includes(index),
        )
        debugTeiHighlights('remove-result', {
          draftCountAfter: next.length,
          strategy: 'matchedIndexes',
        })
        return next
      }

      const fallbackIndex = previous.findIndex(
        (highlight) => highlight.featureId === item.featureId,
      )
      if (fallbackIndex >= 0) {
        const next = previous.filter((_, index) => index !== fallbackIndex)
        debugTeiHighlights('remove-result', {
          draftCountAfter: next.length,
          strategy: 'fallbackFeatureId',
          fallbackIndex,
        })
        return next
      }

      debugTeiHighlights('remove-result', {
        draftCountAfter: previous.length,
        strategy: 'no-op',
      })
      return previous
    })
    setForcedChangedFeatureIds((previous) =>
      previous.includes(item.featureId)
        ? previous
        : [...previous, item.featureId],
    )
  }

  const allFeatureOptions = useMemo<FeatureOption[]>(
    () =>
      allResolvedFeatures.map((feature) => ({
        value: feature.id,
        label: feature.label,
        color: feature.color,
        description: feature.description,
      })),
    [allResolvedFeatures],
  )

  const selectedFeatureOptions = useMemo(
    () =>
      allFeatureOptions.filter((option) =>
        visibleFeatureIds.includes(option.value),
      ),
    [allFeatureOptions, visibleFeatureIds],
  )

  const changedFeatureIds = useMemo(() => {
    const baselineByFeature = groupByFeature(baselineHighlights)
    const draftByFeature = groupByFeature(draftHighlights)
    const featureIds = new Set([
      ...Object.keys(baselineByFeature),
      ...Object.keys(draftByFeature),
    ])

    const changedByDiff = [...featureIds].filter((featureId) => {
      const baselineValues = getComparableValues(
        toResultValues(baselineByFeature[featureId] || []),
      )
      const draftValues = getComparableValues(
        toResultValues(draftByFeature[featureId] || []),
      )
      return JSON.stringify(baselineValues) !== JSON.stringify(draftValues)
    })
    return [...new Set([...changedByDiff, ...forcedChangedFeatureIds])]
  }, [baselineHighlights, draftHighlights, forcedChangedFeatureIds])

  useEffect(() => {
    debugTeiHighlights('dirty-state', {
      baselineHighlights: baselineHighlights.length,
      draftHighlights: draftHighlights.length,
      changedFeatureIds,
      removedTeiHighlightIds,
      forcedChangedFeatureIds,
    })
  }, [
    baselineHighlights.length,
    changedFeatureIds,
    draftHighlights.length,
    forcedChangedFeatureIds,
    removedTeiHighlightIds,
  ])

  const unsavedFeatureCount = changedFeatureIds.length
  const hasUnsavedChanges = unsavedFeatureCount > 0
  const unsavedTextLineCount = useMemo(
    () =>
      editableOriginalLines.filter((line) => line.text !== line.originalText)
        .length,
    [editableOriginalLines],
  )
  const hasUnsavedTextChanges = unsavedTextLineCount > 0

  const showPane = ocred || !!editionId
  const annotationSourceFailed =
    ocred && annotationTeiQuery.isError && !annotationTeiQuery.data
  const editionSourceFailed =
    !!editionId && editionTeiQuery.isError && !editionTeiQuery.data
  const allCandidateSourcesFailed =
    (ocred ? annotationSourceFailed : true) &&
    (editionId ? editionSourceFailed : true)
  const isFetchingCandidateSource =
    (ocred && annotationTeiQuery.isFetching) ||
    (!!editionId && editionTeiQuery.isFetching)

  const saveMutation = useMutation({
    mutationFn: (results: feature_Result[]) =>
      FeatureResultsService.postDatasetsAnnotationsResults({
        dataSetId: datasetId,
        id: annotationId,
        result: results,
      }),
  })
  const textEditMutation = useMutation({
    mutationFn: (payload: annotationrule_TextBlockCorrections) =>
      AnnotationsApplyRulesService.putDatasetsAnnotationsApplyTextBlockCorrections(
        {
          dataSetId: datasetId,
          id: annotationId,
          annotationTextBlockCorrections: payload,
        },
      ),
  })

  const handleStartTextEdit = () => {
    if (hasUnsavedChanges || !isAuthenticated || !teiContents) {
      return
    }
    if (!orderedSelectedViewModes.includes('original')) {
      setTeiViewModes(
        normalizeTeiViewModes(
          ['original', ...teiViewModes],
          availableViewModes,
        ),
      )
    }
    setEditableOriginalLines(baseEditableOriginalLines)
    setFeatureModalState(null)
    setModalFeatureId('')
    setTextEditError(null)
    setIsTextEditMode(true)
  }

  const handleCancelTextEdit = () => {
    setEditableOriginalLines(baseEditableOriginalLines)
    setTextEditError(null)
    setIsTextEditMode(false)
  }

  const handleSaveTextEdit = async () => {
    if (!datasetId || !annotationId || !isAuthenticated || !teiContents) {
      return
    }

    const grouped = new Map<string, { old: string[]; correction: string[] }>()
    for (const line of editableOriginalLines) {
      if (line.text === line.originalText) {
        continue
      }
      const current = grouped.get(line.blockId) || {
        old: [],
        correction: [],
      }
      current.old.push(line.originalText)
      current.correction.push(line.text)
      grouped.set(line.blockId, current)
    }

    if (!grouped.size) {
      setIsTextEditMode(false)
      setTextEditError(null)
      return
    }

    const parsedPage = Number.parseInt(String(currentPageOrKey), 10)
    const payload: annotationrule_TextBlockCorrections = {
      type: 'text_blocks_corrections',
      corrections: [...grouped.entries()].map(([blockId, value]) => ({
        text_block_id: zoneToServerTextBlockId[blockId] || blockId,
        old: value.old,
        correction: value.correction,
        page: Number.isFinite(parsedPage) ? parsedPage : undefined,
      })),
    }

    try {
      setTextEditError(null)
      await textEditMutation.mutateAsync(payload)
      setIsTextEditMode(false)
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: annotationTeiQueryKey(
            datasetId,
            annotationId,
            currentPageOrKey,
          ),
        }),
        queryClient.invalidateQueries({
          queryKey: [
            'featureResults',
            datasetId,
            annotationId,
            currentPageOrKey,
          ],
        }),
        editionId
          ? queryClient.invalidateQueries({
              queryKey: editionTeiQueryKey(editionId, currentPageOrKey),
            })
          : Promise.resolve(),
      ])
    } catch (error) {
      setTextEditError(
        error instanceof Error ? error.message : 'Failed to save text edits.',
      )
    }
  }

  const handleSave = async () => {
    if (!datasetId || !annotationId || !isAuthenticated || isTextEditMode) {
      return
    }

    const draftByFeature = groupByFeature(draftHighlights)
    const payloads: feature_Result[] = changedFeatureIds.map((featureId) => {
      const existing = baselineResultsByFeature[featureId]
      return {
        ...(existing || {}),
        dataset_id: datasetId,
        annotation_id: annotationId,
        feature_id: featureId,
        page_key: existing?.page_key || String(currentPageOrKey),
        values: toResultValues(draftByFeature[featureId] || []),
      }
    })

    if (!payloads.length) {
      return
    }

    try {
      setSaveError(null)
      await saveMutation.mutateAsync(payloads)
      setBaselineHighlights(draftHighlights)
      await featureResultsQuery.refetch()
      await queryClient.invalidateQueries({
        queryKey: annotationTeiQueryKey(
          datasetId,
          annotationId,
          currentPageOrKey,
        ),
      })
      if (editionId) {
        await queryClient.invalidateQueries({
          queryKey: editionTeiQueryKey(editionId, currentPageOrKey),
        })
      }
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : 'Failed to save highlights.',
      )
    }
  }

  const addHighlight = (selection: SelectionDraft, featureId: string) => {
    const localId = [
      'manual',
      featureId,
      selection.paragraphIndex,
      selection.start,
      selection.end,
      Date.now(),
      Math.random().toString(36).slice(2, 8),
    ].join(':')

    const next: DraftHighlight = {
      localId,
      sourceId: localId,
      paragraphIndex: selection.paragraphIndex,
      start: selection.start,
      end: selection.end,
      featureId,
      categoryId: featureId,
      surface: selection.surface,
      normalized: '',
      institution: '',
      ancientPersona: '',
      fromAnchorId: '',
      toAnchorId: '',
    }

    setDraftHighlights((previous) => [...previous, next])
  }

  const openModalForAdd = (selection: SelectionDraft) => {
    if (isTextEditMode || allFeatureOptions.length === 0) {
      return
    }
    setFeatureModalState({ selection })
    setModalFeatureId(allFeatureOptions[0]?.value || '')
  }

  const submitFeatureModal = () => {
    if (!featureModalState || !modalFeatureId || isTextEditMode) {
      return
    }

    addHighlight(featureModalState.selection, modalFeatureId)

    setFeatureModalState(null)
    setModalFeatureId('')
  }

  const closeFeatureModal = () => {
    setFeatureModalState(null)
    setModalFeatureId('')
  }

  if (!showPane || allCandidateSourcesFailed) {
    return null
  }

  if (!teiContents && isFetchingCandidateSource) {
    return null
  }

  const sourceOptions: SourceOption[] = availableSources.map((source) => ({
    value: source,
    label: source === 'edition' ? 'Edition' : 'Annotation',
  }))
  const selectedSourceOption =
    sourceOptions.find((option) => option.value === effectiveTeiSource) || null
  const getViewModeLabel = (mode: TeiViewMode) => {
    if (mode === 'original') {
      return 'Original'
    }
    const rawLabel =
      teiTranslations.find((translation) => translation.id === mode)?.label ||
      mode
    return VIEW_LABEL_MAP[rawLabel] || rawLabel
  }

  const isLoading =
    (effectiveTeiSource === 'edition' && editionTeiQuery.isFetching) ||
    (effectiveTeiSource === 'annotation' && annotationTeiQuery.isFetching)
  const error =
    (effectiveTeiSource === 'edition' && editionTeiQuery.error) ||
    (effectiveTeiSource === 'annotation' && annotationTeiQuery.error)

  const paneTitle = hasUnsavedChanges
    ? `Contents (${unsavedFeatureCount} unsaved feature${unsavedFeatureCount === 1 ? '' : 's'})`
    : isTextEditMode && hasUnsavedTextChanges
      ? `Contents (${unsavedTextLineCount} unsaved text line${unsavedTextLineCount === 1 ? '' : 's'})`
      : 'Contents'

  const modalFeatureOption =
    allFeatureOptions.find((option) => option.value === modalFeatureId) || null
  const canStartTextEdit =
    isAuthenticated &&
    !!teiContents &&
    baseEditableOriginalLines.length > 0 &&
    !hasUnsavedChanges &&
    !isTextEditMode

  return (
    <>
      <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 h-full bg-white">
        <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
          <div>{paneTitle}</div>
          <div className="flex items-center gap-2">
            {textEditError && (
              <span className="text-xs text-red-600">{textEditError}</span>
            )}
            {saveError && (
              <span className="text-xs text-red-600">{saveError}</span>
            )}
            {isAuthenticated && isTextEditMode && (
              <>
                <button
                  type="button"
                  className="px-2 py-1 rounded border border-gray-300 text-gray-700 bg-white hover:bg-gray-50 text-xs"
                  onClick={handleCancelTextEdit}
                  disabled={textEditMutation.isPending}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="px-2 py-1 rounded border border-teal-300 text-teal-700 bg-white hover:bg-teal-50 text-xs"
                  onClick={() => {
                    void handleSaveTextEdit()
                  }}
                  disabled={
                    textEditMutation.isPending || !hasUnsavedTextChanges
                  }
                >
                  {textEditMutation.isPending ? 'Saving…' : 'Save'}
                </button>
              </>
            )}
            {isAuthenticated && hasUnsavedChanges && !isTextEditMode && (
              <button
                type="button"
                className="px-2 py-1 rounded border border-teal-300 text-teal-700 bg-white hover:bg-teal-50 text-xs"
                onClick={() => {
                  void handleSave()
                }}
                disabled={saveMutation.isPending}
              >
                {saveMutation.isPending ? 'Saving…' : 'Save'}
              </button>
            )}
            {!isTextEditMode && !hasUnsavedChanges && (
              <button
                type="button"
                className="px-2 py-1 rounded border border-teal-300 text-teal-700 bg-white hover:bg-teal-50 text-xs disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={handleStartTextEdit}
                disabled={!canStartTextEdit}
              >
                Edit transcription
              </button>
            )}
          </div>
        </div>

        <div className="flex-1 min-h-0 overflow-hidden p-2.5 box-border flex flex-col">
          <div className="flex gap-2 items-center flex-wrap mb-2.5">
            {sourceOptions.length > 1 && (
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium text-gray-600">
                  Source:
                </span>
                <div className="w-36">
                  <Select
                    value={selectedSourceOption}
                    onChange={(option: SourceOption | null) => {
                      if (option) {
                        setStoredTeiSource(option.value)
                      }
                    }}
                    options={sourceOptions}
                    isClearable={false}
                    isDisabled={isTextEditMode}
                    styles={selectStyles<SourceOption>()}
                    menuPortalTarget={document.body}
                    menuPosition="fixed"
                  />
                </div>
              </div>
            )}
            {availableViewModes.length > 1 && (
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium text-gray-600">
                  Views:
                </span>
                <MultiSelectDropdown<TeiViewMode>
                  allItems={availableViewModes}
                  selectedItems={orderedSelectedViewModes}
                  setSelectedItems={(items) => {
                    if (isTextEditMode) {
                      return
                    }
                    if (!items || items.length === 0) {
                      setTeiViewModes(['original'])
                      return
                    }
                    setTeiViewModes(items)
                  }}
                  itemsLabel="views"
                  getItemLabel={getViewModeLabel}
                  minWidth="180px"
                />
              </div>
            )}
            <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
              <input
                type="checkbox"
                checked={alignLines}
                onChange={(e) => setAlignLines(e.target.checked)}
                disabled={isTextEditMode}
                className="rounded border-gray-300"
              />
              <span>Align lines</span>
            </label>
            <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
              <input
                type="checkbox"
                checked={enableHoverSync}
                onChange={(event) =>
                  onEnableHoverSyncChange(event.target.checked)
                }
                className="rounded border-gray-300"
              />
              <span>Highlight facsimile</span>
            </label>
            {showMinCertControl && (
              <RangeInput
                label="Min certainty"
                value={minCert}
                min={0.8}
                max={1}
                step={0.001}
                title="Hide tokens below certainty threshold"
                onChange={(value) =>
                  setMinCert(Math.round(value * 1000) / 1000)
                }
              />
            )}
            <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
              <input
                type="checkbox"
                checked={showTeiSource}
                onChange={(event) => setShowTeiSource(event.target.checked)}
                disabled={isTextEditMode}
                className="rounded border-gray-300"
              />
              <span>TEI source code</span>
            </label>
            {allFeatureOptions.length > 0 && (
              <>
                <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
                  <input
                    type="checkbox"
                    checked={isFeatureSelectExpanded}
                    onChange={(event) =>
                      setIsFeatureSelectExpanded(event.target.checked)
                    }
                    disabled={isTextEditMode}
                    className="rounded border-gray-300"
                  />
                  <span>Features select</span>
                </label>
                {isFeatureSelectExpanded && (
                  <div className="flex items-center gap-1.5 min-w-65">
                    <Select<FeatureOption, true>
                      isMulti
                      value={selectedFeatureOptions}
                      onChange={(options) => {
                        const values = (options || []).map(
                          (option) => option.value,
                        )
                        setStoredVisibleFeatureIds(values)
                      }}
                      options={allFeatureOptions}
                      closeMenuOnSelect={false}
                      hideSelectedOptions={false}
                      isLoading={featuresQuery.isLoading}
                      isDisabled={isTextEditMode}
                      placeholder="Select features"
                      styles={featureSelectStyles}
                      menuPortalTarget={document.body}
                      menuPosition="fixed"
                      formatOptionLabel={(option, { context }) =>
                        formatFeatureOptionLabel(option, context)
                      }
                    />
                  </div>
                )}
              </>
            )}
          </div>

          {isLoading && !teiContents && (
            <p className="text-gray-500 text-sm py-2">Loading TEI…</p>
          )}
          {effectiveTeiSource === 'annotation' &&
            !teiContents &&
            !annotationTeiQuery.isFetching &&
            !annotationTeiQuery.error &&
            (!datasetId || !annotationId) && (
              <p className="text-amber-700 text-sm py-2">
                Select a dataset and an annotation to view annotation TEI.
              </p>
            )}
          {error && !teiContents && (
            <p className="text-red-600 text-sm py-2">
              {effectiveTeiSource === 'edition' &&
              (error as Error)?.message?.includes('404')
                ? 'Edition TEI is not available for this page. Use annotation TEI or another source.'
                : 'Failed to load TEI. Try switching source.'}
            </p>
          )}
          {!isAuthenticated && teiContents && (
            <p className="text-gray-600 text-xs py-1">
              Log in to add, edit, remove, and save highlights.
            </p>
          )}
          {teiContents && showTeiSource && (
            <textarea
              className={`w-full mt-4 h-36 box-border resize-y border border-gray-300 rounded-lg p-2.5 outline-none font-mono text-xs leading-snug ${!showTeiSource ? 'hidden' : ''}`}
              spellCheck={false}
              placeholder="TEI XML..."
              value={teiContents || ''}
              readOnly
            />
          )}
          {teiContents && (
            <div className="mt-4 flex-1 min-h-0 overflow-y-auto overscroll-none">
              <div className="flex flex-wrap gap-3">
                {orderedSelectedViewModes.map((viewMode) => (
                  <div
                    key={`${viewMode}:${effectiveTeiSource}:${String(currentPageOrKey)}`}
                    className="min-w-105 basis-105 flex-1"
                  >
                    {isTextEditMode && viewMode === 'original' ? (
                      <OriginalViewEditor
                        lines={editableOriginalLines}
                        showViewLabel={availableViewModes.length > 1}
                        onChangeLine={(lineId, text) => {
                          setEditableOriginalLines((previous) =>
                            previous.map((line) =>
                              line.id === lineId ? { ...line, text } : line,
                            ),
                          )
                        }}
                      />
                    ) : (
                      <Tei
                        data={teiContents}
                        minCert={minCert}
                        viewMode={viewMode}
                        viewLabel={getViewModeLabel(viewMode)}
                        showViewLabel={availableViewModes.length > 1}
                        alignLines={alignLines}
                        highlightConfig={highlightConfig}
                        editable={isAuthenticated && !isTextEditMode}
                        canAddHighlight={allFeatureOptions.length > 0}
                        activeLineMatchIds={activeHoveredLineMatchIds}
                        enableHoverSync={enableHoverSync}
                        onHoverLineMatchIds={onHoverLineMatchIds}
                        onRequestAddHighlight={openModalForAdd}
                        onRequestRemoveHighlight={removeHighlightFromTooltip}
                      />
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </section>

      {featureModalState &&
        !isTextEditMode &&
        createPortal(
          <div className="fixed inset-0 z-[12500] flex items-center justify-center bg-black/40 p-4">
            <div className="w-full max-w-md rounded-xl bg-white border border-gray-200 p-4 shadow-xl">
              <div className="text-sm font-semibold text-gray-900 mb-2">
                Highlight a feature
              </div>
              <div className="text-xs text-gray-600 mb-2">
                "{featureModalState.selection.surface}"
              </div>
              <Select<FeatureOption, false>
                value={modalFeatureOption}
                onChange={(option) => {
                  setModalFeatureId(option?.value || '')
                }}
                options={allFeatureOptions}
                isClearable={false}
                styles={featureModalStyles}
                menuPortalTarget={document.body}
                menuPosition="fixed"
                formatOptionLabel={(option, { context }) =>
                  formatFeatureOptionLabel(option, context)
                }
              />
              <div className="mt-4 flex items-center justify-end gap-2">
                <button
                  type="button"
                  className="px-3 py-1.5 rounded border border-gray-300 text-gray-700 hover:bg-gray-50 text-sm"
                  onClick={closeFeatureModal}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="px-3 py-1.5 rounded border border-teal-300 text-teal-700 hover:bg-teal-50 text-sm"
                  onClick={submitFeatureModal}
                  disabled={!modalFeatureId}
                >
                  Apply
                </button>
              </div>
            </div>
          </div>,
          document.body,
        )}
    </>
  )
}
