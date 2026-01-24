import { createContext } from 'react'
import type { model_Annotation, model_Dataset } from '../api'

export interface AppState {
  datasetId: string
  annotationId: string
  currentPage: number
}

export interface AppStateContextType {
  dataset: model_Dataset | null
  annotation: model_Annotation | null
  state: AppState
  setState: (updates: Partial<AppState>) => void
  jumpToPage: (nextPage: number) => void
  searchResultHighlight: string | null
  setSearchResultHighlight: (id: string | null) => void
  refetch: () => void
}

export const AppStateContext = createContext<AppStateContextType | null>(null)
