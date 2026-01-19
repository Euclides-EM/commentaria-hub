import { useQuery } from '@tanstack/react-query'
import { DatasetsService } from '../api'

export const datasetsQueryKey = () => ['datasets'] as const

export function useDatasetsQuery() {
  return useQuery({
    queryKey: datasetsQueryKey(),
    queryFn: () => DatasetsService.getDatasets({}),
  })
}

export const datasetPageImageQueryKey = (datasetId: string, pageNum: number) =>
  ['datasets', datasetId, 'pages', pageNum, 'image'] as const

export function useDatasetPageImageQuery(datasetId: string, pageNum: number) {
  return useQuery({
    queryKey: datasetPageImageQueryKey(datasetId, pageNum),
    queryFn: () =>
      DatasetsService.getDatasetsPagesImage({
        dataSetId: datasetId,
        pageNum: pageNum.toString(),
      }),
    enabled: !!datasetId,
  })
}
