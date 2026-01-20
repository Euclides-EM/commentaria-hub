import { useAppState } from '../context/AppStateContext'
import { IndexMenu } from './IndexMenu.tsx'
import { PageNavigation } from './PageNavigation.tsx'
import { ImagePane } from './ImagePane'
import { AnnotationDetailsPane } from './AnnotationDetailsPane.tsx'
import { TeiPane } from './TeiPane.tsx'
import { useState } from 'react'
import { ToggleButton } from './ToggleButton.tsx'

export function MainApp() {
  const { annotation, state } = useAppState()
  // TODO - local storage
  const [showAnnotationDetails, setShowAnnotationDetails] = useState(false)
  const [showRules, setShowRules] = useState(false)

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
        <div className="flex w-full gap-4 p-3 border-b border-gray-200 bg-white">
          <PageNavigation />
          <ToggleButton
            title="annotation details"
            isOn={showAnnotationDetails}
            toggle={() => setShowAnnotationDetails((b) => !b)}
          />
          <ToggleButton
            title="suggested rules"
            isOn={showRules}
            toggle={() => setShowRules((b) => !b)}
          />
        </div>
        {showAnnotationDetails && <AnnotationDetailsPane />}

        <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
          <ImagePane />
          <TeiPane />
        </div>
      </main>
    </div>
  )
}
