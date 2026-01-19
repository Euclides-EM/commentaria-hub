import { useAnnotationsQuery } from '../queries/annotations'
import { TeiPane } from './TeiPane'
import { LoadingSpinner } from './LoadingSpinner'
import { ErrorFallback } from './ErrorFallback'

interface TeiPaneWrapperProps {
  datasetId: string
  onlyTranscribed: boolean
  teiStatus: string
  showAnnotationDetails: boolean
  onToggleAnnotationDetails: () => void
  showTeiSource: boolean
  onToggleTeiSource: () => void
  minCert: number
  onMinCertChange: (value: number) => void
  teiInput: string
  onTeiInputChange: (value: string) => void
  onRenderTei: () => void
  selectedAnnotationId: string
  renderedTei: string
}

export function TeiPaneWrapper({
  datasetId,
  onlyTranscribed,
  ...teiProps
}: TeiPaneWrapperProps) {
  const {
    data: annotations,
    isLoading: loading,
    error,
  } = useAnnotationsQuery(datasetId, onlyTranscribed)

  if (loading) {
    return (
      <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
        <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
          <div>TEI</div>
          <div className="text-xs opacity-75">{teiProps.teiStatus}</div>
        </div>
        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          <LoadingSpinner size="sm" message="Loading annotations..." />
        </div>
      </section>
    )
  }

  if (error) {
    return (
      <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
        <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
          <div>TEI</div>
          <div className="text-xs opacity-75">Error</div>
        </div>
        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          <ErrorFallback error={error} message="Failed to load annotations" />
        </div>
      </section>
    )
  }

  return <TeiPane {...teiProps} annotations={annotations || []} />
}
