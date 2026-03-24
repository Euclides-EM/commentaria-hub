import type { annotation_Annotation } from '@hub-api'

export const TITLE_PAGES_DATASET_ID = 'tps'

const IMAGE_KEY_EXTENSION_PATTERN = /\.(png|jpe?g|tiff?|gif|webp)$/i

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

const normalizeEditionLookupKey = (value: string): string =>
  value.trim().toLowerCase()

const stripImageKeyExtension = (value: string): string =>
  value.replace(IMAGE_KEY_EXTENSION_PATTERN, '')

export const buildEditionLookupCandidates = (
  value: string | null | undefined,
): string[] => {
  const trimmed = String(value ?? '').trim()
  if (!trimmed) return []

  const withoutExtension = stripImageKeyExtension(trimmed)
  return [
    ...new Set([trimmed, withoutExtension].map(normalizeEditionLookupKey)),
  ]
}

export const findMatchingEditionKey = (
  value: string | null | undefined,
  editionKeys: Array<string | null | undefined>,
): string | null => {
  const candidates = buildEditionLookupCandidates(value)
  if (!candidates.length) return null

  for (const editionKey of editionKeys) {
    const normalizedEditionKeyCandidates =
      buildEditionLookupCandidates(editionKey)
    if (
      normalizedEditionKeyCandidates.some((candidate) =>
        candidates.includes(candidate),
      )
    ) {
      return String(editionKey).trim()
    }
  }

  return null
}

export const hasAnnotationPages = (
  annotation: Pick<annotation_Annotation, 'pages'> | null | undefined,
): boolean => Boolean(annotation?.pages?.trim())

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
