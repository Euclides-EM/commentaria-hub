import type { model_Edition, model_Facsimile } from '@hub-api'

export type CopyrightStatus = 'unknown' | 'known'

export const COPYRIGHT_FILTER_OPTIONS: CopyrightStatus[] = ['unknown', 'known']

export const formatCopyright = (copyright?: string) =>
  copyright?.trim().toLowerCase() || 'unknown copyright'

export function getFacsimileCopyright(
  edition?: model_Edition,
  facsimile?: model_Facsimile,
) {
  const shelfmarks = edition?.shelfmarks ?? []
  const scanUrl = facsimile?.scan_url?.trim()
  const matchingShelfmark = scanUrl
    ? shelfmarks.find((shelfmark) => shelfmark.scan?.trim() === scanUrl)
    : undefined

  if (matchingShelfmark) {
    return formatCopyright(matchingShelfmark.copyright)
  }

  const copyrights = [
    ...new Set(
      shelfmarks
        .map((shelfmark) => shelfmark.copyright?.trim())
        .filter((copyright): copyright is string => !!copyright),
    ),
  ]
  return formatCopyright(copyrights.join('; '))
}

export const getCopyrightStatus = (copyright: string): CopyrightStatus =>
  copyright === 'unknown copyright' ? 'unknown' : 'known'
