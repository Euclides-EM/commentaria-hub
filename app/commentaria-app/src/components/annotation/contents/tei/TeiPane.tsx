import {
  getTeiHighlightCategories,
  hasTeiCertaintyDegrees,
  getTeiTranslations,
  teiToHtml,
  type TeiHighlightConfig,
  type TeiTranslation,
  type TeiViewMode,
} from './tei.ts'
import { useAppState } from '../../../../context/useAppState.ts'
import { useEffect, useMemo, useRef, useState } from 'react'
import {
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
import type { feature_Feature } from '@hub-api'

const VIEW_LABEL_MAP: Record<string, string> = {
  modern_en: 'English',
}

const FALLBACK_COLORS = [
  '#FDE68A',
  '#BFDBFE',
  '#FBCFE8',
  '#C7D2FE',
  '#A7F3D0',
  '#FED7AA',
  '#DDD6FE',
  '#99F6E4',
]

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
    return ['original']
  }
  return next
}

const normalizeMatchKey = (value: string | null | undefined) =>
  (value || '').toLowerCase().replace(/[^a-z0-9]+/g, '')

const fallbackColorForId = (id: string) => {
  let hash = 0
  for (const ch of id) {
    hash = (hash * 31 + ch.charCodeAt(0)) >>> 0
  }
  return FALLBACK_COLORS[hash % FALLBACK_COLORS.length]
}

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
  categoryId: string
  label: string
  description: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
  color: string
}

type TeiTooltipState = {
  x: number
  y: number
  items: TeiTooltipItem[]
}

const TEI_HIGHLIGHT_SELECTOR = '[data-tei-highlight="true"]'

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
          categoryId: String(value.categoryId || ''),
          label: String(value.label || ''),
          description: String(value.description || ''),
          surface: String(value.surface || ''),
          normalized: String(value.normalized || ''),
          institution: String(value.institution || ''),
          ancientPersona: String(value.ancientPersona || ''),
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
    .map((item) => item.id)
    .sort()
    .join('|')

type Props = {
  data: string
  minCert: number
  viewMode: TeiViewMode
  viewLabel: string
  showViewLabel: boolean
  alignLines: boolean
  highlightConfig?: TeiHighlightConfig
}

const Tei = ({
  minCert,
  data,
  viewMode,
  viewLabel,
  showViewLabel,
  alignLines,
  highlightConfig,
}: Props) => {
  const { searchResultHighlight } = useAppState()
  const [tooltipState, setTooltipState] = useState<TeiTooltipState | null>(null)
  const hideTooltipTimerRef = useRef<number | null>(null)
  const tooltipHoveredRef = useRef(false)

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
              key={`${item.id}:${item.categoryId}`}
              className="flex flex-col"
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
            </div>
          ))}
        </div>
      </div>,
      document.body,
    )

  return (
    <div
      className="relative"
      onMouseMove={(event) => {
        clearHideTooltipTimer()
        const elements = document.elementsFromPoint(
          event.clientX,
          event.clientY,
        )
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
        setTooltipState((previous) => {
          if (!previous) {
            return { x: position.x, y: position.y, items }
          }
          const previousKey = getTooltipItemsKey(previous.items)
          const nextKey = getTooltipItemsKey(items)
          if (previousKey !== nextKey) {
            return previous
          }
          return { x: position.x, y: position.y, items }
        })
      }}
      onMouseLeave={() => {
        if (tooltipHoveredRef.current) return
        scheduleHideTooltip()
      }}
    >
      {showViewLabel && (
        <div className="absolute top-2 right-2 z-10 rounded bg-white/90 border border-gray-300 px-1.5 py-0.5 text-[10px] font-medium text-gray-700">
          {viewLabel}
        </div>
      )}
      <div
        className={`text-xs leading-relaxed border border-gray-300 rounded-xl bg-gray-50 p-2 ${showViewLabel ? 'pt-7' : ''} [&_p]:mb-2 [&_p:last-child]:mb-0 [&_[data-tei-selected='true']]:bg-yellow-200/70 [&_[data-tei-selected='true']]:text-gray-900 [&_[data-tei-selected='true']]:rounded-sm [&_[data-tei-selected='true']]:px-0.5`}
        style={{ whiteSpace: 'normal' }}
        dangerouslySetInnerHTML={{ __html: html }}
      />
      {tooltip}
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
  }),
  multiValueLabel: (base) => ({
    ...base,
    color: '#111827',
    fontWeight: 600,
    maxWidth: 240,
  }),
  multiValueRemove: (base) => ({
    ...base,
    color: '#111827',
    ':hover': { backgroundColor: 'rgba(255,255,255,0.6)', color: '#111827' },
  }),
}

