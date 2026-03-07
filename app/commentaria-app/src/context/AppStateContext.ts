import { createContext } from 'react'
import type { annotation_Annotation, model_Dataset } from '@hub-api'

export type ViewMode = 'models' | 'annotations' | 'jobs' | 'backups'
export type PageOrKey = number | string

export interface AppState {
  viewMode: ViewMode | null
  datasetId: string
  annotationId: string
  currentPageOrKey: PageOrKey
}

export interface AppStateContextType {
  dataset: model_Dataset | null
  annotation: annotation_Annotation | null
  state: AppState
  setState: (updates: Partial<AppState>) => void
  jumpToPage: (nextPage: PageOrKey) => void
  searchResultHighlight: string | null
  setSearchResultHighlight: (id: string | null) => void
  modelSearchPrefill: string | null
  setModelSearchPrefill: (value: string | null) => void
  refetch: () => void
}

export const AppStateContext = createContext<AppStateContextType | null>(null)
