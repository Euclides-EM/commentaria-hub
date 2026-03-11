import {
  EditionsService,
  type model_Edition,
  type search_Query,
} from '@hub-api'

export type EditionItem = {
  key: string
  year: string | null
  authors: string[]
  cities: string[]
  shortTitle: string | null
  title: string | null
  studyCorpora: string[]
}

const STUDY_CORPORA_DISPLAY: Record<string, string> = {
  dh: 'DH core texts',
  dotted_lines: 'Dotted Lines',
}

const dedupe = (values: string[]) => Array.from(new Set(values))

const startCase = (value: string) =>
  value
    .replace(/[_-]+/g, ' ')
    .trim()
    .split(/\s+/)
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ')

const normalizeText = (value: unknown): string => {
  if (typeof value === 'string') return value.trim()
  if (value && typeof value === 'object' && 'name' in value) {
    return String((value as { name?: unknown }).name ?? '').trim()
  }
  return String(value ?? '').trim()
}

const mapStudyCorpus = (value: string): string =>
  STUDY_CORPORA_DISPLAY[value] || startCase(value)

export const STUDY_CORPORA_FILTER = 'Title pages'

export const mapEditionsToItems = (editions: model_Edition[]): EditionItem[] =>
  editions
    .filter((edition): edition is model_Edition & { key: string } =>
      Boolean(edition.key),
    )
    .map((edition) => {
      const studyCorpora = dedupe(
        (edition.corpus ?? []).map((corpus) => mapStudyCorpus(corpus)),
      )
      if (
        (!Number(edition.year) || Number(edition.year) <= 1700) &&
        studyCorpora.includes('Origin Eip Csv') &&
        !edition.languages?.includes('CHINESE') &&
        edition.title &&
        edition.title !== '?'
      ) {
        studyCorpora.push(STUDY_CORPORA_FILTER)
      }

      return {
        key: edition.key,
        year: edition.year != null ? String(edition.year) : null,
        authors: (edition.editor ?? [])
          .map((name) => normalizeText(name))
          .filter(Boolean),
        cities: (edition.cities ?? [])
          .map((city) => normalizeText(city))
          .filter(Boolean),
        shortTitle: edition.shortTitle || null,
        title: edition.title || null,
        studyCorpora: dedupe(studyCorpora),
      }
    })
    .sort(
      (left, right) =>
        (left.year || '').localeCompare(right.year || '') ||
        left.key.localeCompare(right.key),
    )

export const listAllEditions = async (): Promise<model_Edition[]> => {
  const limit = 500
  let offset = 0
  const results: model_Edition[] = []

  while (true) {
    const page = await EditionsService.postEditionsSearch({
      edition: {
        offset,
        limit,
        filter_includes: {
          titlePageStatus: false,
        },
        fields_filter: {
          isManuscript: ['false'],
          titlePageStatus: ['No', 'Unknown'],
        },
        order_by: [{ field: 'year' }, { field: 'cities' }],
      },
    })
    const items = page.items || []
    results.push(...items)
    if (
      items.length === 0 ||
      items.length < limit ||
      (page.total !== undefined && results.length >= page.total)
    ) {
      break
    }
    offset += limit
  }

  return results
}
