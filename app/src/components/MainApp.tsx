import { useMemo, useState } from 'react'
import { usePageDataQuery } from '../queries/pageData'
import { useAppState } from '../context/AppStateContext'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'
import { ImagePane } from './ImagePane'
import { TeiPaneWrapper } from './TeiPaneWrapper'

export function MainApp() {
  const {
    state,
    setState,
    jumpToPage,
    toggleAnnotationDetails,
    toggleTeiSource,
  } = useAppState()
  const [teiInput, setTeiInput] = useState<string | null>(null)

  const { pageData, isLoading, error, refetch } = usePageDataQuery(
    state.dataset,
    state.annotation,
    state.page,
  )

  const reqSummary = useMemo(() => {
    if (!state.dataset || !state.annotation) {
      return ''
    }
    if (isLoading) {
      return `Loading dataset ${state.dataset}, annotation ${state.annotation}, page ${state.page}...`
    }
    if (error) {
      return `Error loading page ${state.page}`
    }
    if (pageData.imgStatus === 'Loaded' && pageData.teiStatus === 'Loaded') {
      return `Loaded page ${state.page}`
    }
    return ''
  }, [
    error,
    isLoading,
    pageData.imgStatus,
    pageData.teiStatus,
    state.annotation,
    state.dataset,
    state.page,
  ])

  if (
    teiInput === null &&
    state.dataset &&
    state.annotation &&
    !isLoading &&
    !error &&
    pageData.imgStatus === 'Loaded' &&
    pageData.teiStatus === 'Loaded'
  ) {
    setTeiInput(pageData.teiData)
  }

  const handleLoadPage = () => {
    refetch()
  }

  return (
    <div className="h-full flex overflow-hidden">
      <Sidebar
        datasetId={state.dataset}
        annotationId={state.annotation}
        onPageJump={jumpToPage}
      />

      <main className="flex-1 flex flex-col overflow-hidden min-w-0">
        <TopBar
          pageNum={state.page}
          onPageNumChange={(page) => setState({ page })}
          onPrevPage={() => jumpToPage(state.page - 1)}
          onNextPage={() => jumpToPage(state.page + 1)}
          onLoad={handleLoadPage}
          reqSummary={reqSummary}
        />

        <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
          <ImagePane
            imageUrl={pageData.imageUrl}
            imgStatus={pageData.imgStatus}
          />

          <TeiPaneWrapper
            datasetId={state.dataset}
            onlyTranscribed={false}
            teiStatus={pageData.teiStatus}
            showAnnotationDetails={state.showDetails}
            onToggleAnnotationDetails={toggleAnnotationDetails}
            showTeiSource={state.showSource}
            onToggleTeiSource={toggleTeiSource}
            minCert={state.minCert}
            onMinCertChange={(value) => setState({ minCert: value })}
            teiInput={teiInput || ''}
            onTeiInputChange={setTeiInput}
            selectedAnnotationId={state.annotation}
          />
        </div>
      </main>
    </div>
  )
}
