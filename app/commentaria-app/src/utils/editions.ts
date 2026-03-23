export const TITLE_PAGES_DATASET_ID = 'tps'

export interface EditionDisplayInfo {
  key?: string | null
  year?: string | number | null
  editors?: Array<string | { name?: string | null }> | null
  cities?: Array<string | { name?: string | null }> | null
  shortTitle?: string | null
  title?: string | null
}

const normalizeText = (value: string | { name?: string | null }): string => {
  if (typeof value === 'string') return value.trim()
  return String(value.name ?? '').trim()
}

export const formatEditionLabel = (item: EditionDisplayInfo) => {
  const details = [
    item.year != null ? String(item.year) : null,
    (item.cities ?? []).map((value) => normalizeText(value)).join(', '),
    (item.editors ?? []).map((value) => normalizeText(value)).join(', '),
  ]
    .filter(Boolean)
    .join(', ')
  let title = item.shortTitle || item.title || ''
  title = title.length > 128 ? title.slice(0, 125) + '...' : title
  if (title === '?') {
    title = ''
  }
  return title ? `${details} - ${title}` : details
}