const sameStringArray = (left: string[] | null, right: string[]) => {
  if (!left) return false
  if (left.length !== right.length) return false
  return left.every((value, index) => value === right[index])
}

export function TeiPane() {
  const {
    annotation,
    dataset,
    state: { datasetId, annotationId, currentPageOrKey },
  } = useAppState()
  const [showTeiSource, setShowTeiSource] = useLocalStorageState(
    'showTeiSource',
    { defaultValue: false },
  )
  const [minCert, setMinCert] = useLocalStorageState('minCert', {
    defaultValue: 0.8,
  })
  const [alignLines, setAlignLines] = useLocalStorageState('alignTeiLines', {
    defaultValue: false,
  })
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
  >('teiSource', { defaultValue: 'annotation' })
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
  const [teiViewModes, setTeiViewModes] = useLocalStorageState<TeiViewMode[]>(
    'teiViewModes',
    { defaultValue: ['original'] },
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

  const teiCategories = useMemo(
    () => (teiContents ? getTeiHighlightCategories(teiContents) : []),
    [teiContents],
  )

  const resolvedTeiFeatures = useMemo<ResolvedTeiFeature[]>(() => {
    const features = featuresQuery.data ?? []
    return teiCategories.map((category) => {
      const matched = matchTeiCategoryToFeature(
        category.id,
        category.label,
        features,
      )
      return {
        id: category.id,
        label: matched?.name?.trim() || category.label,
        description: matched?.description?.trim() || '',
        color: matched?.color || fallbackColorForId(category.id),
        isDefault: !!matched?.is_default,
        renderMode: isVerbLike(
          matched?.id,
          matched?.name,
          category.id,
          category.label,
        )
          ? 'outline'
          : 'fill',
      }
    })
  }, [featuresQuery.data, teiCategories])

  const highlightStorageKey = datasetId
    ? `teiVisibleHighlightFeatures:${datasetId}`
    : 'teiVisibleHighlightFeatures'
  const [storedVisibleFeatureIds, setStoredVisibleFeatureIds] =
    useLocalStorageState<string[] | null>(highlightStorageKey, {
      defaultValue: null,
    })

  const visibleFeatureIds = useMemo(() => {
    const availableIds = resolvedTeiFeatures.map((feature) => feature.id)
    if (!availableIds.length) {
      return []
    }

    const availableSet = new Set(availableIds)
    const defaultIds = resolvedTeiFeatures
      .filter((feature) => feature.isDefault)
      .map((feature) => feature.id)

    const order = (ids: string[]) =>
      availableIds.filter((id) => ids.includes(id))

    if (storedVisibleFeatureIds === null) {
      return order(defaultIds.length > 0 ? defaultIds : availableIds)
    }

    const filtered = order(
      storedVisibleFeatureIds.filter((id) => availableSet.has(id)),
    )

    if (storedVisibleFeatureIds.length === 0) {
      return []
    }

    if (filtered.length > 0) {
      return filtered
    }

    return order(defaultIds.length > 0 ? defaultIds : availableIds)
  }, [resolvedTeiFeatures, storedVisibleFeatureIds])

  useEffect(() => {
    if (!resolvedTeiFeatures.length) {
      if (storedVisibleFeatureIds !== null) {
        setStoredVisibleFeatureIds(null)
      }
      return
    }

    if (!sameStringArray(storedVisibleFeatureIds, visibleFeatureIds)) {
      setStoredVisibleFeatureIds(visibleFeatureIds)
    }
  }, [
    resolvedTeiFeatures,
    setStoredVisibleFeatureIds,
    storedVisibleFeatureIds,
    visibleFeatureIds,
  ])

  const highlightConfig = useMemo<TeiHighlightConfig | undefined>(() => {
    if (!resolvedTeiFeatures.length) {
      return undefined
    }

    const categoryConfigById = resolvedTeiFeatures.reduce<
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
      selectedCategoryIds: visibleFeatureIds,
      categoryConfigById,
    }
  }, [resolvedTeiFeatures, visibleFeatureIds])

  const featureOptions = useMemo<FeatureOption[]>(
    () =>
      resolvedTeiFeatures.map((feature) => ({
        value: feature.id,
        label: feature.label,
        color: feature.color,
        description: feature.description,
      })),
    [resolvedTeiFeatures],
  )

  const selectedFeatureOptions = useMemo(
    () =>
      featureOptions.filter((option) =>
        visibleFeatureIds.includes(option.value),
      ),
    [featureOptions, visibleFeatureIds],
  )

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

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 h-full bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Contents</div>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden p-2.5 box-border flex flex-col">
        <div className="flex gap-2 items-center flex-wrap mb-2.5">
          {sourceOptions.length > 1 && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs font-medium text-gray-600">Source:</span>
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
                  styles={selectStyles<SourceOption>()}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                />
              </div>
            </div>
          )}
          {availableViewModes.length > 1 && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs font-medium text-gray-600">Views:</span>
              <MultiSelectDropdown<TeiViewMode>
                allItems={availableViewModes}
                selectedItems={orderedSelectedViewModes}
                setSelectedItems={(items) => {
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
              className="rounded border-gray-300"
            />
            <span>Align lines</span>
          </label>
          {showMinCertControl && (
            <RangeInput
              label="Min certainty"
              value={minCert}
              min={0.8}
              max={1}
              step={0.001}
              title="Hide tokens below certainty threshold"
              onChange={(value) => setMinCert(Math.round(value * 1000) / 1000)}
            />
          )}
          <button
            className={`px-2.5 py-1.5 border rounded-lg font-semibold text-xs ${
              showTeiSource
                ? 'bg-black text-white border-black'
                : 'border-gray-300 bg-white hover:bg-gray-50'
            }`}
            onClick={() => setShowTeiSource(!showTeiSource)}
          >
            TEI source code
          </button>
          {featureOptions.length > 0 && (
            <div className="flex items-center gap-1.5 min-w-65">
              <Select<FeatureOption, true>
                isMulti
                value={selectedFeatureOptions}
                onChange={(options) => {
                  const values = (options || []).map((option) => option.value)
                  setStoredVisibleFeatureIds(values)
                }}
                options={featureOptions}
                closeMenuOnSelect={false}
                hideSelectedOptions={false}
                isLoading={featuresQuery.isLoading}
                placeholder="Select features"
                styles={featureSelectStyles}
                menuPortalTarget={document.body}
                menuPosition="fixed"
                formatOptionLabel={(option, { context }) =>
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
                }
              />
            </div>
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
        {teiContents && showTeiSource && (
          <>
            <textarea
              className={`w-full mt-4 h-36 box-border resize-y border border-gray-300 rounded-lg p-2.5 outline-none font-mono text-xs leading-snug ${!showTeiSource ? 'hidden' : ''}`}
              spellCheck={false}
              placeholder="TEI XML..."
              value={teiContents || ''}
              readOnly
            />
          </>
        )}
        {teiContents && (
          <div className="mt-4 flex-1 min-h-0 overflow-y-auto">
            <div className="flex flex-wrap gap-3">
              {orderedSelectedViewModes.map((viewMode) => (
                <div key={viewMode} className="min-w-105 basis-105 flex-1">
                  <Tei
                    data={teiContents}
                    minCert={minCert}
                    viewMode={viewMode}
                    viewLabel={getViewModeLabel(viewMode)}
                    showViewLabel={availableViewModes.length > 1}
                    alignLines={alignLines}
                    highlightConfig={highlightConfig}
                  />
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </section>
  )
}
