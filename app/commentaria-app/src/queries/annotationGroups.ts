import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  AnnotationGroupsService,
  type annotation_Group,
  type annotation_Reference,
} from '@hub-api'

export const annotationGroupsQueryKey = () => ['annotation-groups'] as const

export function useAnnotationGroupsQuery() {
  return useQuery({
    queryKey: annotationGroupsQueryKey(),
    queryFn: () => AnnotationGroupsService.getAnnotationGroups(),
    select: (data) => data ?? [],
  })
}

export function useCreateAnnotationGroupMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      name,
      description,
      annotations,
    }: {
      name: string
      description?: string
      annotations: annotation_Reference[]
    }) =>
      AnnotationGroupsService.postAnnotationGroups({
        group: {
          name,
          description,
          annotations,
        } satisfies annotation_Group,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: annotationGroupsQueryKey() })
    },
  })
}

export function useUpdateAnnotationGroupMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      groupId,
      group,
    }: {
      groupId: string
      group: annotation_Group
    }) => AnnotationGroupsService.putAnnotationGroups({ groupId, group }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: annotationGroupsQueryKey() })
    },
  })
}

export function useDeleteAnnotationGroupMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (groupId: string) =>
      AnnotationGroupsService.deleteAnnotationGroups({ groupId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: annotationGroupsQueryKey() })
    },
  })
}
