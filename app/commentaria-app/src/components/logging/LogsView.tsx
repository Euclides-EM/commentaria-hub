import { useState } from 'react'
import { Button } from '../core/Button.tsx'
import { ErrorMessage } from '../core/ErrorMessage.tsx'
import { LoadingSpinner } from '../core/LoadingSpinner.tsx'
import { useLogsQuery } from '../../queries/logs.ts'

const LINE_COUNT_OPTIONS = [100, 200, 500, 1000]

export function LogsView() {
  const [lineCount, setLineCount] = useState(200)
  const { data, isLoading, error, isFetching, refetch } =
    useLogsQuery(lineCount)

  return (
    <div className="h-full flex flex-col px-8 py-6 gap-4">
      <div className="flex flex-wrap items-center justify-between gap-4 rounded-lg border border-gray-200 bg-white px-6 py-4">
        <div>
          <h2 className="text-lg font-semibold text-gray-900">Server Logs</h2>
          <p className="text-sm text-gray-500">
            Service: {data?.service || 'commentaria-hub-api'}
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <label className="flex items-center gap-2 text-sm text-gray-700">
            <span>Lines</span>
            <select
              value={lineCount}
              onChange={(e) => setLineCount(Number(e.target.value))}
              className="h-9 rounded-md border border-gray-300 bg-white px-2 text-sm text-gray-700 focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100"
            >
              {LINE_COUNT_OPTIONS.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
            </select>
          </label>
          <Button
            type="button"
            onClick={() => void refetch()}
            className="px-3 py-2 text-sm"
          >
            {isFetching ? 'Refreshing...' : 'Refresh'}
          </Button>
        </div>
      </div>

      {error && <ErrorMessage error={error} />}

      <div className="rounded-lg border border-gray-200 bg-gray-950 text-gray-100 flex-1 min-h-0 overflow-hidden">
        {isLoading ? (
          <div className="h-full flex items-center justify-center">
            <LoadingSpinner message="Loading logs..." />
          </div>
        ) : (
          <div className="h-full overflow-auto">
            <div className="flex items-center justify-between border-b border-gray-800 px-4 py-2 text-xs text-gray-400">
              <span>{data?.count ?? 0} lines shown</span>
              <span>Auto-refresh every 15s</span>
            </div>
            <pre className="whitespace-pre-wrap break-words px-4 py-3 text-xs leading-5 font-mono">
              {data?.lines?.length
                ? data.lines.join('\n')
                : 'No log lines returned.'}
            </pre>
          </div>
        )}
      </div>
    </div>
  )
}
