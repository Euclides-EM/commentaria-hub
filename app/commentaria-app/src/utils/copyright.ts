import type { model_Edition, model_Facsimile } from '@hub-api'

type EditionWithShelfmarks = model_Edition & {
  shelfmarks?: Array<{ id?: string; scan?: string; copyright?: string }>
}

export type CopyrightStatus = 'unknown' | 'known'

export const COPYRIGHT_FILTER_OPTIONS: CopyrightStatus[] = ['unknown', 'known']

export const formatCopyright = (copyright?: string) =>
  copyright?.trim().toLowerCase() || 'unknown copyright'

export function getFacsimileCopyright(
  edition?: EditionWithShelfmarks,
  facsimile?: model_Facsimile,
) {
  const shelfmarks = edition?.shelfmarks ?? []
  const shelfmarkId = facsimile?.shelfmark_id?.trim()
  const scanUrl = facsimile?.scan_url?.trim()
  const matchingShelfmark =
    (shelfmarkId
      ? shelfmarks.find((shelfmark) => shelfmark.id?.trim() === shelfmarkId)
      : undefined) ??
    (scanUrl
      ? shelfmarks.find((shelfmark) => shelfmark.scan?.trim() === scanUrl)
      : undefined)

  if (matchingShelfmark) {
    return formatCopyright(matchingShelfmark.copyright)
  }

  return formatCopyright()
}

export const getCopyrightStatus = (copyright: string): CopyrightStatus =>
  copyright === 'unknown copyright' ? 'unknown' : 'known'
