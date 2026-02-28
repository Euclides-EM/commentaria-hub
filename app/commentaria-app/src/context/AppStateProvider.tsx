import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { parseAsString, useQueryStates } from 'nuqs'
import { useDatasetsQuery } from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import type {
  AppState,
  AppStateContextType,
  PageOrKey,
  ViewMode,
} from './AppStateContext'
import { AppStateContext } from './AppStateContext'

interface AppStateProviderProps {
  children: ReactNode
}

export function AppStateProvider({ children }: AppStateProviderProps) {
  const [queryState, setQueryState] = useQueryStates({
    viewMode: parseAsString.withDefault(''),
    datasetId: parseAsString.withDefault(''),
    annotationId: parseAsString.withDefault(''),
    currentPageOrKey: parseAsString.withDefault(''),
  })
  const [searchResultHighlight, setSearchResultHighlight] = useState<
    string | null
  >(null)
  const [modelSearchPrefill, setModelSearchPrefill] = useState<string | null>(
    null,
  )
  const parsedViewMode: ViewMode | null =
    queryState.viewMode === 'models' ||
    queryState.viewMode === 'groundTruths' ||
    queryState.viewMode === 'jobs'
      ? queryState.viewMode
      : null
  const state = useMemo<AppState>(
    () => ({
      viewMode: parsedViewMode,
      datasetId: queryState.datasetId,
      annotationId: queryState.annotationId,
      currentPageOrKey: queryState.currentPageOrKey,
    }),
    [
      parsedViewMode,
      queryState.annotationId,
      queryState.currentPageOrKey,
      queryState.datasetId,
    ],
  )
  const { data: datasets, refetch: refetchDatasets } = useDatasetsQuery()
  const { data: annotations, refetch: refetchAnnotations } =
    useAnnotationsQuery(state.datasetId)

  const wrappedSetState = useCallback(
    (updates: Partial<AppState>) => {
      const nextUpdates: {
        viewMode?: string
        datasetId?: string
        annotationId?: string
        currentPageOrKey?: string
      } = {}

      if (updates.viewMode !== undefined) {
        nextUpdates.viewMode = updates.viewMode || ''
      }
      if (updates.datasetId !== undefined) {
        nextUpdates.datasetId = updates.datasetId
      }
      if (updates.annotationId !== undefined) {
        nextUpdates.annotationId = updates.annotationId
        if (!updates.annotationId) {
          nextUpdates.currentPageOrKey = ''
        }
      }
      if (updates.currentPageOrKey !== undefined) {
        nextUpdates.currentPageOrKey = String(updates.currentPageOrKey)
      }

      if (
        updates.datasetId !== undefined ||
        updates.annotationId !== undefined
      ) {
        nextUpdates.viewMode = ''
        setSearchResultHighlight(null)
      }
      history.pushState(state, '', window.location.href)
      setQueryState(nextUpdates)
    },
    [setQueryState, state],
  )

  const jumpToPage = useCallback(
    (nextPageOrKey: PageOrKey) => {
      setSearchResultHighlight(null)
      setQueryState({ currentPageOrKey: String(nextPageOrKey) })
    },
    [setQueryState],
  )

  const dataset = useMemo(
    () =>
      datasets?.find((d) => state.datasetId && d.id === state.datasetId) ||
      null,
    [datasets, state.datasetId],
  )

  const annotation = useMemo(
    () =>
      annotations?.find(
        (a) => state.annotationId && a.id === state.annotationId,
      ) || null,
    [annotations, state.annotationId],
  )

  const refetch = useCallback(() => {
    refetchDatasets()
    refetchAnnotations()
  }, [refetchDatasets, refetchAnnotations])

  useEffect(() => {
    if (annotations?.length === 1) {
      setQueryState((s) => ({ ...s, annotationId: annotations[0].id! }))
    }
  }, [annotations, setQueryState])

  const contextValue = useMemo<AppStateContextType>(
    () => ({
      state,
      setState: wrappedSetState,
      searchResultHighlight,
      setSearchResultHighlight,
      modelSearchPrefill,
      setModelSearchPrefill,
      dataset,
      annotation,
      jumpToPage,
      refetch,
    }),
    [
      state,
      wrappedSetState,
      searchResultHighlight,
      setSearchResultHighlight,
      modelSearchPrefill,
      setModelSearchPrefill,
      dataset,
      annotation,
      jumpToPage,
      refetch,
    ],
  )

  useEffect(() => {
    if (!import.meta.env.DEV) return
    ;(
      window as typeof window & {
        __appStateContext?: AppStateContextType
      }
    ).__appStateContext = contextValue
  }, [contextValue])

  return (
    <AppStateContext.Provider value={contextValue}>
      {children}
    </AppStateContext.Provider>
  )
}
