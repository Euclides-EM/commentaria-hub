import { useQuery } from '@tanstack/react-query'
import { EditionsService } from '@hub-api'

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
