import { useAnnotationIndexQuery } from '../../../../queries/annotations.ts'
import { LoadingSpinner } from '../../../core/LoadingSpinner.tsx'
import { ErrorMessage } from '../../../core/ErrorMessage'
import type { annotation_IndexNode } from '@hub-api'
import { useAppState } from '../../../../context/useAppState.ts'
import { useEffect, useMemo, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { useScopedStorageKey } from '../../../../context/storageScope.ts'
import { SearchInput } from '../../../core/SearchInput.tsx'
import type { PageOrKey } from '../../../../context/AppStateContext.ts'
import { startCase } from 'lodash'

const getPageNumber = (page: string | undefined): number | undefined => {
  if (!page) return undefined
  const parsedPage = Number(page)
  return Number.isNaN(parsedPage) ? undefined : parsedPage
}

type NavigationIndexNode = Omit<annotation_IndexNode, 'children'> & {
  navigationKey: string
  children?: NavigationIndexNode[]
}

const getIndexTypeHue = (
  indexType: string | undefined,
  availableTypes: string[],
): number => {
  const typeIndex = Math.max(availableTypes.indexOf(indexType || 'default'), 0)
  return (265 + typeIndex * 137.508) % 360
}

const addNavigationKeys = (
  nodes: annotation_IndexNode[],
  parentKey = 'root',
): NavigationIndexNode[] =>
  nodes.map((node, index) => {
    const navigationKey = `${parentKey}.${index}`
    return {
      ...node,
      navigationKey,
      children: node.children
        ? addNavigationKeys(node.children, navigationKey)
        : undefined,
    }
  })

const getExpandableNodeKeys = (nodes: NavigationIndexNode[]): string[] =>
  nodes.flatMap((node) => {
    if (!node.children?.length) return []
    return [node.navigationKey, ...getExpandableNodeKeys(node.children)]
  })

const getFilteredNode = (
  search: string,
  node: NavigationIndexNode,
): NavigationIndexNode | null => {
  if (!search) return node

  const normalizedSearch = search.toLowerCase()
  const matchingChildren =
    node.children
      ?.map((child) => getFilteredNode(normalizedSearch, child))
      .filter((child): child is NavigationIndexNode => child !== null) ?? []
  const matchesSelf = node.content?.toLowerCase().includes(normalizedSearch)

  if (matchesSelf || matchingChildren.length > 0) {
    return {
      ...node,
      children: matchingChildren,
    }
  }

  return null
}

const getNextSiblingPage = (
  nodes: annotation_IndexNode[],
  startIndex: number,
  currentNodePage: number | undefined,
): number | undefined => {
  if (currentNodePage === undefined || currentNodePage === null)
    return undefined
  for (let i = startIndex + 1; i < nodes.length; i += 1) {
    const page = getPageNumber(nodes[i].location?.page)
    if (page !== undefined && page !== null && page > currentNodePage) {
      return page
    }
  }
  return undefined
}

const getActivePathKeys = (
  nodes: NavigationIndexNode[],
  currentPage: number,
): string[] => {
  const activeIndex = nodes.findIndex((node, index) => {
    const nodePage = getPageNumber(node.location?.page)
    if (nodePage === undefined || currentPage < nodePage) return false
    const nextSiblingPage = getNextSiblingPage(nodes, index, nodePage)
    return nextSiblingPage === undefined || currentPage < nextSiblingPage
  })
  const activeNode = nodes[activeIndex]
  if (!activeNode?.children?.length) return []
  return [
    activeNode.navigationKey,
    ...getActivePathKeys(activeNode.children, currentPage),
  ]
}

const Node = ({
  node,
  jumpToPage,
  level,
  currentPage,
  nextSiblingPage,
  forceExpanded,
  expandedNodeKeys,
  onExpandedChange,
  availableTypes,
}: {
  node: NavigationIndexNode
  jumpToPage: (page: PageOrKey) => void
  level: number
  currentPage: number
  nextSiblingPage?: number
  forceExpanded?: boolean
  expandedNodeKeys: Set<string>
  onExpandedChange: (nodeKey: string, expanded: boolean) => void
  availableTypes: string[]
}) => {
  const hasChildren = node.children && node.children.length > 0
  const isCurated = node.category?.startsWith('curated-heading') ?? false
  const typeHue = getIndexTypeHue(node.type, availableTypes)
  const isExpanded = forceExpanded || expandedNodeKeys.has(node.navigationKey)
  const nodePage = getPageNumber(node.location?.page)
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
        data-index-active={isActive || undefined}
      >
        {hasChildren && (
          <button
            title={isExpanded ? 'Collapse' : 'Expand'}
            aria-label={isExpanded ? 'Collapse entry' : 'Expand entry'}
            aria-expanded={isExpanded}
            onClick={() => onExpandedChange(node.navigationKey, !isExpanded)}
            className="px-1 mr-2 hover:bg-gray-200 rounded cursor-pointer"
          >
            {isExpanded ? '▼' : '▶'}
          </button>
        )}
        {!hasChildren && (
          <span className="px-1 mr-2 invisible" aria-hidden="true">
            ▶
          </span>
        )}
        <button
          onClick={() => node.location?.page && jumpToPage(node.location.page)}
          className="flex-1 text-left cursor-pointer"
        >
          {isCurated && (
            <span
              className="mr-1.5 rounded border px-1 py-0.5 text-[0.6rem] font-semibold tracking-wide"
              style={{
                borderColor: `hsl(${typeHue} 70% 80%)`,
                backgroundColor: `hsl(${typeHue} 80% 97%)`,
                color: `hsl(${typeHue} 60% 40%)`,
              }}
            >
              {startCase(node.type || 'default')}
            </span>
          )}
          <span
            style={isCurated ? { color: `hsl(${typeHue} 60% 15%)` } : undefined}
          >
            {node.content} {node.location?.page && `(p. ${node.location.page})`}
          </span>
        </button>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {node.children?.map((child, idx) => (
            <Node
              node={child}
              jumpToPage={jumpToPage}
              level={level + 1}
              currentPage={currentPage}
              nextSiblingPage={
                node.children
                  ? getNextSiblingPage(
                      node.children,
                      idx,
                      getPageNumber(child.location?.page),
                    )
                  : undefined
              }
              forceExpanded={forceExpanded}
              expandedNodeKeys={expandedNodeKeys}
              onExpandedChange={onExpandedChange}
              availableTypes={availableTypes}
              key={child.navigationKey}
            />
          ))}
        </div>
      )}
    </div>
  )
}

