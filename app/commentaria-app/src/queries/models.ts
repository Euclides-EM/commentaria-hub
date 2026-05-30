import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ModelsService,
  type model_Model,
  type model_ModelTraining,
} from '@hub-api'

const modelsQueryKey = () => ['models'] as const

export function useModelsQuery() {
  return useQuery({
    queryKey: modelsQueryKey(),
    queryFn: () => ModelsService.getModels({ expand: ['used_in_annotations'] }),
  })
}

export function useDeleteModelMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, deep }: { id: string; deep?: boolean }) =>
      ModelsService.deleteModels({
        id,
        deep: deep || undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelsQueryKey() })
    },
  })
}

export function useCreateModelMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      file,
      name,
      description,
      baseModelId,
      baseAnnotations,
    }: {
      file: File
      name: string
      description?: string
      baseModelId?: string
      baseAnnotations?: string
    }) =>
      ModelsService.postModelsUpload({
        file,
        name,
        description,
        baseAnnotations,
        baseModelId,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelsQueryKey() })
    },
  })
}

export function useTrainModelMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (training: model_ModelTraining) =>
      ModelsService.postModelsTrain({ training }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelsQueryKey() })
      queryClient.invalidateQueries({ queryKey: ['jobs'] })
    },
  })
}

export type TrainModelResult = model_ModelTraining

export function useUpdateModelMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, model }: { id: string; model: model_Model }) =>
      ModelsService.putModels({ id, model }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelsQueryKey() })
    },
  })
}
