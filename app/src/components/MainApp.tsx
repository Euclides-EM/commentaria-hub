import { useAppState } from '../context/AppStateContext'
import { IndexMenu } from './IndexMenu.tsx'
import { PageNavigation } from './PageNavigation.tsx'
import { ImagePane } from './ImagePane'
import { AnnotationDetailsPane } from './AnnotationDetailsPane.tsx'
import { TeiPane } from './TeiPane.tsx'
import { ToggleButton } from './ToggleButton.tsx'
import { SuggestedRulesPane } from './SuggestedRulesPane.tsx'
import useLocalStorageState from 'use-local-storage-state'
import { useAuthStore } from '../store/authStore.ts'

export function MainApp() {
  const { annotation, state } = useAppState()
  const [showAnnotationDetails, setShowAnnotationDetails] =
    useLocalStorageState('showAnnotationDetails', { defaultValue: false })
  const [showRules, setShowRules] = useLocalStorageState('showRules', {
    defaultValue: false,
  })
  const isAuthenticated = !!useAuthStore((store) => store.token)

  if (!state.datasetId || !state.annotationId) {
    return (
      <div className="w-full m-10 font-medium text-center">
        Please select dataset and annotation
      </div>
    )
  }

  const hasOcr = annotation?.ocred
  return (
    <div className="h-full flex overflow-hidden">
      {hasOcr && <IndexMenu />}

      <main className="flex-1 flex flex-col overflow-hidden min-w-0">
        <div className="flex w-full gap-4 p-3 border-b border-gray-200 bg-white">
          <PageNavigation />
          {hasOcr && (
            <>
              <ToggleButton
                title="Annotation details"
                isOn={showAnnotationDetails}
                toggle={() => setShowAnnotationDetails((b) => !b)}
              />
              {isAuthenticated && (
                <ToggleButton
                  title="Suggested rules"
                  isOn={showRules}
                  toggle={() => setShowRules((b) => !b)}
                />
              )}
            </>
          )}
        </div>
        {hasOcr && showAnnotationDetails && <AnnotationDetailsPane />}
        {hasOcr && showRules && <SuggestedRulesPane />}

        <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
          <ImagePane />
          <TeiPane />
          {!hasOcr && (
            <div>
              <div className="flex gap-4">
                <ToggleButton
                  title="Annotation details"
                  isOn={showAnnotationDetails}
                  toggle={() => setShowAnnotationDetails((b) => !b)}
                />
                {isAuthenticated && (
                  <ToggleButton
                    title="Suggested rules"
                    isOn={showRules}
                    toggle={() => setShowRules((b) => !b)}
                  />
                )}
              </div>
              {showAnnotationDetails && <AnnotationDetailsPane />}
              {showRules && <SuggestedRulesPane />}
            </div>
          )}
        </div>
      </main>
    </div>
  )
}
