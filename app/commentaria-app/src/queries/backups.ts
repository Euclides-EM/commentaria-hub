import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { BackupsService } from '@hub-api'
import { nonCompletedIntegrationJobsQueryKey } from './integrations'

export const backupsQueryKey = () => ['backups'] as const

export function useBackupsQuery() {
  return useQuery({
    queryKey: backupsQueryKey(),
    queryFn: async () => {
      const backups = await BackupsService.getBackups()
      return [...backups].sort((a, b) => b.localeCompare(a))
    },
  })
}

export function useCreateBackupMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => BackupsService.postBackups({ async: true }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: backupsQueryKey() })
      queryClient.invalidateQueries({
        queryKey: nonCompletedIntegrationJobsQueryKey(),
      })
    },
  })
}

export function useCreateBackupFromZipMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => BackupsService.postBackupsFromzip({ file }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: backupsQueryKey() })
    },
  })
}

export function useSyncBackupToDriveMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (backupId: string) =>
      BackupsService.putBackupsSync({ backupId, async: true }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: backupsQueryKey() })
      queryClient.invalidateQueries({
        queryKey: nonCompletedIntegrationJobsQueryKey(),
      })
    },
  })
}

export function useRestoreBackupMutation() {
  return useMutation({
    mutationFn: (backupId: string) =>
      BackupsService.putBackupsRestore({ backupId }),
  })
}
