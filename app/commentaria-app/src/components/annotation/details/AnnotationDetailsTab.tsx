import { AnnotationDetailsPane } from './AnnotationDetailsPane.tsx'
import { SuggestedRulesPane } from '../../rules/SuggestedRulesPane.tsx'

export function AnnotationDetailsTab() {
  return (
    <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
      <AnnotationDetailsPane />
      <SuggestedRulesPane />
    </div>
  )
}
