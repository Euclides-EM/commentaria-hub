import { useAppState } from '../context/AppStateContext'
import { IndexMenu } from './IndexMenu.tsx'
import { PageNavigation } from './PageNavigation.tsx'
import { ImagePane } from './ImagePane'
import { AnnotationDetailsPane } from './AnnotationDetailsPane.tsx'
import { TeiPane } from './TeiPane.tsx'

export function MainApp() {
  const { annotation, state } = useAppState()

  if (!state.datasetId || !state.annotationId) {
    return (
      <div className="w-full m-10 font-medium text-center">
        Please select dataset and annotation
      </div>
    )
  }

  return (
    <div className="h-full flex overflow-hidden">
      {annotation?.ocred && <IndexMenu />}

      <main className="flex-1 flex flex-col overflow-hidden min-w-0">
        <PageNavigation />

        <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
          <ImagePane />
          <TeiPane />
          <AnnotationDetailsPane />
        </div>
      </main>
    </div>
  )
}
