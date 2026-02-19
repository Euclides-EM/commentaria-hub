import { useAnnotationIndexQuery } from '../../../../queries/annotations.ts'
import { LoadingSpinner } from '../../../core/LoadingSpinner.tsx'
import { ErrorMessage } from '../../../core/ErrorMessage'
import type { annotation_IndexNode } from '../../../../api'
import { useAppState } from '../../../../context/useAppState.ts'
import { useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { SearchInput } from '../../../core/SearchInput.tsx'

const matchToFilter = (search: string, node: annotation_IndexNode): boolean => {
  return (
    !search ||
    node.content?.toLowerCase().includes(search.toLowerCase()) ||
    node.children?.some((child: annotation_IndexNode) =>
      matchToFilter(search, child),
    ) ||
    false
  )
}

const getNextSiblingPage = (
  nodes: annotation_IndexNode[],
  startIndex: number,
  currentNodePage: number | undefined,
): number | undefined => {
  if (currentNodePage === undefined || currentNodePage === null)
    return undefined
  for (let i = startIndex + 1; i < nodes.length; i += 1) {
    const page = nodes[i].location?.page
    if (page !== undefined && page !== null && page > currentNodePage) {
      return page
    }
  }
  return undefined
}

const Node = ({
  node,
  jumpToPage,
  level,
  currentPage,
  nextSiblingPage,
}: {
  node: annotation_IndexNode
  jumpToPage: (page: number) => void
  level: number
  currentPage: number
  nextSiblingPage?: number
}) => {
  const [isExpanded, setIsExpanded] = useState(false)
  const hasChildren = node.children && node.children.length > 0
  const nodePage = node.location?.page
  const isActive =
    nodePage !== undefined &&
    nodePage !== null &&
    currentPage >= nodePage &&
    (nextSiblingPage === undefined || currentPage < nextSiblingPage)

  return (
    <div>
      <div
        className={`py-1 px-0 border-b border-gray-200 text-xs hover:bg-black/5 transition-colors flex items-center ${isActive ? 'bg-black/5 font-semibold' : ''}`}
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
          {node.children?.map((child: annotation_IndexNode, idx: number) => (
            <Node
              node={child}
              jumpToPage={jumpToPage}
              level={level + 1}
              currentPage={currentPage}
              nextSiblingPage={
                node.children
                  ? getNextSiblingPage(node.children, idx, child.location?.page)
                  : undefined
              }
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
    <div className="flex flex-col min-h-0 h-full">
      {isLoading ? (
        <LoadingSpinner size="sm" message="Loading index..." />
      ) : error ? (
        <ErrorMessage error={error} variant="empty" />
      ) : !annotationIndex?.nodes ? (
        <div className="text-gray-500 text-sm italic text-center p-5">
          No index available
        </div>
      ) : (
        <>
          <div className="px-3">
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Filter..."
            />
          </div>

          <div className="overflow-auto p-3 flex-1 min-h-0">
            <div>
              {annotationIndex.nodes
                .filter((node: annotation_IndexNode) =>
                  matchToFilter(searchTerm, node),
                )
                .map((item: annotation_IndexNode, idx: number) => (
                  <Node
                    node={item}
                    jumpToPage={jumpToPage}
                    key={idx}
                    level={0}
                    currentPage={state.currentPage}
                    nextSiblingPage={getNextSiblingPage(
                      annotationIndex.nodes || [],
                      idx,
                      item.location?.page,
                    )}
                  />
                ))}
            </div>
          </div>
        </>
      )}
    </div>
  )
}
