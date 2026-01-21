import { useQuery } from '@tanstack/react-query'
import { DatasetsService } from '../api'

export const datasetsQueryKey = () => ['datasets'] as const

export function useDatasetsQuery() {
  return useQuery({
    queryKey: datasetsQueryKey(),
    queryFn: () => DatasetsService.getDatasets({}),
  })
}

export function useDatasetSuggestedRules(datasetId: string) {
  return useQuery({
    queryKey: ['datasets', datasetId, 'suggestedRules'],
    queryFn: () =>
      DatasetsService.getDatasetsSuggestedRules({
        dataSetId: datasetId,
      }),
    enabled: !!datasetId,
  })
}
