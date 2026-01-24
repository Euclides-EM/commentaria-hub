import { useAnnotationIndexQuery } from '../queries/annotations'
import { LoadingSpinner } from './LoadingSpinner'
import type { model_AnnotationIndexNode } from '../api'
import { useAppState } from '../context/AppStateContext.tsx'
import { useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'

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
}) => {
  const [isExpanded, setIsExpanded] = useState(false)
  const hasChildren = node.children && node.children.length > 0

  return (
    <div>
      <div
        className="py-1 px-0 border-b border-gray-200 text-xs hover:bg-black/5 transition-colors flex items-center"
        style={{ marginLeft: `${level * 16}px` }}
      >
        {hasChildren && (
          <button
            title={isExpanded ? 'Collapse' : 'Expand'}
            onClick={() => setIsExpanded(!isExpanded)}
            className="px-1 mr-2 hover:bg-gray-200 rounded cursor-pointer"
          >
            {isExpanded ? '▼' : '▶'}
          </button>
        )}
        <button
          onClick={() => node.location?.page && jumpToPage(node.location.page)}
          className="flex-1 text-left cursor-pointer"
        >
          {node.content} {node.location?.page && `(p. ${node.location.page})`}
        </button>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {node.children?.map((child, idx) => (
            <Node
              node={child}
              jumpToPage={jumpToPage}
              level={level + 1}
              key={idx}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function IndexMenu() {
  const { state, jumpToPage } = useAppState()
  const [searchTerm, setSearchTerm] = useLocalStorageState('indexSearch', {
    defaultValue: '',
  })
  const {
    data: annotationIndex,
    isLoading,
    error,
  } = useAnnotationIndexQuery(state.datasetId, state.annotationId)

  return (
    <>
      <div className="font-semibold text-sm px-3 pt-4 border-t border-gray-300">
        Index
      </div>
      {isLoading ? (
        <LoadingSpinner size="sm" message="Loading index..." />
      ) : error ? (
        <div className="w-full m-10 font-medium text-center text-red-800">
          {error.message}
        </div>
      ) : !annotationIndex?.nodes ? (
        <div className="text-gray-500 text-sm italic text-center p-5">
          No index available
        </div>
      ) : (
        <>
          <div className="p-3 border-b border-gray-200">
            <div className="flex gap-2 items-center">
              <input
                className="flex-1 min-w-0 border border-gray-300 rounded-lg px-3 py-2 font-mono text-xs"
                placeholder="Filter…"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
          </div>

          <div className="overflow-auto p-3 pr-1.5 flex-1 min-h-0">
            <div>
              {annotationIndex.nodes
                .filter((node: model_AnnotationIndexNode) =>
                  matchToFilter(searchTerm, node),
                )
                .map((item: model_AnnotationIndexNode, idx: number) => (
                  <Node
                    node={item}
                    jumpToPage={jumpToPage}
                    key={idx}
                    level={0}
                  />
                ))}
            </div>
          </div>
        </>
      )}
    </>
  )
}
