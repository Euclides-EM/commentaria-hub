import { useQuery } from '@tanstack/react-query'
import { FeatureResultsService } from '@hub-api'

export function useFeatureResultsQuery({
  scope,
  datasetId,
  annotationId,
  enabled = true,
}: {
  scope: 'dataset' | 'editions'
  datasetId?: string
  annotationId?: string
  enabled?: boolean
}) {
  return useQuery({
    queryKey: [
      'featureResults',
      'browser',
      scope,
      datasetId || 'all',
      annotationId || 'all',
    ] as const,
    queryFn: () =>
      FeatureResultsService.getFeaturesResults({
        scope,
        dataset: scope === 'dataset' ? datasetId : undefined,
        annotation: scope === 'dataset' ? annotationId : undefined,
        fallbackToOrigin: scope === 'dataset' ? true : undefined,
      }),
    enabled:
      enabled &&
      (scope === 'editions' || (Boolean(datasetId) && Boolean(annotationId))),
    refetchOnWindowFocus: false,
  })
}
