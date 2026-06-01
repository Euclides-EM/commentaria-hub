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
