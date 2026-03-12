import type { StylesConfig } from 'react-select'
import { selectStyles } from '../../../../styles/selectStyles.ts'
import type { TeiViewMode } from './tei.ts'
import type {
  DraftHighlight,
  FeatureOption,
  ResolvedTeiFeature,
  TeiDatasetFeature,
  TeiFeatureResult,
  TeiFeatureResultValue,
  TeiTooltipItem,
} from './TeiPane.types.ts'

export const VIEW_LABEL_MAP: Record<string, string> = {
  modern_en: 'English',
  en: 'English',
}

export const normalizeTeiViewModes = (
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

export const matchTeiCategoryToFeature = (
  categoryId: string,
  categoryLabel: string,
  features: TeiDatasetFeature[],
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

export const isVerbLike = (...values: Array<string | null | undefined>) =>
  values.some((value) => normalizeMatchKey(value).includes('verb'))

export const parseLineMatchIds = (value: string | null | undefined) =>
  [
    ...new Set(
      (value || '')
        .split(/\s+/)
        .map((entry) => entry.trim())
        .filter(Boolean),
    ),
  ].sort()

export const sameStringArray = (left: string[] | null, right: string[]) => {
  if (!left) return false
  if (left.length !== right.length) return false
  return left.every((value, index) => value === right[index])
}

export const toResultValues = (
  highlights: DraftHighlight[],
): TeiFeatureResultValue[] =>
  highlights
    .map((highlight) => {
      const properties: Record<string, string> = {
        paragraph_index: String(highlight.paragraphIndex),
        start: String(highlight.start),
        end: String(highlight.end),
        category_id: highlight.categoryId || highlight.featureId,
      }
      if (highlight.sourceId && !highlight.sourceId.startsWith('manual:')) {
        properties.highlight_id = highlight.sourceId
      }
      if (highlight.fromAnchorId) {
        properties.from_anchor_id = highlight.fromAnchorId
      }
      if (highlight.toAnchorId) {
        properties.to_anchor_id = highlight.toAnchorId
      }
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

export const getComparableValues = (values: TeiFeatureResultValue[]) =>
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

export const groupByFeature = (highlights: DraftHighlight[]) => {
  const grouped: Record<string, DraftHighlight[]> = {}
  for (const highlight of highlights) {
    grouped[highlight.featureId] = grouped[highlight.featureId] || []
    grouped[highlight.featureId].push(highlight)
  }
  return grouped
}

export const hasTeiPositionProperties = (value: TeiFeatureResultValue) => {
  const properties = value.properties || {}
  const paragraphIndex = Number.parseInt(properties.paragraph_index || '', 10)
  const start = Number.parseInt(properties.start || '', 10)
  const end = Number.parseInt(properties.end || '', 10)
  return (
    !Number.isNaN(paragraphIndex) &&
    !Number.isNaN(start) &&
    !Number.isNaN(end) &&
    end > start
  )
}

export const toDraftHighlightsFromResults = (
  results: TeiFeatureResult[],
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

export const formatFeatureOptionLabel = (
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

export const featureSelectStyles: StylesConfig<FeatureOption, true> = {
  control: (base, state) => ({
    ...base,
    minHeight: 32,
    border: `1px solid ${state.isFocused ? '#14b8a6' : '#9ca3af'}`,
    borderRadius: 6,
    boxShadow: state.isFocused ? '0 0 0 3px rgba(20, 184, 166, 0.15)' : 'none',
  }),
  valueContainer: (base) => ({
    ...base,
    padding: '6px',
    gap: '6px',
    rowGap: '6px',
    columnGap: '6px',
    alignItems: 'flex-start',
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
    margin: 0,
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
  input: (base) => ({
    ...base,
    margin: 0,
    padding: 0,
  }),
}

export const featureModalStyles: StylesConfig<FeatureOption, false> = {
  ...selectStyles<FeatureOption>(),
  menuPortal: (base) => ({ ...base, zIndex: 13000 }),
}

export const removeHighlightFromDrafts = (
  previous: DraftHighlight[],
  item: TeiTooltipItem,
) => {
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
      Math.max(highlight.start, item.start) < Math.min(highlight.end, item.end))

  const matchedIndexes: number[] = []
  for (let index = 0; index < previous.length; index++) {
    if (isMatch(previous[index])) {
      matchedIndexes.push(index)
    }
  }

  if (matchedIndexes.length > 0) {
    return previous.filter((_, index) => !matchedIndexes.includes(index))
  }

  const fallbackIndex = previous.findIndex(
    (highlight) => highlight.featureId === item.featureId,
  )
  if (fallbackIndex >= 0) {
    return previous.filter((_, index) => index !== fallbackIndex)
  }

  return previous
}

export const toFeatureOptions = (
  features: ResolvedTeiFeature[],
): FeatureOption[] =>
  features.map((feature) => ({
    value: feature.id,
    label: feature.label,
    color: feature.color,
    description: feature.description,
  }))
