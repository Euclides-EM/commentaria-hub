import type { ReactNode } from 'react'
import { Component } from 'react'
import { ErrorFallback } from './ErrorFallback'

interface TopLevelErrorBoundaryProps {
  children: ReactNode
}

interface TopLevelErrorBoundaryState {
  error: Error | null
  resetKey: number
}

export class TopLevelErrorBoundary extends Component<
  TopLevelErrorBoundaryProps,
  TopLevelErrorBoundaryState
> {
  state: TopLevelErrorBoundaryState = {
    error: null,
    resetKey: 0,
  }

  static getDerivedStateFromError(error: Error): TopLevelErrorBoundaryState {
    return { error, resetKey: 0 }
  }

  componentDidCatch(error: Error) {
    console.error('Top-level error boundary caught an error:', error)
  }

  private handleRetry = () => {
    this.setState((prevState) => ({
      error: null,
      resetKey: prevState.resetKey + 1,
    }))
  }

  render() {
    const { error, resetKey } = this.state

    if (error) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 p-6">
          <div className="max-w-lg w-full rounded-xl border border-gray-200 bg-white shadow-sm">
            <ErrorFallback error={error} onRetry={this.handleRetry} />
          </div>
        </div>
      )
    }

    return <div key={resetKey}>{this.props.children}</div>
  }
}
