import { type AnnotationRule, getRuleDisplayName } from '../utils/rules.ts'

interface RuleDisplayProps {
  rule: AnnotationRule
  isApplied?: boolean
  onRun?: () => void
}

export function RuleDisplay({ rule, isApplied, onRun }: RuleDisplayProps) {
  return (
    <div
      className={`p-3 border rounded-lg transition-colors ${
        isApplied
          ? 'bg-green-50 border-green-300'
          : 'bg-gray-50 border-gray-200 hover:bg-gray-100'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1">
          <div className="font-medium text-sm mb-1">
            {getRuleDisplayName(rule)}
          </div>
          <details className="text-xs">
            <summary className="cursor-pointer text-gray-600 hover:text-gray-800">
              Payload
            </summary>
            <pre className="mt-2 p-2 bg-white rounded border border-gray-200 font-mono overflow-auto whitespace-pre-wrap break-all">
              {JSON.stringify(rule, null, 2)}
            </pre>
          </details>
        </div>
        <div className="flex items-center gap-2">
          {isApplied && (
            <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">
              Applied
            </span>
          )}
          {onRun && (
            <button
              onClick={onRun}
              className="inline-flex items-center px-3 py-1 text-xs font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Run
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
