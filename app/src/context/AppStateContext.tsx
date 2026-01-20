import { createContext, useContext, ReactNode } from 'react'
import {
  parseAsBoolean,
  parseAsFloat,
  parseAsInteger,
  parseAsString,
  parseAsArrayOf,
  useQueryStates,
} from 'nuqs'
import { type AnnotationFilter } from '../components/AnnotationFilterDropdown'

export interface AppState {
  dataset: string
  annotation: string
  annotationFilters: string[]
  page: number
  showDetails: boolean
  showSource: boolean
  minCert: number
}

interface AppStateContextType {
  state: AppState
  setState: (updates: Partial<AppState>) => void
  toggleFilter: (filter: AnnotationFilter) => void
  jumpToPage: (newPage: number) => void
  toggleAnnotationDetails: () => void
  toggleTeiSource: () => void
}

const AppStateContext = createContext<AppStateContextType | undefined>(
  undefined,
)

export function useAppState() {
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
    dataset: parseAsString.withDefault(''),
    annotation: parseAsString.withDefault(''),
    annotationFilters: parseAsArrayOf(parseAsString).withDefault([]),
    page: parseAsInteger.withDefault(89),
    showDetails: parseAsBoolean.withDefault(false),
    showSource: parseAsBoolean.withDefault(false),
    minCert: parseAsFloat.withDefault(0.8),
  })

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
    setState({ page: Math.max(0, newPage) })
  }

  const toggleAnnotationDetails = () => {
    setState({
      showDetails: !state.showDetails,
      showSource: false,
    })
  }

  const toggleTeiSource = () => {
    setState({
      showSource: !state.showSource,
      showDetails: false,
    })
  }

  const contextValue: AppStateContextType = {
    state: {
      dataset: state.dataset,
      annotation: state.annotation,
      annotationFilters: state.annotationFilters,
      page: state.page,
      showDetails: state.showDetails,
      showSource: state.showSource,
      minCert: state.minCert,
    },
    setState,
    toggleFilter,
    jumpToPage,
    toggleAnnotationDetails,
    toggleTeiSource,
  }

  return (
    <AppStateContext.Provider value={contextValue}>
      {children}
    </AppStateContext.Provider>
  )
}
