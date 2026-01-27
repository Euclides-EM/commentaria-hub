import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ModelsService } from '../api'

export const modelsQueryKey = () => ['models'] as const

export function useModelsQuery() {
  return useQuery({
    queryKey: modelsQueryKey(),
    queryFn: () => ModelsService.getModels(),
  })
}

export function useDeleteModelMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, deep }: { id: string; deep?: boolean }) =>
      ModelsService.deleteModels({
        id,
        deep: deep ? 'true' : undefined,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: modelsQueryKey() })
    },
  })
}
