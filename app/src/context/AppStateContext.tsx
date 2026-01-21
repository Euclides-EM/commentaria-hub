import { createContext, type ReactNode, useContext } from 'react'
import { type AnnotationFilter } from '../components/AnnotationFilterDropdown'
import type { model_Annotation, model_Dataset } from '../api'
import { useDatasetsQuery } from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import { parseAsArrayOf, parseAsInteger, parseAsString, useQueryStates, } from 'nuqs'

export interface AppState {
  datasetId: string
  annotationId: string
  annotationFilters: AnnotationFilter[]
  currentPage: number
}

interface AppStateContextType {
  dataset: model_Dataset | null
  annotation: model_Annotation | null
  state: AppState
  setState: (updates: Partial<AppState>) => void
  toggleFilter: (filter: AnnotationFilter) => void
  jumpToPage: (nextPage: number) => void
  refetch: () => void
}

const AppStateContext = createContext<AppStateContextType | undefined>(
  undefined,
)

export const useAppState = () => {
  const context = useContext(AppStateContext)
  if (context === undefined) {
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
    annotationFilters: parseAsArrayOf(parseAsString).withDefault([]),
    currentPage: parseAsInteger.withDefault(89),
  })
  const { data: datasets, refetch: refetchDatasets } = useDatasetsQuery()
  const { data: annotations, refetch: refetchAnnotations } =
    useAnnotationsQuery(
      state.datasetId,
      state.annotationFilters as AnnotationFilter[],
    )

  const toggleFilter = (filter: AnnotationFilter) => {
    const currentFilters = state.annotationFilters as AnnotationFilter[]
    if (currentFilters.includes(filter)) {
      setState({
        annotationFilters: currentFilters.filter((f) => f !== filter),
      })
    } else {
      setState({ annotationFilters: [...currentFilters, filter] })
    }
  }

  const jumpToPage = (newPage: number) => {
    setState({ currentPage: Math.max(0, newPage) })
  }

  const contextValue: AppStateContextType = {
    state: {
      ...state,
      annotationFilters: state.annotationFilters as AnnotationFilter[],
    },
    setState,
    dataset:
      datasets?.find((d) => state.datasetId && d.id === state.datasetId) ||
      null,
    annotation:
      annotations?.find(
        (a) => state.annotationId && a.id === state.annotationId,
      ) || null,
    toggleFilter,
    jumpToPage,
    refetch: () => {
      refetchDatasets()
      refetchAnnotations()
    },
  }

  return (
    <AppStateContext.Provider value={contextValue}>
      {children}
    </AppStateContext.Provider>
  )
}
