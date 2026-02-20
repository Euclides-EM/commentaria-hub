import { useQuery } from '@tanstack/react-query'
import { IntegrationService } from '../api'

export const runningIntegrationJobsQueryKey = () =>
  ['integrations', 'jobs', 'running'] as const

export const nonCompletedIntegrationJobsQueryKey = () =>
  ['integrations', 'jobs', 'non-completed'] as const

export function useRunningIntegrationJobsQuery() {
  return useQuery({
    queryKey: runningIntegrationJobsQueryKey(),
    queryFn: async () => {
      const jobs = (await IntegrationService.getIntegrationsJobs()) || []
      return jobs.filter(
        (job) =>
          job.task === 'Export' &&
          (job.status === 'pending' || job.status === 'running'),
      )
    },
    staleTime: 10 * 1000,
    refetchInterval: 10 * 1000,
  })
}

export function useNonCompletedIntegrationJobsQuery() {
  return useQuery({
    queryKey: nonCompletedIntegrationJobsQueryKey(),
    queryFn: async () => {
      const jobs = (await IntegrationService.getIntegrationsJobs()) || []
      return jobs.filter((job) => job.status !== 'completed')
    },
    staleTime: 10 * 1000,
    refetchInterval: 10 * 1000,
  })
}
