import type { AppState } from './AppStateContext'

export type AppStateQueryState = {
  viewMode: string
  datasetId: string
  annotationId: string
  currentPageOrKey: string
  datasetTab: string
  annotationTab: string
}

const APP_STATE_QUERY_KEYS: Array<keyof AppStateQueryState> = [
  'viewMode',
  'datasetId',
  'annotationId',
  'currentPageOrKey',
  'datasetTab',
  'annotationTab',
]

export const getNextAppStateQueryState = (
  queryState: AppStateQueryState,
  updates: Partial<AppState>,
): AppStateQueryState => {
  const nextQueryState = { ...queryState }

  if (updates.viewMode !== undefined) {
    nextQueryState.viewMode = updates.viewMode || ''
  }
  if (updates.datasetId !== undefined) {
    nextQueryState.datasetId = updates.datasetId
    if (updates.currentPageOrKey === undefined) {
      nextQueryState.currentPageOrKey = ''
    }
  }
  if (updates.annotationId !== undefined) {
    nextQueryState.annotationId = updates.annotationId
    if (updates.currentPageOrKey === undefined && updates.annotationId === '') {
      nextQueryState.currentPageOrKey = ''
    }
  }
  if (updates.currentPageOrKey !== undefined) {
    nextQueryState.currentPageOrKey = String(updates.currentPageOrKey)
  }
  if (updates.datasetTab !== undefined) {
    nextQueryState.datasetTab = updates.datasetTab
  }
  if (updates.annotationTab !== undefined) {
    nextQueryState.annotationTab = updates.annotationTab
  }

  if (updates.datasetId !== undefined || updates.annotationId !== undefined) {
    nextQueryState.viewMode = ''
  }
  if (updates.datasetId === '') {
    nextQueryState.datasetTab = ''
    nextQueryState.annotationTab = ''
  }
  if (updates.annotationId === '') {
    nextQueryState.annotationTab = ''
  }

  return nextQueryState
}

export const buildAppStateUrl = (
  currentUrl: string,
  queryState: AppStateQueryState,
  updates: Partial<AppState>,
): string => {
  const url = new URL(currentUrl, window.location.origin)
  const nextQueryState = getNextAppStateQueryState(queryState, updates)

  APP_STATE_QUERY_KEYS.forEach((key) => {
    const value = nextQueryState[key]
    if (value) {
      url.searchParams.set(key, value)
      return
    }
    url.searchParams.delete(key)
  })

  return url.toString()
}
