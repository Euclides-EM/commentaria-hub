import { useAuthStore } from '../store/authStore.ts'
import { AnnotationDetailsPane } from './AnnotationDetailsPane.tsx'
import { SuggestedRulesPane } from './SuggestedRulesPane.tsx'

export function AnnotationDetailsTab() {
  const isAuthenticated = !!useAuthStore((store) => store.token)

  return (
    <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
      <AnnotationDetailsPane />
      {isAuthenticated && <SuggestedRulesPane />}
    </div>
  )
}
