import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import TimeAgo from 'javascript-time-ago'
import en from 'javascript-time-ago/locale/en'
import { NuqsAdapter } from 'nuqs/adapters/react'
import { initializeAPI } from './config/api'
import './index.css'
import { App } from './App.tsx'
import { TopLevelErrorBoundary } from './components/core/TopLevelErrorBoundary'

TimeAgo.addDefaultLocale(en)

initializeAPI()

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      gcTime: 10 * 60 * 1000,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <TopLevelErrorBoundary>
      <NuqsAdapter>
        <QueryClientProvider client={queryClient}>
          <App />
        </QueryClientProvider>
      </NuqsAdapter>
    </TopLevelErrorBoundary>
  </StrictMode>,
)
