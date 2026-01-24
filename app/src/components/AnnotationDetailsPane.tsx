import { useAppState } from '../context/AppStateContext.tsx'
import type { model_Annotation } from '../api'
import TimeAgo from 'javascript-time-ago'
import en from 'javascript-time-ago/locale/en'
import { RuleDisplay } from './RuleDisplay.tsx'
import type { AnnotationRule } from '../utils/rules.ts'

TimeAgo.addDefaultLocale(en)
const timeAgo = new TimeAgo('en-US')

const Timestamp = ({ date }: { date: string | undefined }) => {
  if (!date) {
    return 'N/A'
  }
  const d = new Date(date)
  return (
    <div className="flex gap-2 items-center">
      <span>{timeAgo.format(d)}</span>
      <span className="text-xs text-gray-500">{d.toLocaleString()}</span>
    </div>
  )
}

const AnnotationDescriptor = ({
  annotation,
}: {
  annotation: model_Annotation
}) => {
  const appliedRules = (annotation.applied_rules || []) as AnnotationRule[]

  return (
    <div className="mt-2.5 border border-gray-200 rounded-lg bg-gray-50 p-3.5 overflow-auto leading-normal text-base box-border">
      <div className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 items-start">
        <div className="font-semibold text-xs opacity-80 pt-0.5">ID</div>
        <div className="text-sm leading-tight break-all font-mono">
          {annotation.id}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Name</div>
        <div className="text-sm leading-tight break-all">{annotation.name}</div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Description
        </div>
        <div className="text-sm leading-tight break-all">
          {annotation.description}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Dataset</div>
        <div className="text-sm leading-tight break-all font-mono">
          {annotation.dataset_id}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Pages</div>
        <div className="text-sm leading-tight break-all">
          {annotation.pages}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Segmented</div>
        <div className="text-sm leading-tight break-all">
          {String(!!annotation.segmented)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Ground truth
        </div>
        <div className="text-sm leading-tight break-all">
          {String(!!annotation.ground_truth)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">OCRed</div>
        <div className="text-sm leading-tight break-all">
          {String(!!annotation.ocred)}
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Created</div>
        <div className="text-sm leading-tight break-all ">
          <Timestamp date={annotation.created_at} />
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">Updated</div>
        <div className="text-sm leading-tight break-all ">
          <Timestamp date={annotation.updated_at} />
        </div>
        <div className="font-semibold text-xs opacity-80 pt-0.5">
          Applied rules
        </div>
        <div className="text-sm leading-tight">
          {appliedRules.length > 0 ? (
            <div className="space-y-2">
              {appliedRules.map((rule, index) => (
                <RuleDisplay key={index} rule={rule} isApplied={true} />
              ))}
            </div>
          ) : (
            <span className="text-gray-500">None</span>
          )}
        </div>
      </div>
    </div>
  )
}

export function AnnotationDetailsPane() {
  const { annotation } = useAppState()

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white m-3 mb-0">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Annotation Details</div>
      </div>

      {annotation && (
        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          <AnnotationDescriptor annotation={annotation} />
        </div>
      )}
    </section>
  )
}
