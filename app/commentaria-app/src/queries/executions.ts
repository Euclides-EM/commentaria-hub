import { useQuery } from '@tanstack/react-query'
import { ExecutionsService } from '@hub-api'

export function useFeatureExecutionsQuery({
  scope,
  datasetId,
  refetchInterval = 5000,
}: {
  scope?: 'all' | 'dataset' | 'editions'
  datasetId?: string
  refetchInterval?: number | false
}) {
  return useQuery({
    queryKey: ['executions', scope || 'all', datasetId || 'all'] as const,
    queryFn: () =>
      ExecutionsService.getFeatureExecutions({
        scope: scope === 'all' ? undefined : scope,
        dataset: datasetId || undefined,
      }),
    refetchInterval,
    refetchOnWindowFocus: false,
  })
}
