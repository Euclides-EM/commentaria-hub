import type { annotation_Annotation } from '@hub-api'

export const TITLE_PAGES_DATASET_ID = 'tps'

const IMAGE_KEY_EXTENSION_PATTERN = /\.(png|jpe?g|tiff?|gif|webp)$/i
const PAGE_SEGMENT_PATTERN = /^\d+(?:\s*-\s*\d+)?$/

export interface EditionDisplayInfo {
  key?: string | null
  year?: string | number | null
  editors?: Array<string | { name?: string | null }> | null
  cities?: Array<string | { name?: string | null }> | null
  shortTitle?: string | null
  title?: string | null
}

export interface ImageLookupValue {
  key?: string | null
  filename?: string | null
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

export const findMatchingImage = <T extends ImageLookupValue>(
  value: string | null | undefined,
  images: T[],
): T | null => {
  const candidates = buildEditionLookupCandidates(value)
  if (!candidates.length) return null

  return (
    images.find((image) => {
      const imageCandidates = [
        ...buildEditionLookupCandidates(image.key),
        ...buildEditionLookupCandidates(image.filename),
      ]
      return imageCandidates.some((candidate) => candidates.includes(candidate))
    }) || null
  )
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

export const findMatchingEditionKeyFromValues = (
  values: Array<string | null | undefined>,
  editionKeys: Array<string | null | undefined>,
): string | null => {
  for (const value of values) {
    const match = findMatchingEditionKey(value, editionKeys)
    if (match) {
      return match
    }
  }

  return null
}

export const hasAnnotationPages = (
  annotation: Pick<annotation_Annotation, 'pages'> | null | undefined,
): boolean => {
  const parts = String(annotation?.pages ?? '')
    .split(',')
    .map((value) => value.trim())
    .filter(Boolean)

  return (
    parts.length > 0 && parts.every((part) => PAGE_SEGMENT_PATTERN.test(part))
  )
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
