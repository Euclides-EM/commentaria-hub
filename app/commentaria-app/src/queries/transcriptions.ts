import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  TranscriptionsService,
  type model_EditionTranscription,
} from '@hub-api'

export const editionTranscriptionsQueryKey = (editionId: string) =>
  ['editions', editionId, 'transcriptions'] as const

export function useEditionTranscriptionsQuery(
  editionId: string | null | undefined,
  enabled = true,
) {
  return useQuery({
    queryKey: editionTranscriptionsQueryKey(editionId || ''),
    queryFn: async () => {
      if (!editionId) {
        return []
      }
      return TranscriptionsService.getEditionsTranscriptions({
        editionId: [editionId],
      })
    },
    enabled: !!editionId && enabled,
    refetchOnWindowFocus: false,
  })
}

export function useUpdateEditionTranscriptionMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      editionId,
      body,
    }: {
      editionId: string
      body: model_EditionTranscription
    }) =>
      TranscriptionsService.putEditionsTranscriptions({
        editionId,
        body,
      }),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: editionTranscriptionsQueryKey(variables.editionId),
      })
    },
  })
}
