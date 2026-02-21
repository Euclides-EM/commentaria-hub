import { useQuery } from '@tanstack/react-query'
import { DatasetsService } from '@hub-api'

export const datasetsQueryKey = () => ['datasets'] as const
const datasetsImagesQueryKey = (datasetId: string) =>
  ['datasets', datasetId, 'images'] as const

export interface DatasetImageKey {
  filename: string
  name: string
}

export function useDatasetsQuery() {
  return useQuery({
    queryKey: datasetsQueryKey(),
    queryFn: () => DatasetsService.getDatasets({}),
    refetchInterval: 10 * 1000,
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

export function useDatasetImageKeysQuery(datasetId: string, enabled = true) {
  return useQuery({
    queryKey: datasetsImagesQueryKey(datasetId),
    queryFn: async (): Promise<DatasetImageKey[]> => {
      const images = await DatasetsService.getDatasetsImages({
        dataSetId: datasetId,
      })

      return images
        .map((image) => {
          const filename = image.filename?.trim() || ''
          const name = image.name?.trim() || ''
          return { filename, name }
        })
        .filter((image) => image.filename.length > 0 && image.name.length > 0)
        .sort((a, b) =>
          a.name.localeCompare(b.name, undefined, { numeric: true }),
        )
    },
    enabled: !!datasetId && enabled,
  })
}
