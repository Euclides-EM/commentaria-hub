import type { TeiSurfaceZone } from './contents/tei/tei.ts'

export type HighlightZoneFilter = 'line' | 'block'

export const HIGHLIGHT_ZONE_FILTER_OPTIONS: HighlightZoneFilter[] = [
  'line',
  'block',
]

export const DEFAULT_HIGHLIGHT_ZONE_FILTERS: HighlightZoneFilter[] = [
  'line',
  'block',
]

export const HIGHLIGHT_ZONE_FILTER_STORAGE_KEY = 'imagePaneHighlightZoneFilters'

export const getHighlightZoneFilterLabel = (value: HighlightZoneFilter) =>
  value === 'line' ? 'Lines' : 'Regions'

export const getHighlightZoneFilterPickerLabel = ({
  allItems,
  selectedItems,
}: {
  allItems: HighlightZoneFilter[]
  selectedItems: HighlightZoneFilter[] | null
}) => {
  const items = selectedItems == null ? allItems : selectedItems
  if (items.length === 0) {
    return '0 selected'
  }
  return items.map(getHighlightZoneFilterLabel).join(', ')
}

const ZONE_CAT_PALETTE: Array<[string, string, string, string]> = [
  [
    'rgba(245,158,11,1)',
    'rgba(245,158,11,0.55)',
    'rgba(251,191,36,0.18)',
    'rgba(251,191,36,0.08)',
  ],
  [
    'rgba(139,92,246,1)',
    'rgba(139,92,246,0.55)',
    'rgba(167,139,250,0.18)',
    'rgba(167,139,250,0.08)',
  ],
  [
    'rgba(239,68,68,1)',
    'rgba(239,68,68,0.55)',
    'rgba(252,165,165,0.18)',
    'rgba(252,165,165,0.08)',
  ],
  [
    'rgba(6,182,212,1)',
    'rgba(6,182,212,0.55)',
    'rgba(103,232,249,0.18)',
    'rgba(103,232,249,0.08)',
  ],
  [
    'rgba(59,130,246,1)',
    'rgba(59,130,246,0.55)',
    'rgba(147,197,253,0.18)',
    'rgba(147,197,253,0.08)',
  ],
  [
    'rgba(236,72,153,1)',
    'rgba(236,72,153,0.55)',
    'rgba(249,168,212,0.18)',
    'rgba(249,168,212,0.08)',
  ],
  [
    'rgba(34,197,94,1)',
    'rgba(34,197,94,0.55)',
    'rgba(134,239,172,0.18)',
    'rgba(134,239,172,0.08)',
  ],
  [
    'rgba(249,115,22,1)',
    'rgba(249,115,22,0.55)',
    'rgba(253,186,116,0.18)',
    'rgba(253,186,116,0.08)',
  ],
  [
    'rgba(20,184,166,1)',
    'rgba(20,184,166,0.55)',
    'rgba(94,234,212,0.18)',
    'rgba(94,234,212,0.08)',
  ],
  [
    'rgba(234,179,8,1)',
    'rgba(234,179,8,0.55)',
    'rgba(253,224,71,0.18)',
    'rgba(253,224,71,0.08)',
  ],
]

const hashStr = (str: string) => {
  let h = 5381
  for (let i = 0; i < str.length; i++) {
    h = ((h << 5) + h) ^ str.charCodeAt(i)
  }
  return Math.abs(h)
}

export const zoneCategoryToColor = (category: string) => {
  const entry = ZONE_CAT_PALETTE[hashStr(category) % ZONE_CAT_PALETTE.length]
  return {
    activeBorder: entry[0],
    inactiveBorder: entry[1],
    activeBg: entry[2],
    inactiveBg: entry[3],
  }
}

export const zoneCategoryLabel = (category: string) =>
  category
    .replace(/^zone_cat_/i, '')
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')

export const filterSurfaceZones = (
  surfaceZones: TeiSurfaceZone[],
  selectedFilters: HighlightZoneFilter[],
) => {
  if (selectedFilters.length === 0) {
    return []
  }
  const selected = new Set(selectedFilters)
  return surfaceZones.filter((zone) => selected.has(zone.zoneType))
}
