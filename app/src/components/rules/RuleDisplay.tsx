import { type AnnotationRule, getRuleDisplayName } from '../../utils/rules.ts'
import type { annotationrule_PipelineStage } from '../../api/models/annotationrule_PipelineStage.ts'
import { Fragment } from 'react'
import { Button } from '../core/Button'
import { getStageDisplayName } from '../../utils/stages.ts'
import { useAppState } from '../../context/useAppState'

interface RuleDisplayProps {
  rule: AnnotationRule
  isApplied?: boolean
  onRun?: () => void
  disabled?: boolean
  applicableStages?: annotationrule_PipelineStage[]
}

export function RuleDisplay({
  rule,
  isApplied,
  onRun,
  disabled,
  applicableStages,
}: RuleDisplayProps) {
  const { setState, setModelSearchPrefill } = useAppState()
  const modelName =
    'model' in rule ? (rule as { model?: string }).model : undefined

  return (
    <div
      className={`p-3 border rounded-lg transition-colors ${
        isApplied
          ? 'bg-green-50 border-green-300'
          : 'bg-gray-50 border-gray-200 hover:bg-gray-100'
      }`}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex flex-col gap-1 flex-1">
          <div className="font-medium text-sm mb-1">
            {getRuleDisplayName(rule)}
          </div>
          {applicableStages && applicableStages.length > 0 && (
            <div className="text-xs text-gray-500 mb-1">
              Applicable Stages:{' '}
              {applicableStages.map((s, i) => (
                <Fragment key={s}>
                  <span className="font-semibold">
                    {getStageDisplayName(s)}
                  </span>
                  {i < applicableStages.length - 1 ? ', ' : ''}
                </Fragment>
              ))}
            </div>
          )}
          {modelName && (
            <div className="text-xs text-gray-500 mb-1">
              Model:{' '}
              <button
                type="button"
                className="text-teal-700 hover:text-teal-900 underline underline-offset-2"
                onClick={() => {
                  setModelSearchPrefill(modelName)
                  setState({ viewingModels: true, viewingGroundTruths: false })
                }}
              >
                {modelName}
              </button>
            </div>
          )}
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
            <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-md">
              Applied
            </span>
          )}
          {onRun && (
            <Button
              onClick={onRun}
              variant="primary"
              className="px-3 py-1 text-xs"
              disabled={disabled}
            >
              Run
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
