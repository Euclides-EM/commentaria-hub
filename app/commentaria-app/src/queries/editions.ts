import { useQuery } from '@tanstack/react-query'
import { EditionsService, type model_Edition } from '@hub-api'

const editionQueryKey = (editionId: string) => ['editions', editionId] as const

export function useEditionQuery(
  editionId: string | null | undefined,
  enabled = true,
) {
  return useQuery({
    queryKey: editionQueryKey(editionId || ''),
    queryFn: () => EditionsService.getEditions({ editionId: editionId! }),
    enabled: !!editionId && enabled,
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
