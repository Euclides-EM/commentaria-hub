import { useContext } from 'react'
import { AppStateContext } from './AppStateContext'
import type { AppStateContextType } from './AppStateContext'

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
