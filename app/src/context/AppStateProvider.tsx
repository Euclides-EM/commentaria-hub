import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react'
import {parseAsBoolean, parseAsInteger, parseAsString, useQueryStates } from 'nuqs'
import { useDatasetsQuery } from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import type { AppState, AppStateContextType } from './AppStateContext'
import { AppStateContext } from './AppStateContext'

interface AppStateProviderProps {
  children: ReactNode
}

export function AppStateProvider({ children }: AppStateProviderProps) {
  const [state, setState] = useQueryStates({
      viewingModels: parseAsBoolean.withDefault(false),
    datasetId: parseAsString.withDefault(''),
    annotationId: parseAsString.withDefault(''),
    currentPage: parseAsInteger.withDefault(89),
  })
  const [searchResultHighlight, setSearchResultHighlight] = useState<
    string | null
  >(null)
  const { data: datasets, refetch: refetchDatasets } = useDatasetsQuery()
  const { data: annotations, refetch: refetchAnnotations } =
    useAnnotationsQuery(state.datasetId)

  const wrappedSetState = useCallback(
    (updates: Partial<AppState>) => {
      if (
        updates.datasetId !== undefined ||
        updates.annotationId !== undefined
      ) {
        updates.viewingModels = false;
        setSearchResultHighlight(null)
      }
      history.pushState(state, '', window.location.href)
      setState(updates)
    },
    [setState, state],
  )

  const jumpToPage = useCallback(
    (newPage: number) => {
      setSearchResultHighlight(null)
      setState({ currentPage: Math.max(0, newPage) })
    },
    [setState],
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

  const contextValue = useMemo<AppStateContextType>(
    () => ({
      state,
      setState: wrappedSetState,
      searchResultHighlight,
      setSearchResultHighlight,
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
