import { type ReactNode, useCallback, useMemo, useState } from 'react'
import { parseAsString, useQueryStates } from 'nuqs'
import type {
  AppState,
  AppStateContextType,
  PageOrKey,
} from './AppStateContext'
import { AppStateContext } from './AppStateContext'
import { useAppState } from './useAppState'
import { useAnnotationSelection } from './useAnnotationSelection'

interface CompareSideProviderProps {
  children: ReactNode
}

const toOtherUpdates = (updates: Partial<AppState>): Partial<AppState> => {
  const { datasetId, annotationId, currentPageOrKey, ...rest } = updates
  const otherUpdates: Partial<AppState> = { ...rest }
  if (datasetId !== undefined) {
    otherUpdates.otherDatasetId = datasetId
  }
  if (annotationId !== undefined) {
    otherUpdates.otherAnnotationId = annotationId
  }
  if (currentPageOrKey !== undefined) {
    otherUpdates.otherPage = currentPageOrKey
  }
  return otherUpdates
}

export function CompareSideProvider({ children }: CompareSideProviderProps) {
  const parent = useAppState()
  const {
    state: parentState,
    setState: parentSetState,
    getUrlForState: parentGetUrlForState,
  } = parent
  const [, setOtherQueryState] = useQueryStates({
    otherAnnotationId: parseAsString.withDefault(''),
    otherPage: parseAsString.withDefault(''),
  })
  const [searchResultHighlight, setSearchResultHighlight] = useState<
    string | null
  >(null)

  const state = useMemo<AppState>(
    () => ({
      ...parentState,
      datasetId: parentState.otherDatasetId,
      annotationId: parentState.otherAnnotationId,
      currentPageOrKey: parentState.otherPage,
    }),
    [parentState],
  )

  const setState = useCallback(
    (updates: Partial<AppState>) => {
      if (
        updates.datasetId !== undefined ||
        updates.annotationId !== undefined
      ) {
        setSearchResultHighlight(null)
      }
      parentSetState(toOtherUpdates(updates))
    },
    [parentSetState],
  )

  const getUrlForState = useCallback(
    (updates: Partial<AppState>) =>
      parentGetUrlForState(toOtherUpdates(updates)),
    [parentGetUrlForState],
  )

  const jumpToPage = useCallback(
    (nextPageOrKey: PageOrKey) => {
      setSearchResultHighlight(null)
      setOtherQueryState({ otherPage: String(nextPageOrKey) })
    },
    [setOtherQueryState],
  )

  const setResolvedAnnotationId = useCallback(
    (otherAnnotationId: string) => setOtherQueryState({ otherAnnotationId }),
    [setOtherQueryState],
  )

  const setResolvedPageOrKey = useCallback(
    (otherPage: string) => setOtherQueryState({ otherPage }),
    [setOtherQueryState],
  )

  const { dataset, annotation, refetch } = useAnnotationSelection({
    datasetId: state.datasetId,
    annotationId: state.annotationId,
    currentPageOrKey: state.currentPageOrKey,
    onAnnotationIdResolved: setResolvedAnnotationId,
    onPageOrKeyResolved: setResolvedPageOrKey,
  })

  const contextValue = useMemo<AppStateContextType>(
    () => ({
      ...parent,
      state,
      setState,
      getUrlForState,
      searchResultHighlight,
      setSearchResultHighlight,
      dataset,
      annotation,
      jumpToPage,
      refetch,
    }),
    [
      parent,
      state,
      setState,
      getUrlForState,
      searchResultHighlight,
      dataset,
      annotation,
      jumpToPage,
      refetch,
    ],
  )

  return (
    <AppStateContext.Provider value={contextValue}>
      {children}
    </AppStateContext.Provider>
  )
}
