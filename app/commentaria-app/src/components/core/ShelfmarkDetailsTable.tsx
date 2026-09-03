import { useQuery } from '@tanstack/react-query'
import {
  FacsimilesService,
  ShelfmarksService,
  type model_EditionShelfmark,
} from '@hub-api'
import { normalizeEditionId } from '../../queries/editions.ts'

interface ShelfmarkDetailsTableProps {
  editionId: string
  facsimileId?: string | null
}

interface ShelfmarkDetailRow {
  key: string
  value: string
  href?: string
}

const normalizeString = (value: string | null | undefined): string | null => {
  if (!value) return null
  const normalized = value.trim()
  return normalized || null
}

const getRows = (
  shelfmark: model_EditionShelfmark | null | undefined,
): ShelfmarkDetailRow[] => {
  if (!shelfmark) return []

  const shelfmarkValue = normalizeString(shelfmark.shelfmark)
  const copyright = normalizeString(shelfmark.copyright)

  return [
    { key: 'Shelfmark', value: shelfmarkValue || 'N/A' },
    {
      key: 'Volume',
      value:
        shelfmark.volume !== undefined && shelfmark.volume !== null
          ? String(shelfmark.volume)
          : 'N/A',
    },
    { key: 'Copyright', value: copyright || 'Unknown copyright' },
  ]
}

export function ShelfmarkDetailsTable({
  editionId,
  facsimileId,
}: ShelfmarkDetailsTableProps) {
  const normalizedEditionId = normalizeEditionId(editionId)
  const facsimileQuery = useQuery({
    queryKey: ['facsimiles', facsimileId],
    queryFn: () => FacsimilesService.getFacsimilies1({ id: facsimileId! }),
    enabled: !!facsimileId,
  })
  const shelfmarksQuery = useQuery({
    queryKey: ['shelfmarks', normalizedEditionId],
    queryFn: () =>
      ShelfmarksService.getShelfmarks({
        editionId: normalizedEditionId ? [normalizedEditionId] : undefined,
      }),
    enabled: !!normalizedEditionId,
  })

  const shelfmarkId = normalizeString(facsimileQuery.data?.shelfmark_id)
  const facsimileScanUrl = normalizeString(facsimileQuery.data?.scan_url)
  const shelfmark = (shelfmarksQuery.data || []).find(
    (item) =>
      (shelfmarkId && normalizeString(item.id) === shelfmarkId) ||
      (facsimileScanUrl && normalizeString(item.scan) === facsimileScanUrl),
  )
  const facsimileUrl = normalizeString(shelfmark?.scan) || facsimileScanUrl
  const rows = [
    ...getRows(shelfmark),
    {
      key: 'Facsimile URL',
      value: facsimileUrl || 'N/A',
      href: facsimileUrl || undefined,
    },
  ]
  const isLoading = facsimileQuery.isLoading || shelfmarksQuery.isLoading
  const error = facsimileQuery.error || shelfmarksQuery.error

  if (!facsimileId) return <span>No facsimile assigned.</span>
  if (isLoading) return <span>Loading…</span>
  if (error) {
    return (
      <span>
        {error instanceof Error ? error.message : 'Failed to load shelfmark.'}
      </span>
    )
  }
  if (!shelfmark) return <span>No shelfmark assigned.</span>
  if (!rows.length) return <span>No shelfmark properties.</span>

  return (
    <div className="border border-gray-200 rounded-md bg-white p-2">
      <div className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1.5 items-start">
        {rows.map((row) => (
          <div key={row.key} className="contents">
            <div className="text-xs font-semibold text-gray-600 break-all">
              {row.key}
            </div>
            <div className="text-xs text-gray-800 wrap-break-word">
              {row.href ? (
                <a
                  href={row.href}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-blue-600 hover:text-blue-700 hover:underline"
                >
                  {row.value}
                  <span
                    aria-hidden="true"
                    className="ml-1 inline-block text-[10px]"
                  >
                    ↗
                  </span>
                </a>
              ) : (
                row.value
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
