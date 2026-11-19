import { useQuery } from '@tanstack/react-query'
import { EditionsService, type model_Edition } from '@hub-api'

export const normalizeEditionId = (editionId: string | null | undefined) =>
  editionId?.replace(/_vol\d+$/, '') || ''

const editionQueryKey = (editionId: string) => ['editions', editionId] as const
const allEditionsQueryKey = (filter?: Record<string, string[]>) =>
  ['editions', 'all', 'items', filter ?? null] as const

export function useEditionQuery(
  editionId: string | null | undefined,
  enabled = true,
) {
  const normalizedEditionId = normalizeEditionId(editionId)

  return useQuery({
    queryKey: editionQueryKey(normalizedEditionId),
    queryFn: () =>
      EditionsService.getEditions({ editionId: normalizedEditionId }),
    enabled: !!normalizedEditionId && enabled,
    refetchOnWindowFocus: false,
  })
}

export const listAllEditions = async (
  filter?: Record<string, string[]>,
): Promise<model_Edition[]> => {
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
          ...(filter || {}),
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

export function useAllEditionsQuery(
  filter?: Record<string, string[]>,
  enabled = true,
) {
  return useQuery({
    queryKey: allEditionsQueryKey(filter),
    queryFn: async () => await listAllEditions(filter),
    enabled,
    refetchOnWindowFocus: false,
    staleTime: Infinity,
  })
}
