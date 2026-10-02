import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { parseAsString, useQueryStates } from 'nuqs'
import { useAuthStore } from '../store/authStore.ts'
import { buildAppStateUrl, getNextAppStateQueryState } from './appStateUrl'
import { useAnnotationSelection } from './useAnnotationSelection'
import type {
  AnnotationTab,
  AppState,
  AppStateContextType,
  DatasetTab,
  PageOrKey,
  ViewMode,
} from './AppStateContext'
import { AppStateContext } from './AppStateContext'

interface AppStateProviderProps {
  children: ReactNode
}

const DEFAULT_DATASET_TAB: DatasetTab = 'details'
const DEFAULT_ANNOTATION_TAB: AnnotationTab = 'details'

export function AppStateProvider({ children }: AppStateProviderProps) {
  const token = useAuthStore((store) => store.token)
  const [queryState, setQueryState] = useQueryStates({
    viewMode: parseAsString.withDefault(''),
    datasetId: parseAsString.withDefault(''),
    annotationId: parseAsString.withDefault(''),
    currentPageOrKey: parseAsString.withDefault(''),
    datasetTab: parseAsString.withDefault(''),
    annotationTab: parseAsString.withDefault(''),
    otherDatasetId: parseAsString.withDefault(''),
    otherAnnotationId: parseAsString.withDefault(''),
    otherPage: parseAsString.withDefault(''),
  })
  const [searchResultHighlight, setSearchResultHighlight] = useState<
    string | null
  >(null)
  const [modelSearchPrefill, setModelSearchPrefill] = useState<string | null>(
    null,
  )
  const parsedViewMode: ViewMode | null =
    queryState.viewMode === 'models' ||
    queryState.viewMode === 'annotations' ||
    queryState.viewMode === 'features' ||
    queryState.viewMode === 'jobs' ||
    queryState.viewMode === 'backups' ||
    queryState.viewMode === 'logs'
      ? queryState.viewMode
      : null
  const parsedDatasetTab: DatasetTab =
    queryState.datasetTab === 'annotations'
      ? 'annotations'
      : queryState.datasetTab === 'features'
        ? 'features'
        : DEFAULT_DATASET_TAB
  const parsedAnnotationTab: AnnotationTab =
    queryState.annotationTab === 'text' ||
    queryState.annotationTab === 'gallery' ||
    queryState.annotationTab === 'featureResults' ||
    queryState.annotationTab === 'featureExecutions' ||
    queryState.annotationTab === 'compare'
      ? queryState.annotationTab
      : DEFAULT_ANNOTATION_TAB
  const state = useMemo<AppState>(
    () => ({
      viewMode: parsedViewMode,
      datasetId: queryState.datasetId,
      annotationId: queryState.annotationId,
      currentPageOrKey: queryState.currentPageOrKey,
      datasetTab: parsedDatasetTab,
      annotationTab: parsedAnnotationTab,
      otherDatasetId: queryState.otherDatasetId,
      otherAnnotationId: queryState.otherAnnotationId,
      otherPage: queryState.otherPage,
    }),
    [
      parsedAnnotationTab,
      parsedDatasetTab,
      parsedViewMode,
      queryState.annotationId,
      queryState.currentPageOrKey,
      queryState.datasetId,
      queryState.otherDatasetId,
      queryState.otherAnnotationId,
      queryState.otherPage,
    ],
  )

  const wrappedSetState = useCallback(
    (updates: Partial<AppState>) => {
      const nextQueryState = getNextAppStateQueryState(queryState, updates)
      if (
        updates.datasetId !== undefined ||
        updates.annotationId !== undefined
      ) {
        setSearchResultHighlight(null)
      }
      history.pushState(state, '', window.location.href)
      setQueryState(nextQueryState)
    },
    [queryState, setQueryState, state],
  )

  const getUrlForState = useCallback(
    (updates: Partial<AppState>) =>
      buildAppStateUrl(window.location.href, queryState, updates),
    [queryState],
  )

  const jumpToPage = useCallback(
    (nextPageOrKey: PageOrKey) => {
      setSearchResultHighlight(null)
      setQueryState({ currentPageOrKey: String(nextPageOrKey) })
    },
    [setQueryState],
  )

  const setResolvedAnnotationId = useCallback(
    (annotationId: string) => setQueryState({ annotationId }),
    [setQueryState],
  )

  const setResolvedPageOrKey = useCallback(
    (currentPageOrKey: string) => setQueryState({ currentPageOrKey }),
    [setQueryState],
  )

  const { dataset, annotation, refetch } = useAnnotationSelection({
    datasetId: state.datasetId,
    annotationId: state.annotationId,
    currentPageOrKey: state.currentPageOrKey,
    onAnnotationIdResolved: setResolvedAnnotationId,
    onPageOrKeyResolved: setResolvedPageOrKey,
  })

  useEffect(() => {
    if (queryState.datasetId || !queryState.datasetTab) {
      return
    }
    setQueryState((s) => ({ ...s, datasetTab: '' }))
  }, [queryState.datasetId, queryState.datasetTab, setQueryState])

  useEffect(() => {
    if (queryState.annotationId || !queryState.annotationTab) {
      return
    }
    setQueryState((s) => ({ ...s, annotationTab: '' }))
  }, [queryState.annotationId, queryState.annotationTab, setQueryState])

  useEffect(() => {
    if (
      queryState.annotationTab === 'compare' ||
      (!queryState.otherDatasetId &&
        !queryState.otherAnnotationId &&
        !queryState.otherPage)
    ) {
      return
    }
    setQueryState((s) => ({
      ...s,
      otherDatasetId: '',
      otherAnnotationId: '',
      otherPage: '',
    }))
  }, [
    queryState.annotationTab,
    queryState.otherAnnotationId,
    queryState.otherDatasetId,
    queryState.otherPage,
    setQueryState,
  ])

  const contextValue = useMemo<AppStateContextType>(
    () => ({
      state,
      setState: wrappedSetState,
      getUrlForState,
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
      getUrlForState,
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

  useEffect(() => {
    if ((parsedViewMode !== 'backups' && parsedViewMode !== 'logs') || token) {
      return
    }
    setQueryState((s) => ({ ...s, viewMode: '' }))
  }, [parsedViewMode, setQueryState, token])

  return (
    <AppStateContext.Provider value={contextValue}>
      {children}
    </AppStateContext.Provider>
  )
}
