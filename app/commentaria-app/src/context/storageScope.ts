import { createContext, useContext } from 'react'

export const StorageScopeContext = createContext('')

export const useScopedStorageKey = (key: string) => {
  const scope = useContext(StorageScopeContext)
  return scope ? `${scope}.${key}` : key
}
