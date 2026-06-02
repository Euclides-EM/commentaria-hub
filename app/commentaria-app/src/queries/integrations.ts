import { useQuery } from '@tanstack/react-query'
import { type job_Job, JobsService } from '@hub-api'

export const runningIntegrationJobsQueryKey = () =>
  ['integrations', 'jobs', 'running'] as const

export const nonCompletedIntegrationJobsQueryKey = () =>
  ['integrations', 'jobs', 'non-completed'] as const

export function useRunningIntegrationJobsQuery() {
  return useQuery({
    queryKey: runningIntegrationJobsQueryKey(),
    queryFn: async () => {
      const jobs = (await JobsService.getJobs()) || []
      return jobs.filter(
        (job) => job.status === 'pending' || job.status === 'running',
      )
    },
    staleTime: 10 * 1000,
    refetchInterval: 10 * 1000,
  })
}

export const isExportJob = (job: job_Job) => job.task === 'Export'

export const isAnnotationRuleApplyJob = (job: job_Job) =>
  job.task === 'AnnotationRuleApply'

export function useNonCompletedIntegrationJobsQuery() {
  return useQuery({
    queryKey: nonCompletedIntegrationJobsQueryKey(),
    queryFn: async () => {
      return (await JobsService.getJobs()) || []
    },
    staleTime: 10 * 1000,
    refetchInterval: 10 * 1000,
  })
}
