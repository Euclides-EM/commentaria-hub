import { useQueryState, parseAsString, parseAsBoolean } from 'nuqs'
import { useAnnotationIndexQuery } from '../queries/annotations'
import { LoadingSpinner } from './LoadingSpinner'
import { ErrorFallback } from './ErrorFallback'
import type { model_AnnotationIndexNode } from '../api'

interface SidebarProps {
  datasetId: string
  annotationId: string
  onPageJump: (page: number) => void
}

export function Sidebar({ datasetId, annotationId, onPageJump }: SidebarProps) {
  const [indexFilter, setIndexFilter] = useQueryState(
    'indexFilter',
    parseAsString.withDefault(''),
  )
  const [collapsed, setCollapsed] = useQueryState(
    'sidebarCollapsed',
    parseAsBoolean.withDefault(false),
  )
  const {
    data: annotationIndex,
    isLoading: loading,
    error,
    refetch: reload,
  } = useAnnotationIndexQuery(datasetId, annotationId)

  const flattenNodes = (
    nodes: model_AnnotationIndexNode[],
  ): model_AnnotationIndexNode[] => {
    const result: model_AnnotationIndexNode[] = []
    for (const node of nodes) {
      result.push(node)
      if (node.children) {
        result.push(...flattenNodes(node.children))
      }
    }
    return result
  }

  const renderIndexTree = () => {
    if (loading) {
      return <LoadingSpinner size="sm" message="Loading index..." />
    }

    if (error) {
      return (
        <ErrorFallback
          error={error}
          message="Failed to load index"
          onRetry={reload}
        />
      )
    }

    if (!annotationIndex?.nodes) {
      return (
        <div className="text-gray-500 text-sm italic text-center p-5">
          No index available
        </div>
      )
    }

    const allNodes = flattenNodes(annotationIndex.nodes)

    return (
      <div>
        {allNodes
          .filter(
            (item: model_AnnotationIndexNode) =>
              !indexFilter ||
              item.content?.toLowerCase().includes(indexFilter.toLowerCase()),
          )
          .map((item: model_AnnotationIndexNode, idx: number) => (
            <div
              key={idx}
              className="cursor-pointer py-1 px-0 border-b border-gray-200 text-xs hover:bg-black/5 transition-colors"
              onClick={() =>
                item.location?.page && onPageJump(item.location.page)
              }
            >
              {item.content}{' '}
              {item.location?.page && `(p. ${item.location.page})`}
            </div>
          ))}
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
            value={indexFilter}
            onChange={(e) => setIndexFilter(e.target.value)}
          />
          <button
            className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
            onClick={() => reload()}
          >
            Reload
          </button>
        </div>
        <div className="text-xs opacity-75 mt-1">Index loaded</div>
      </div>

      <div
        className={`overflow-auto p-2.5 pr-1.5 flex-1 min-h-0 ${collapsed ? 'hidden' : ''}`}
      >
        {renderIndexTree()}
      </div>
    </aside>
  )
}
