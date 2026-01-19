import { useEffect } from 'react'
import {
  useQueryStates,
  parseAsString,
  parseAsBoolean,
  parseAsInteger,
  parseAsFloat,
} from 'nuqs'
import { usePageDataQuery } from './queries/pageData'
import { useAuthStore } from './store/authStore'
import { AuthForm } from './components/AuthForm'
import { Sidebar } from './components/Sidebar'
import { TopBar } from './components/TopBar'
import { ImagePane } from './components/ImagePane'
import { TeiPaneWrapper } from './components/TeiPaneWrapper'

function App() {
  const { token, username } = useAuthStore()

  if (!token || !username) {
    return <AuthForm />
  }

  return <AuthenticatedApp />
}

function AuthenticatedApp() {
  const [state, setState] = useQueryStates({
    dataset: parseAsString.withDefault(''),
    annotation: parseAsString.withDefault(''),
    onlyTranscribed: parseAsBoolean.withDefault(true),
    page: parseAsInteger.withDefault(89),
    showDetails: parseAsBoolean.withDefault(false),
    showSource: parseAsBoolean.withDefault(false),
    minCert: parseAsFloat.withDefault(0.8),
    teiInput: parseAsString.withDefault(''),
    renderedTei: parseAsString.withDefault(
      '<div class="text-gray-500 text-sm italic text-center p-5">Click Load.</div>',
    ),
    reqSummary: parseAsString.withDefault(''),
  })

  const { pageData, isLoading, error, refetch } = usePageDataQuery(
    state.dataset,
    state.annotation,
    state.page,
  )

  useEffect(() => {
    if (pageData.teiData && state.teiInput !== pageData.teiData) {
      setState({ teiInput: pageData.teiData })
      renderTei(pageData.teiData)
    }
  }, [pageData.teiData])

  useEffect(() => {
    if (!state.dataset || !state.annotation) {
      setState({ reqSummary: '' })
      return
    }

    if (isLoading) {
      setState({
        reqSummary: `Loading dataset ${state.dataset}, annotation ${state.annotation}, page ${state.page}...`,
      })
    } else if (error) {
      setState({ reqSummary: `Error loading page ${state.page}` })
    } else if (
      pageData.imgStatus === 'Loaded' &&
      pageData.teiStatus === 'Loaded'
    ) {
      setState({ reqSummary: `Loaded page ${state.page}` })
    }
  }, [
    state.dataset,
    state.annotation,
    state.page,
    isLoading,
    error,
    pageData.imgStatus,
    pageData.teiStatus,
  ])

  const handleLoadPage = () => {
    refetch()
  }

  const renderTei = (teiData: string) => {
    const escapeHtml = (text: string) => {
      const div = document.createElement('div')
      div.textContent = text
      return div.innerHTML
    }
    setState({
      renderedTei: `<pre class="font-mono text-sm leading-relaxed">${escapeHtml(teiData)}</pre>`,
    })
  }

  const jumpToPage = (newPage: number) => {
    setState({ page: Math.max(0, newPage) })
  }

  const handleToggleAnnotationDetails = () => {
    setState({
      showDetails: !state.showDetails,
      showSource: false,
    })
  }

  const handleToggleTeiSource = () => {
    setState({
      showSource: !state.showSource,
      showDetails: false,
    })
  }

  return (
    <div className="h-screen flex overflow-hidden">
      <Sidebar
        datasetId={state.dataset}
        annotationId={state.annotation}
        onPageJump={jumpToPage}
      />

      <main className="flex-1 flex flex-col overflow-hidden min-w-0">
        <TopBar
          selectedDatasetId={state.dataset}
          onDatasetChange={(id) => setState({ dataset: id })}
          selectedAnnotationId={state.annotation}
          onAnnotationChange={(id) => setState({ annotation: id })}
          onlyTranscribed={state.onlyTranscribed}
          onOnlyTranscribedChange={(checked) =>
            setState({ onlyTranscribed: checked })
          }
          pageNum={state.page}
          onPageNumChange={(page) => setState({ page })}
          onPrevPage={() => jumpToPage(state.page - 1)}
          onNextPage={() => jumpToPage(state.page + 1)}
          onLoad={handleLoadPage}
          reqSummary={state.reqSummary}
        />

        <div className="flex-1 min-h-0 grid grid-cols-2 gap-3 p-3 box-border overflow-hidden">
          <ImagePane
            imageUrl={pageData.imageUrl}
            imgStatus={pageData.imgStatus}
          />

          <TeiPaneWrapper
            datasetId={state.dataset}
            onlyTranscribed={state.onlyTranscribed}
            teiStatus={pageData.teiStatus}
            showAnnotationDetails={state.showDetails}
            onToggleAnnotationDetails={handleToggleAnnotationDetails}
            showTeiSource={state.showSource}
            onToggleTeiSource={handleToggleTeiSource}
            minCert={state.minCert}
            onMinCertChange={(value) => setState({ minCert: value })}
            teiInput={state.teiInput}
            onTeiInputChange={(value) => setState({ teiInput: value })}
            onRenderTei={() => renderTei(state.teiInput)}
            selectedAnnotationId={state.annotation}
            renderedTei={state.renderedTei}
          />
        </div>
      </main>
    </div>
  )
}

export default App
