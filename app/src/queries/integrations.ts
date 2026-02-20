import { useQuery } from '@tanstack/react-query'
import { IntegrationService } from '../api'

export const runningIntegrationJobsQueryKey = () =>
  ['integrations', 'jobs', 'running'] as const

export function useRunningIntegrationJobsQuery() {
  return useQuery({
    queryKey: runningIntegrationJobsQueryKey(),
    queryFn: async () => {
      const jobs = await IntegrationService.getIntegrationsJobs()
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
