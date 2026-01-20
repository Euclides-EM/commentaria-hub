import { useQuery } from '@tanstack/react-query'
import { DatasetsService } from '../api'

export const datasetsQueryKey = () => ['datasets'] as const

export function useDatasetsQuery() {
  return useQuery({
    queryKey: datasetsQueryKey(),
    queryFn: () => DatasetsService.getDatasets({}),
  })
}
