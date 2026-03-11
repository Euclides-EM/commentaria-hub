import { useQuery } from '@tanstack/react-query'
import {
  DatasetImagesService,
  DatasetsService,
  FeaturePropertiesService,
  FeaturesService,
} from '@hub-api'
import { normalizeFeatureProperties } from '../utils/featureProperties.ts'

export const datasetsQueryKey = () => ['datasets'] as const
const datasetsImagesQueryKey = (datasetId: string) =>
  ['datasets', datasetId, 'images'] as const

export interface DatasetImageKey {
  key: string
  filename: string
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

export function useDatasetImageKeysQuery(
  datasetId: string,
  enabled = true,
  keys: string[] | null = null,
) {
  return useQuery({
    queryKey: datasetsImagesQueryKey(datasetId),
    queryFn: async (): Promise<DatasetImageKey[]> => {
      const images = await DatasetImagesService.getDatasetsImages({
        dataSetId: datasetId,
        uniqueOnly: true,
      })

      const mapped = images
        .map((image) => {
          const key = image.key?.trim() || ''
          const filename = image.filename?.trim() || ''
          return { key, filename }
        })
        .filter(
          (image) =>
            image.filename.length > 0 &&
            image.key.length > 0 &&
            (!keys || keys.includes(image.key)),
        )

      if (keys) {
        const existingKeys = new Set(mapped.map((image) => image.key))
        const missingKeys = keys.filter((key) => !existingKeys.has(key))
        mapped.push(...missingKeys.map((key) => ({ key, filename: '' })))
      }

      return mapped.sort((a, b) =>
        a.key.localeCompare(b.key, undefined, { numeric: true }),
      )
    },
    enabled: !!datasetId && enabled,
  })
}

export function useDatasetFeaturesQuery(datasetId: string, enabled = true) {
  return useQuery({
    queryKey: ['features', 'definitions', datasetId] as const,
    queryFn: () =>
      FeaturesService.getDatasetsFeatures({
        dataSetId: datasetId,
        expand: ['revisions'],
      }),
    enabled: !!datasetId && enabled,
    refetchOnWindowFocus: false,
  })
}

export function useFeaturePropertiesQuery(enabled = true) {
  return useQuery({
    queryKey: ['features', 'properties'] as const,
    queryFn: async () => {
      const properties = await FeaturePropertiesService.getFeaturesProperties()
      return normalizeFeatureProperties(properties)
    },
    enabled,
    refetchOnWindowFocus: false,
  })
}
