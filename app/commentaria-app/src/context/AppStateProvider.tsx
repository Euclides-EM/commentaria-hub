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
import { expandRange } from '../utils/pages.ts'
import { TITLE_PAGES_DATASET_ID } from '../utils/editions.ts'
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
    queryState.viewMode === 'jobs' ||
    queryState.viewMode === 'backups'
      ? queryState.viewMode
      : null
  const parsedDatasetTab: DatasetTab =
    queryState.datasetTab === 'features' ? 'features' : DEFAULT_DATASET_TAB
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
      const nextUpdates: {
        viewMode?: string
        datasetId?: string
        annotationId?: string
        currentPageOrKey?: string
        datasetTab?: string
        annotationTab?: string
      } = {}

      if (updates.viewMode !== undefined) {
        nextUpdates.viewMode = updates.viewMode || ''
      }
      if (updates.datasetId !== undefined) {
        nextUpdates.datasetId = updates.datasetId
        if (updates.currentPageOrKey === undefined) {
          nextUpdates.currentPageOrKey = ''
        }
      }
      if (updates.annotationId !== undefined) {
        nextUpdates.annotationId = updates.annotationId
        if (
          updates.currentPageOrKey === undefined &&
          updates.annotationId === ''
        ) {
          nextUpdates.currentPageOrKey = ''
        }
      }
      if (updates.currentPageOrKey !== undefined) {
        nextUpdates.currentPageOrKey = String(updates.currentPageOrKey)
      }
      if (updates.datasetTab !== undefined) {
        nextUpdates.datasetTab = updates.datasetTab
      }
      if (updates.annotationTab !== undefined) {
        nextUpdates.annotationTab = updates.annotationTab
      }

      if (
        updates.datasetId !== undefined ||
        updates.annotationId !== undefined
      ) {
        nextUpdates.viewMode = ''
        setSearchResultHighlight(null)
      }
      if (updates.datasetId === '') {
        nextUpdates.datasetTab = ''
        nextUpdates.annotationTab = ''
      }
      if (updates.annotationId === '') {
        nextUpdates.annotationTab = ''
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
  const isTitlePagesDataset = annotation?.dataset_id === TITLE_PAGES_DATASET_ID
  const shouldLoadImageKeys =
    !!annotation && (isTitlePagesDataset || !annotation.pages)
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    state.datasetId,
    shouldLoadImageKeys,
    annotation?.pages && !isTitlePagesDataset
      ? annotation.pages.split(',')
      : null,
  )
  const availablePageOrKeys = useMemo(() => {
    if (!annotation) {
      return []
    }
    if (annotation.pages && !isTitlePagesDataset) {
      const pages = annotation.pages.split(',').flatMap((p) => expandRange(p))
      return [...new Set(pages)].sort((a, b) =>
        a.localeCompare(b, undefined, { numeric: true }),
      )
    }
    return imageKeys.map((image) => image.key)
  }, [annotation, imageKeys, isTitlePagesDataset])

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
    setQueryState((s) => ({
      ...s,
      currentPageOrKey: getDefaultPageOrKey(availablePageOrKeys),
    }))
  }, [annotation, availablePageOrKeys, setQueryState, state.currentPageOrKey])

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

  useEffect(() => {
    if (parsedViewMode !== 'backups' || token) {
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
