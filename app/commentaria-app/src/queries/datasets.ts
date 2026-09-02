import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  DatasetImagesService,
  DatasetsService,
  EditionFeaturesService,
  FeaturePropertiesService,
} from '@hub-api'
import { normalizeFeatureProperties } from '../utils/featureProperties.ts'

export const datasetsQueryKey = () => ['datasets'] as const
export const datasetsImagesQueryKey = (
  datasetId: string,
  keys: string[] | null = null,
) => ['datasets', datasetId, 'images', keys ?? null] as const

export interface DatasetImageKey {
  key: string
  filename: string
  modifiedAt: string
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
    queryKey: datasetsImagesQueryKey(datasetId, keys),
    queryFn: async (): Promise<DatasetImageKey[]> => {
      const images = await DatasetImagesService.getDatasetsImages({
        dataSetId: datasetId,
        uniqueOnly: true,
      })

      const mapped = images
        .map((image) => {
          const key = image.key?.trim() || ''
          const filename = image.filename?.trim() || ''
          const modifiedAt = image.modified_at?.trim() || ''
          return { key, filename, modifiedAt }
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
        mapped.push(
          ...missingKeys.map((key) => ({ key, filename: '', modifiedAt: '' })),
        )
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
      EditionFeaturesService.getFeatures({
        scope: 'dataset',
        dataset: datasetId,
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

export function useReplaceDatasetImageMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      datasetId,
      key,
      type,
      file,
    }: {
      datasetId: string
      key?: string
      type: 'facsimile' | 'tp'
      file: File
    }) =>
      DatasetImagesService.postDatasetsImagesUpload({
        dataSetId: datasetId,
        key,
        type,
        file,
      }),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ['datasets', variables.datasetId, 'images'],
      })
    },
  })
}
