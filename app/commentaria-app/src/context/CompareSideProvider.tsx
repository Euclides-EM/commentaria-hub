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
import { StorageScopeContext } from './storageScope'
import { COMPARE_SIDE_KEYS, type CompareSide } from './compareSides'

interface CompareSideProviderProps {
  side: CompareSide
  children: ReactNode
}

const toSideUpdates = (
  side: CompareSide,
  updates: Partial<AppState>,
): Partial<AppState> => {
  const keys = COMPARE_SIDE_KEYS[side]
  const { datasetId, annotationId, currentPageOrKey, ...rest } = updates
  const sideUpdates: Partial<AppState> = { ...rest }
  if (datasetId !== undefined) {
    sideUpdates[keys.datasetId] = datasetId
  }
  if (annotationId !== undefined) {
    sideUpdates[keys.annotationId] = annotationId
  }
  if (currentPageOrKey !== undefined) {
    sideUpdates[keys.page] = currentPageOrKey
  }
  return sideUpdates
}

export function CompareSideProvider({
  side,
  children,
}: CompareSideProviderProps) {
  const parent = useAppState()
  const {
    state: parentState,
    setState: parentSetState,
    getUrlForState: parentGetUrlForState,
  } = parent
  const [, setCompareQueryState] = useQueryStates({
    leftAnnotationId: parseAsString.withDefault(''),
    leftPage: parseAsString.withDefault(''),
    rightAnnotationId: parseAsString.withDefault(''),
    rightPage: parseAsString.withDefault(''),
  })
  const [searchResultHighlight, setSearchResultHighlight] = useState<
    string | null
  >(null)

  const keys = COMPARE_SIDE_KEYS[side]
  const state = useMemo<AppState>(
    () => ({
      ...parentState,
      annotationTab: 'text',
      datasetId: parentState[keys.datasetId],
      annotationId: parentState[keys.annotationId],
      currentPageOrKey: parentState[keys.page],
    }),
    [parentState, keys],
  )

  const setState = useCallback(
    (updates: Partial<AppState>) => {
      if (
        updates.datasetId !== undefined ||
        updates.annotationId !== undefined
      ) {
        setSearchResultHighlight(null)
      }
      parentSetState(toSideUpdates(side, updates))
    },
    [parentSetState, side],
  )

  const getUrlForState = useCallback(
    (updates: Partial<AppState>) =>
      parentGetUrlForState(toSideUpdates(side, updates)),
    [parentGetUrlForState, side],
  )

  const setSidePage = useCallback(
    (page: string) =>
      setCompareQueryState(
        side === 'left' ? { leftPage: page } : { rightPage: page },
      ),
    [setCompareQueryState, side],
  )

  const jumpToPage = useCallback(
    (nextPageOrKey: PageOrKey) => {
      setSearchResultHighlight(null)
      setSidePage(String(nextPageOrKey))
    },
    [setSidePage],
  )

  const setResolvedAnnotationId = useCallback(
    (annotationId: string) =>
      setCompareQueryState(
        side === 'left'
          ? { leftAnnotationId: annotationId }
          : { rightAnnotationId: annotationId },
      ),
    [setCompareQueryState, side],
  )

  const { dataset, annotation, refetch } = useAnnotationSelection({
    datasetId: state.datasetId,
    annotationId: state.annotationId,
    currentPageOrKey: state.currentPageOrKey,
    onAnnotationIdResolved: setResolvedAnnotationId,
    onPageOrKeyResolved: setSidePage,
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
      <StorageScopeContext.Provider value={`compare.${side}`}>
        {children}
      </StorageScopeContext.Provider>
    </AppStateContext.Provider>
  )
}
