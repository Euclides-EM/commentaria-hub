import { createContext } from 'react'
import type { annotation_Annotation, model_Dataset } from '../api'

export interface AppState {
  viewingModels: boolean
  viewingGroundTruths: boolean
  datasetId: string
  annotationId: string
  currentPage: number
}

export interface AppStateContextType {
  dataset: model_Dataset | null
  annotation: annotation_Annotation | null
  state: AppState
  setState: (updates: Partial<AppState>) => void
  jumpToPage: (nextPage: number) => void
  searchResultHighlight: string | null
  setSearchResultHighlight: (id: string | null) => void
  modelSearchPrefill: string | null
  setModelSearchPrefill: (value: string | null) => void
  refetch: () => void
}

export const AppStateContext = createContext<AppStateContextType | null>(null)
