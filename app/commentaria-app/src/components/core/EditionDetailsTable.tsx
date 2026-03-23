import { useEditionQuery } from '../../queries/editions.ts'

interface EditionDetailsTableProps {
  editionId: string
  omitTitle?: boolean
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
        key?: string | null
        shortTitle?: string | null
        title?: string | null
        year?: string | null
        editor?: string[] | null
        cities?: string[] | null
        languages?: string[] | null
      }
    | null
    | undefined,
  omitTitle: boolean,
): EditionDetailRow[] => {
  if (!edition) return []

  const rows: EditionDetailRow[] = []

  const title = omitTitle ? null : normalizeString(edition.title)
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

export function EditionDetailsTable({
  editionId,
  omitTitle = false,
}: EditionDetailsTableProps) {
  const editionQuery = useEditionQuery(editionId)
  const rows = getRows(editionQuery.data, omitTitle)
  const isLoading = editionQuery.isLoading
  const resourceBoxBaseUrl = normalizeString(
    import.meta.env.VITE_RESOURCEBOX_APP_URL,
  )
  const editionKey = normalizeString(editionQuery.data?.key) || editionId
  const resourceBoxUrl =
    resourceBoxBaseUrl && editionKey
      ? `${resourceBoxBaseUrl.replace(/\/$/, '')}/catalogue?editionKey=${encodeURIComponent(editionKey)}`
      : null
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
      {resourceBoxUrl ? (
        <div className="mt-2 pt-2 border-t border-gray-200">
          <a
            href={resourceBoxUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-blue-600 hover:text-blue-700 hover:underline"
          >
            View in Elements Resource Box
            <span aria-hidden="true" className="ml-1 inline-block text-[10px]">
              ↗
            </span>
          </a>
        </div>
      ) : null}
    </div>
  )
}
