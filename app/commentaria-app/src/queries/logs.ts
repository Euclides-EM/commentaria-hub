import { useQuery } from '@tanstack/react-query'
import { LogsService } from '@hub-api'

const logsQueryKey = (lineCount: number) => ['logs', lineCount] as const

export function useLogsQuery(lineCount: number) {
  return useQuery({
    queryKey: logsQueryKey(lineCount),
    queryFn: () => LogsService.getLogs({ n: lineCount }),
    refetchInterval: 15 * 1000,
  })
}
