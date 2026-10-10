import { useEffect, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { useScopedStorageKey } from '../../../context/storageScope.ts'
import { PageNavigation } from './navigation/PageNavigation.tsx'

interface AnnotationNavigationProps {
  floating: boolean
  onFloatingChange: (floating: boolean) => void
  floatingOpen: boolean
  onFloatingOpenChange: (open: boolean) => void
}

export const AnnotationNavigation = ({
  floating,
  onFloatingChange,
  floatingOpen,
  onFloatingOpenChange,
}: AnnotationNavigationProps) => {
  const [frozenCollapsed, setFrozenCollapsed] = useLocalStorageState(
    useScopedStorageKey('sidebarCollapsed'),
    {
      defaultValue: false,
      storageSync: false,
    },
  )
  const [sidebarWidth, setSidebarWidth] = useLocalStorageState(
    useScopedStorageKey('sidebarWidth'),
    {
      defaultValue: 380,
      storageSync: false,
    },
  )
  const [isResizing, setIsResizing] = useState(false)
  const asideRef = useRef<HTMLElement | null>(null)
  const pointerInsideRef = useRef(false)

  useEffect(() => {
    const clampWidth = (value: number) => {
      const maxWidth = Math.round(window.innerWidth * 0.6)
      return Math.min(maxWidth, Math.max(240, value))
    }

    if (isResizing) {
      const onPointerMove = (event: PointerEvent) => {
        const container = asideRef.current
        if (!container) return
        const rect = container.getBoundingClientRect()
        const raw = event.clientX - rect.left
        setSidebarWidth(clampWidth(raw))
      }

      const onPointerUp = () => {
        setIsResizing(false)
      }

      window.addEventListener('pointermove', onPointerMove)
      window.addEventListener('pointerup', onPointerUp)
      return () => {
        window.removeEventListener('pointermove', onPointerMove)
        window.removeEventListener('pointerup', onPointerUp)
      }
    }

    const onResize = () => {
      setSidebarWidth((current) => clampWidth(current))
    }

    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
    }
  }, [isResizing, setSidebarWidth])

  const collapsed = floating ? !floatingOpen : frozenCollapsed
  const setCollapsed = floating
    ? (value: boolean) => onFloatingOpenChange(!value)
    : setFrozenCollapsed

  useEffect(() => {
    if (!floating || collapsed) return
    const onPointerDown = () => {
      if (!pointerInsideRef.current) {
        onFloatingOpenChange(false)
      }
      pointerInsideRef.current = false
    }
    window.addEventListener('pointerdown', onPointerDown)
    return () => {
      window.removeEventListener('pointerdown', onPointerDown)
    }
  }, [floating, collapsed, onFloatingOpenChange])

  const handleFloatingChange = (value: boolean) => {
    if (value) {
      onFloatingOpenChange(true)
    } else {
      setFrozenCollapsed(false)
    }
    onFloatingChange(value)
  }

  if (floating && collapsed) {
    return (
      <button
        className="absolute top-2 left-2 z-40 p-1.5 cursor-pointer text-white bg-gray-700 border border-gray-800 rounded-md shadow-[0_4px_14px_rgba(15,23,42,0.45)] hover:bg-gray-800"
        onClick={() => setCollapsed(false)}
        title="Expand navigation"
        aria-label="Expand navigation"
      >
        <svg
          className="w-5 h-5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M16 5H3M16 12H3M16 19H3M21 5h.01M21 12h.01M21 19h.01"
          />
        </svg>
      </button>
    )
  }

  return (
    <aside
      ref={asideRef}
      onPointerDownCapture={() => {
        pointerInsideRef.current = true
      }}
      className={`border-r flex flex-col overflow-hidden ${floating ? 'absolute top-0 left-0 bottom-0 z-40 border-gray-400/70 bg-white/55 backdrop-blur-xl backdrop-saturate-150 shadow-[0_10px_40px_rgba(15,23,42,0.32),0_0_0_1px_rgba(15,23,42,0.08),inset_1px_1px_0_rgba(255,255,255,0.7)]' : `border-gray-300 bg-white transition-all duration-200 flex-1 relative ${collapsed ? 'rounded-r-lg' : ''}`}`}
      style={{
        width: collapsed ? '20px' : `${sidebarWidth}px`,
        minWidth: collapsed ? '20px' : `${sidebarWidth}px`,
        maxWidth: collapsed ? '20px' : `${sidebarWidth}px`,
      }}
    >
      <div
        className={`${collapsed ? 'px-0' : 'px-2.5'} py-2 text-sm font-semibold ${floating ? 'bg-white/30' : collapsed ? 'bg-gray-200 hover:bg-gray-300 cursor-pointer' : 'bg-gray-50'} flex items-center gap-2.5 ${collapsed ? 'flex-1 justify-center' : 'justify-between border-b border-gray-300'}`}
        onClick={collapsed ? () => setCollapsed(false) : undefined}
        title={collapsed ? 'Expand navigation' : undefined}
      >
        {!collapsed && <div className="font-semibold">Navigation</div>}
        <div className="flex items-center gap-1">
          {!collapsed && (
            <button
              onClick={() => handleFloatingChange(!floating)}
              className="p-1 cursor-pointer rounded-md text-gray-700 hover:text-gray-900 hover:bg-gray-200/70"
              title={
                floating
                  ? 'Pin navigation: keep it docked beside the contents'
                  : 'Unpin navigation: float it over the contents and hide it when you click outside'
              }
              aria-label={
                floating
                  ? 'Pin navigation: keep it docked beside the contents'
                  : 'Unpin navigation: float it over the contents and hide it when you click outside'
              }
            >
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                {floating ? (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 17v5M9 10.76a2 2 0 0 1-1.11 1.79l-1.78.9A2 2 0 0 0 5 15.24V16a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-.76a2 2 0 0 0-1.11-1.79l-1.78-.9A2 2 0 0 1 15 10.76V7a1 1 0 0 1 1-1 2 2 0 0 0 0-4H8a2 2 0 0 0 0 4 1 1 0 0 1 1 1z"
                  />
                ) : (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 17v5M15 9.34V7a1 1 0 0 1 1-1 2 2 0 0 0 0-4H7.89M2 2l20 20M9 9v1.76a2 2 0 0 1-1.11 1.79l-1.78.9A2 2 0 0 0 5 15.24V16a1 1 0 0 0 1 1h11"
                  />
                )}
              </svg>
            </button>
          )}
          <button
            className={`py-1.5 cursor-pointer ${collapsed ? 'px-0 rotate-180' : 'px-2.5'} transition-transform`}
            onClick={() => setCollapsed(!collapsed)}
            title={collapsed ? 'Expand navigation' : 'Minimize navigation'}
            aria-label={collapsed ? 'Expand navigation' : 'Minimize navigation'}
          >
            ⟨
          </button>
        </div>
      </div>
      {!collapsed && (
        <>
          <PageNavigation />
          <div
            role="separator"
            aria-label="Resize navigation"
            className={`absolute top-0 right-0 h-full w-2 cursor-col-resize flex items-center justify-center transition-colors ${floating ? 'bg-white/20 hover:bg-white/40' : 'bg-gray-100 hover:bg-gray-200'}`}
            onPointerDown={(event) => {
              event.preventDefault()
              setIsResizing(true)
            }}
          >
            <div className="h-10 w-0.5 rounded-full bg-gray-300" />
          </div>
        </>
      )}
    </aside>
  )
}