export function IndexMenu({
  disableHighlight,
}: {
  disableHighlight?: boolean
}) {
  const { state, jumpToPage } = useAppState()
  const [searchTerm, setSearchTerm] = useLocalStorageState(
    useScopedStorageKey('indexSearch'),
    {
      defaultValue: '',
      storageSync: false,
    },
  )
  const [expandedNodeKeys, setExpandedNodeKeys] = useState<Set<string>>(
    () => new Set(),
  )
  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const scrolledPageRef = useRef<number | null>(null)
  const [indexTypes, setIndexTypes] = useLocalStorageState<string[]>(
    useScopedStorageKey(`indexTypes.${state.datasetId}.${state.annotationId}`),
    {
      defaultValue: ['default'],
      storageSync: false,
    },
  )
  const [applyPrinterErrorCorrection, setApplyPrinterErrorCorrection] =
    useLocalStorageState(
      useScopedStorageKey('indexApplyPrinterErrorCorrection'),
      {
        defaultValue: false,
        storageSync: false,
      },
    )
  const {
    data: annotationIndex,
    isLoading,
    error,
  } = useAnnotationIndexQuery(
    state.datasetId,
    state.annotationId,
    indexTypes,
    applyPrinterErrorCorrection,
  )
  const normalizedSearchTerm = searchTerm.trim()
  const navigationNodes = useMemo(
    () =>
      addNavigationKeys(
        annotationIndex?.nodes ?? [],
        `${state.datasetId}.${state.annotationId}`,
      ),
    [annotationIndex?.nodes, state.annotationId, state.datasetId],
  )
  const expandableNodeKeys = useMemo(
    () => getExpandableNodeKeys(navigationNodes),
    [navigationNodes],
  )
  const filteredNodes = useMemo(
    () =>
      navigationNodes
        .map((node) => getFilteredNode(normalizedSearchTerm, node))
        .filter((node): node is NavigationIndexNode => node !== null),
    [navigationNodes, normalizedSearchTerm],
  )
  const currentPage = disableHighlight
    ? -1
    : Number(state.currentPageOrKey) || 0

  const [expandedActivePath, setExpandedActivePath] = useState<{
    page: number
    nodes: NavigationIndexNode[]
  } | null>(null)
  if (
    currentPage > 0 &&
    (expandedActivePath?.page !== currentPage ||
      expandedActivePath.nodes !== navigationNodes)
  ) {
    setExpandedActivePath({ page: currentPage, nodes: navigationNodes })
    const activePathKeys = getActivePathKeys(navigationNodes, currentPage)
    if (!activePathKeys.every((key) => expandedNodeKeys.has(key))) {
      setExpandedNodeKeys(new Set([...expandedNodeKeys, ...activePathKeys]))
    }
  }

  useEffect(() => {
    if (currentPage <= 0 || scrolledPageRef.current === currentPage) return
    const activePathKeys = getActivePathKeys(navigationNodes, currentPage)
    if (
      normalizedSearchTerm.length === 0 &&
      !activePathKeys.every((key) => expandedNodeKeys.has(key))
    )
      return
    const container = scrollContainerRef.current
    const activeRows = container?.querySelectorAll<HTMLElement>(
      '[data-index-active="true"]',
    )
    const activeRow = activeRows?.[activeRows.length - 1]
    if (!container || !activeRow) return
    scrolledPageRef.current = currentPage
    const containerRect = container.getBoundingClientRect()
    const rowRect = activeRow.getBoundingClientRect()
    if (
      rowRect.top >= containerRect.top &&
      rowRect.bottom <= containerRect.bottom
    )
      return
    container.scrollTop +=
      rowRect.top -
      containerRect.top -
      (container.clientHeight - rowRect.height) / 2
  }, [
    currentPage,
    navigationNodes,
    expandedNodeKeys,
    filteredNodes,
    normalizedSearchTerm,
  ])

  const allExpanded =
    expandableNodeKeys.length > 0 &&
    expandableNodeKeys.every((key) => expandedNodeKeys.has(key))
  const allCollapsed = !expandableNodeKeys.some((key) =>
    expandedNodeKeys.has(key),
  )

  const onExpandedChange = (nodeKey: string, expanded: boolean) => {
    setExpandedNodeKeys((current) => {
      const next = new Set(current)
      if (expanded) next.add(nodeKey)
      else next.delete(nodeKey)
      return next
    })
  }

  return (
    <div className="flex flex-col min-h-0 h-full">
      {(annotationIndex?.available_types?.length ?? 0) > 1 && (
        <fieldset className="mx-3 mb-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-600">
          <span className="mr-1 font-medium">Index layers</span>
          {annotationIndex?.available_types?.map((indexType) => (
            <label
              className="flex cursor-pointer items-center gap-1.5"
              key={indexType}
            >
              <input
                type="checkbox"
                checked={indexTypes.includes(indexType)}
                onChange={(event) => {
                  if (event.target.checked) {
                    setIndexTypes([...new Set([...indexTypes, indexType])])
                  } else if (indexTypes.length > 1) {
                    setIndexTypes(
                      indexTypes.filter((value) => value !== indexType),
                    )
                  }
                }}
                className="accent-violet-600"
              />
              {startCase(indexType)}
            </label>
          ))}
        </fieldset>
      )}
      {isLoading ? (
        <LoadingSpinner size="sm" message="Loading index..." />
      ) : error ? (
        <ErrorMessage error={error} variant="empty" />
      ) : !annotationIndex?.nodes?.length ? (
        <div className="text-gray-500 text-sm italic text-center p-5">
          No index available
        </div>
      ) : (
        <>
          <div className="px-3 space-y-2">
            <SearchInput
              value={searchTerm}
              onChange={setSearchTerm}
              placeholder="Filter..."
            />
            <div className="flex items-center justify-end gap-2">
              <label
                className="mr-auto flex cursor-pointer items-center gap-1 text-[0.65rem] text-gray-400 hover:text-gray-600"
                title="Show printer error corrections instead of the printed text"
              >
                <input
                  type="checkbox"
                  checked={applyPrinterErrorCorrection}
                  onChange={(event) => {
                    scrolledPageRef.current = null
                    setApplyPrinterErrorCorrection(event.target.checked)
                  }}
                  className="h-3 w-3 accent-gray-500"
                />
                Corrected
              </label>
              <button
                type="button"
                className="text-xs text-gray-600 hover:text-gray-900 disabled:text-gray-300 disabled:cursor-not-allowed"
                onClick={() => setExpandedNodeKeys(new Set())}
                disabled={allCollapsed || normalizedSearchTerm.length > 0}
                title={
                  normalizedSearchTerm
                    ? 'Clear the filter to collapse all entries'
                    : 'Collapse all index entries'
                }
              >
                Collapse all
              </button>
              <span className="text-gray-300" aria-hidden="true">
                |
              </span>
              <button
                type="button"
                className="text-xs text-gray-600 hover:text-gray-900 disabled:text-gray-300 disabled:cursor-not-allowed"
                onClick={() => setExpandedNodeKeys(new Set(expandableNodeKeys))}
                disabled={
                  expandableNodeKeys.length === 0 ||
                  allExpanded ||
                  normalizedSearchTerm.length > 0
                }
                title={
                  normalizedSearchTerm
                    ? 'Clear the filter to expand all entries'
                    : 'Expand all index entries'
                }
              >
                Expand all
              </button>
            </div>
          </div>

          <div
            ref={scrollContainerRef}
            className="overflow-auto p-3 flex-1 min-h-0"
          >
            {filteredNodes.length === 0 ? (
              <div className="text-gray-500 text-xs italic text-center py-6">
                No matching entries
              </div>
            ) : (
              <div>
                {filteredNodes.map((item, idx) => (
                  <Node
                    node={item}
                    jumpToPage={jumpToPage}
                    key={item.navigationKey}
                    level={0}
                    currentPage={currentPage}
                    forceExpanded={normalizedSearchTerm.length > 0}
                    expandedNodeKeys={expandedNodeKeys}
                    onExpandedChange={onExpandedChange}
                    availableTypes={annotationIndex?.available_types ?? []}
                    nextSiblingPage={getNextSiblingPage(
                      filteredNodes,
                      idx,
                      getPageNumber(item.location?.page),
                    )}
                  />
                ))}
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}
