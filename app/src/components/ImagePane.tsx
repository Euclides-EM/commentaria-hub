import { LoadingSpinner } from './LoadingSpinner.tsx'
import { useDatasetPageImageQuery } from '../queries/datasets.ts'
import { useAppState } from '../context/AppStateContext.tsx'

export function ImagePane() {
  const {
    state: { datasetId, currentPage },
  } = useAppState()
  const { data, isLoading, error } = useDatasetPageImageQuery(
    datasetId,
    currentPage,
  )

  if (isLoading) {
    return <LoadingSpinner size="sm" message="Loading image..." />
  }
  if (error || !data) {
    return (
      <div className="w-full m-10 font-medium text-center text-red-800">
        Error: {error?.message || 'Failed to fetch image'}
      </div>
    )
  }

  const imageUrl = URL.createObjectURL(new Blob([data], { type: 'image/png' }))
  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Page {currentPage} Facsimile</div>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden flex items-center justify-center p-0">
        <img
          id="pageImg"
          alt="Page image"
          className="max-w-full max-h-full w-auto h-auto object-contain block"
          src={imageUrl}
        />
      </div>
    </section>
  )
}
