export const TITLE_PAGES_DATASET_ID = 'tps'

export interface EditionDisplayInfo {
  key?: string | null
  year?: string | number | null
  authors?: Array<string | { name?: string | null }> | null
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
    (item.authors ?? []).map((value) => normalizeText(value)).join(', '),
    (item.cities ?? []).map((value) => normalizeText(value)).join(', '),
  ]
    .filter(Boolean)
    .join(', ')
  const title = item.shortTitle || item.title
  if (!details && !title) return item.key || ''
  if (!title) return details
  if (!details) return title
  return `${details} - ${title}`
}
