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
  let title = item.shortTitle || item.title || ''
  title = title.length > 128 ? title.slice(0, 125) + '...' : title
  if (title === '?') {
    title = ''
  }
  return title ? `${details} - ${title}` : details
}
