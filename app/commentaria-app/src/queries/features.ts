import { useQuery, useQueries } from '@tanstack/react-query'
import { EditionFeaturesService } from '@hub-api'

export function useGlobalFeaturesQuery(
  expand: string[] = ['revisions'],
  enabled = true,
) {
  return useQuery({
    queryKey: ['features', 'editions', ...expand] as const,
    queryFn: () =>
      EditionFeaturesService.getFeatures({
        scope: 'editions',
        expand,
      }),
    enabled,
    refetchOnWindowFocus: false,
  })
}

export function useAllDatasetsFeaturesQueries(
  datasetIds: string[],
  expand: string[] = ['revisions'],
) {
  return useQueries({
    queries: datasetIds.map((datasetId) => ({
      queryKey: ['features', 'dataset', datasetId, ...expand] as const,
      queryFn: () =>
        EditionFeaturesService.getFeatures({
          scope: 'dataset',
          dataset: datasetId,
          expand,
        }),
      refetchOnWindowFocus: false,
    })),
  })
}

export function useFeaturesForExecutionsQuery(datasetIds: string[]) {
  const expand = ['revisions']
  return useQueries({
    queries: [
      {
        queryKey: ['features', 'editions', 'executions-browser'] as const,
        queryFn: () =>
          EditionFeaturesService.getFeatures({
            scope: 'editions',
            expand,
          }),
        refetchOnWindowFocus: false,
      },
      ...datasetIds.map((datasetId) => ({
        queryKey: [
          'features',
          'dataset',
          datasetId,
          'executions-browser',
        ] as const,
        queryFn: () =>
          EditionFeaturesService.getFeatures({
            scope: 'dataset',
            dataset: datasetId,
            expand,
          }),
        refetchOnWindowFocus: false,
      })),
    ],
  })
}
