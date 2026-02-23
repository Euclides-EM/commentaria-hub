import { ImagePane } from './ImagePane.tsx'
import { TeiPane } from './tei/TeiPane.tsx'
import { AnnotationNavigation } from './AnnotationNavigation.tsx'
import { useAppState } from '../../../context/useAppState.ts'

export function AnnotationContentsTab() {
  const { annotation, dataset } = useAppState()
  const showTeiPane = !!annotation?.ocred || !!dataset?.edition_id
  return (
    <div className="h-full flex overflow-hidden">
      <AnnotationNavigation />
      <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
        <ImagePane />
        {showTeiPane && <TeiPane />}
      </div>
    </div>
  )
}
