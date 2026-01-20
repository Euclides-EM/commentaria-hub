import { parseAsBoolean, parseAsString, useQueryState } from 'nuqs'
import { useAnnotationIndexQuery } from '../queries/annotations'
import { LoadingSpinner } from './LoadingSpinner'
import type { model_AnnotationIndexNode } from '../api'
import { useAppState } from '../context/AppStateContext.tsx'

const matchToFilter = (
  search: string,
  node: model_AnnotationIndexNode,
): boolean => {
  return (
    !search ||
    node.content?.toLowerCase().includes(search.toLowerCase()) ||
    node.children?.some((child) => matchToFilter(search, child)) ||
    false
  )
}

const Node = ({
  node,
  jumpToPage,
  level,
}: {
  node: model_AnnotationIndexNode
  jumpToPage: (page: number) => void
  level: number
}) => (
  <div
    className={`cursor-pointer py-1 px-0 border-b border-gray-200 text-xs hover:bg-black/5 transition-colors ml-[${level * 12}px]`}
  >
    <button
      onClick={() => node.location?.page && jumpToPage(node.location.page)}
    >
      {node.content} {node.location?.page && `(p. ${node.location.page})`}
    </button>
    {node.children?.map((child) => (
      <Node node={child} jumpToPage={jumpToPage} level={level + 1} />
    ))}
  </div>
)

export function IndexMenu() {
  const { state, jumpToPage } = useAppState()

  // TODO - to local storage instead of query
  const [searchTerm, setSearchTerm] = useQueryState(
    'indexSearch',
    parseAsString.withDefault(''),
  )
  const [collapsed, setCollapsed] = useQueryState(
    'sidebarCollapsed',
    parseAsBoolean.withDefault(false),
  )
  const {
    data: annotationIndex,
    isLoading,
    error,
  } = useAnnotationIndexQuery(state.datasetId, state.annotationId)

  if (isLoading) {
    return <LoadingSpinner size="sm" message="Loading index..." />
  }

  if (error) {
    return (
      <div className="w-full m-10 font-medium text-center text-red-800">
        {error.message}
      </div>
    )
  }

  if (!annotationIndex?.nodes) {
    return (
      <div className="text-gray-500 text-sm italic text-center p-5">
        No index available
      </div>
    )
  }

  return (
    <aside
      className={`${collapsed ? 'w-11 min-w-11 max-w-11' : 'w-80 min-w-80 max-w-80'} border-r border-gray-200 flex flex-col overflow-hidden bg-white transition-all duration-200`}
    >
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div className={`font-semibold ${collapsed ? 'hidden' : ''}`}>
          Index
        </div>
        <button
          className={`px-2.5 py-1.5 ${collapsed ? 'rotate-180' : ''} transition-transform`}
          onClick={() => setCollapsed(!collapsed)}
          title="Minimize index"
          aria-label="Minimize index"
        >
          ⟨
        </button>
      </div>

      <div
        className={`p-2.5 border-b border-gray-200 ${collapsed ? 'hidden' : ''}`}
      >
        <div className="flex gap-2 items-center">
          <input
            className="flex-1 min-w-0 border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs"
            placeholder="Filter…"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
      </div>

      <div
        className={`overflow-auto p-2.5 pr-1.5 flex-1 min-h-0 ${collapsed ? 'hidden' : ''}`}
      >
        <div>
          {annotationIndex.nodes
            .filter((node: model_AnnotationIndexNode) =>
              matchToFilter(searchTerm, node),
            )
            .map((item: model_AnnotationIndexNode, idx: number) => (
              <Node node={item} jumpToPage={jumpToPage} key={idx} level={0} />
            ))}
        </div>
      </div>
    </aside>
  )
}
