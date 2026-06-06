import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react'
import { parseAsString, useQueryStates } from 'nuqs'
import {
  useDatasetImageKeysQuery,
  useDatasetsQuery,
} from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import { useAuthStore } from '../store/authStore.ts'
import { parsePageEntries } from '../utils/pages.ts'
import { findMatchingImage, hasAnnotationPages } from '../utils/editions.ts'
import { buildAppStateUrl, getNextAppStateQueryState } from './appStateUrl'
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

const getDefaultPageOrKey = (availablePages: string[]): string => {
  if (!availablePages.length) return ''
  if (availablePages[0] !== '1') {
    return availablePages[0]
  }
  return availablePages[Math.floor(availablePages.length / 2)]
}

export function AppStateProvider({ children }: AppStateProviderProps) {
  const token = useAuthStore((store) => store.token)
  const [queryState, setQueryState] = useQueryStates({
    viewMode: parseAsString.withDefault(''),
    datasetId: parseAsString.withDefault(''),
    annotationId: parseAsString.withDefault(''),
    currentPageOrKey: parseAsString.withDefault(''),
    datasetTab: parseAsString.withDefault(''),
    annotationTab: parseAsString.withDefault(''),
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
    queryState.annotationTab === 'featureExecutions'
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
    }),
    [
      parsedAnnotationTab,
      parsedDatasetTab,
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
  const hasPages = hasAnnotationPages(annotation)
  const annotationPageEntries = useMemo(
    () => (annotation ? parsePageEntries(annotation.pages || '') : []),
    [annotation],
  )
  const shouldLoadImageKeys = !!annotation && !hasPages
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    state.datasetId,
    shouldLoadImageKeys,
    annotationPageEntries.length > 0 ? annotationPageEntries : null,
  )
  const availablePageOrKeys = useMemo(() => {
    if (!annotation) {
      return []
    }
    if (annotationPageEntries.length > 0) {
      return [...new Set(annotationPageEntries)].sort((a, b) =>
        a.localeCompare(b, undefined, { numeric: true }),
      )
    }
    return imageKeys.map((image) => image.key)
  }, [annotation, annotationPageEntries, imageKeys])

  const refetch = useCallback(() => {
    refetchDatasets()
    refetchAnnotations()
  }, [refetchDatasets, refetchAnnotations])

  useEffect(() => {
    if (annotations?.length === 1) {
      setQueryState((s) => ({
        ...s,
        annotationId: annotations[0].id!,
      }))
    }
  }, [annotations, setQueryState])

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
    if (!annotation || !availablePageOrKeys.length) {
      return
    }
    if (availablePageOrKeys.includes(String(state.currentPageOrKey))) {
      return
    }
    if (!hasPages) {
      const matchedImage = findMatchingImage(
        String(state.currentPageOrKey),
        imageKeys,
      )
      if (matchedImage?.key) {
        setQueryState((s) => ({
          ...s,
          currentPageOrKey: matchedImage.key,
        }))
        return
      }
    }
    setQueryState((s) => ({
      ...s,
      currentPageOrKey: getDefaultPageOrKey(availablePageOrKeys),
    }))
  }, [
    annotation,
    availablePageOrKeys,
    hasPages,
    imageKeys,
    setQueryState,
    state.currentPageOrKey,
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
