import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { model_Annotation, model_Dataset } from '../api'
import { useDatasetsQuery } from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import { parseAsInteger, parseAsString, useQueryStates } from 'nuqs'

export interface AppState {
  datasetId: string
  annotationId: string
  currentPage: number
}

interface AppStateContextType {
  dataset: model_Dataset | null
  annotation: model_Annotation | null
  state: AppState
  setState: (updates: Partial<AppState>) => void
  jumpToPage: (nextPage: number) => void
  searchResultHighlight: string | null
  setSearchResultHighlight: (id: string | null) => void
  refetch: () => void
}

const AppStateContext = createContext<AppStateContextType | null>(null)

export const useAppState = () => {
  const context = useContext(AppStateContext)
  if (!context) {
    if (import.meta.env.DEV) {
      const fallback = (
        window as typeof window & {
          __appStateContext?: AppStateContextType
        }
      ).__appStateContext
      if (fallback) return fallback
    }
    throw new Error('useAppState must be used within an AppStateProvider')
  }
  return context
}

interface AppStateProviderProps {
  children: ReactNode
}

export function AppStateProvider({ children }: AppStateProviderProps) {
  const [state, setState] = useQueryStates({
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

  const wrappedSetState = (updates: Partial<AppState>) => {
    history.pushState(state, '', window.location.href)
    setSearchResultHighlight(null)
    setState(updates)
  }
  const jumpToPage = (newPage: number) => {
    wrappedSetState({ currentPage: Math.max(0, newPage) })
  }

  const contextValue: AppStateContextType = {
    state,
    setState: (updates: Partial<AppState>) => {
      history.pushState(state, '', window.location.href)
      setSearchResultHighlight(null)
      setState(updates)
    },
    searchResultHighlight,
    setSearchResultHighlight,
    dataset:
      datasets?.find((d) => state.datasetId && d.id === state.datasetId) ||
      null,
    annotation:
      annotations?.find(
        (a) => state.annotationId && a.id === state.annotationId,
      ) || null,
    jumpToPage,
    refetch: () => {
      refetchDatasets()
      refetchAnnotations()
    },
  }

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
