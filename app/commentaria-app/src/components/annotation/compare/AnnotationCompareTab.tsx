import { AnnotationContentsTab } from '../contents/AnnotationContentsTab.tsx'
import { CompareSideProvider } from '../../../context/CompareSideProvider.tsx'
import { useAppState } from '../../../context/useAppState.ts'

export function AnnotationCompareTab() {
  const { state } = useAppState()

  return (
    <div className="h-full w-full flex overflow-hidden">
      <div className="flex-1 min-w-0 h-full overflow-hidden">
        <AnnotationContentsTab />
      </div>
      <div
        role="separator"
        aria-orientation="vertical"
        className="w-1.5 shrink-0 bg-gray-800 rounded-full mx-1"
      />
      <div className="flex-1 min-w-0 h-full overflow-hidden">
        {state.otherDatasetId && state.otherAnnotationId ? (
          <CompareSideProvider>
            <AnnotationContentsTab />
          </CompareSideProvider>
        ) : (
          <div className="h-full flex items-center justify-center p-6 text-center font-medium text-gray-600">
            Select a dataset and an annotation to compare with, using the
            controls in the top bar.
          </div>
        )}
      </div>
    </div>
  )
}
