import { useEditionQuery } from '../../queries/editions.ts'

interface EditionDetailsTableProps {
  editionId: string
}

interface EditionDetailRow {
  key: string
  value: string
}

const normalizeString = (value: string | null | undefined): string | null => {
  if (!value) return null
  const normalized = value.trim()
  return normalized || null
}

const normalizeStringList = (
  value: string[] | null | undefined,
): string | null => {
  if (!value?.length) return null
  const normalized = value.map((item) => item.trim()).filter(Boolean)
  return normalized.length ? normalized.join(', ') : null
}

const toTitleCase = (value: string): string =>
  value
    .toLowerCase()
    .split(/\s+/)
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')

const getRows = (
  edition:
    | {
        title?: string | null
        year?: string | null
        editor?: string[] | null
        cities?: string[] | null
        languages?: string[] | null
      }
    | null
    | undefined,
): EditionDetailRow[] => {
  if (!edition) return []

  const rows: EditionDetailRow[] = []

  const title = normalizeString(edition.title)
  if (title) rows.push({ key: 'Title', value: title })

  const year = normalizeString(edition.year)
  if (year) rows.push({ key: 'Year', value: year })

  const editor = normalizeStringList(edition.editor)
  if (editor) rows.push({ key: 'Editor', value: editor })

  const cities = normalizeStringList(edition.cities)
  if (cities) rows.push({ key: 'Cities', value: cities })

  const languages = normalizeStringList(
    edition.languages?.map((language) => toTitleCase(language)),
  )
  if (languages) rows.push({ key: 'Languages', value: languages })

  return rows
}

export function EditionDetailsTable({ editionId }: EditionDetailsTableProps) {
  const editionQuery = useEditionQuery(editionId)
  const rows = getRows(editionQuery.data)
  const isLoading = editionQuery.isLoading
  const error =
    editionQuery.error instanceof Error
      ? editionQuery.error.message
      : editionQuery.error
        ? 'Failed to load edition.'
        : null

  if (isLoading) return <span>Loading…</span>
  if (error) return <span>{error}</span>
  if (!rows.length) return <span>No edition properties.</span>

  return (
    <div className="border border-gray-200 rounded-md bg-white p-2">
      <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1.5 items-start">
        {rows.map((row) => (
          <div key={row.key} className="contents">
            <div className="text-xs font-semibold text-gray-600 break-all">
              {row.key}
            </div>
            <div className="text-xs text-gray-800 wrap-break-word">
              {row.value}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
