import { ImagePane } from './ImagePane'
import { TeiPane } from './TeiPane.tsx'
import { AnnotationNavigation } from './AnnotationNavigation.tsx'

export function AnnotationTextTab() {
  return (
    <div className="h-full flex overflow-hidden">
      <AnnotationNavigation />
      <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
        <ImagePane />
        <TeiPane />
      </div>
    </div>
  )
}
