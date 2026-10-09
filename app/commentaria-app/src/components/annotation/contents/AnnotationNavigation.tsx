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
        className="absolute top-2 left-2 z-40 px-2.5 py-1.5 text-sm font-semibold cursor-pointer bg-gray-50 border border-gray-200 rounded-md shadow-md hover:bg-gray-100"
        onClick={() => setCollapsed(false)}
        title="Expand index"
        aria-label="Expand index"
      >
        ⟩
      </button>
    )
  }

  return (
    <aside
      ref={asideRef}
      onPointerDownCapture={() => {
        pointerInsideRef.current = true
      }}
      className={`border-r border-gray-200 flex flex-col overflow-hidden bg-white ${floating ? 'absolute top-0 left-0 bottom-0 z-40 shadow-xl' : 'transition-all duration-200 flex-1 relative'}`}
      style={{
        width: collapsed ? '44px' : `${sidebarWidth}px`,
        minWidth: collapsed ? '44px' : `${sidebarWidth}px`,
        maxWidth: collapsed ? '44px' : `${sidebarWidth}px`,
      }}
    >
      <div
        className={`px-2.5 py-2 text-sm font-semibold bg-gray-50 flex items-center gap-2.5 ${collapsed ? 'flex-1 justify-center' : 'justify-between border-b border-gray-200'}`}
      >
        {!collapsed && (
          <div className="flex items-center gap-3">
            <div className="font-semibold">Navigation</div>
            <label
              className="flex items-center gap-1.5 cursor-pointer text-xs font-medium"
              title={
                floating
                  ? 'Pin navigation beside contents'
                  : 'Float navigation over contents'
              }
            >
              <button
                role="switch"
                aria-checked={floating}
                onClick={() => handleFloatingChange(!floating)}
                className={`relative inline-flex h-4 w-7 shrink-0 cursor-pointer rounded-full transition-colors ${floating ? 'bg-teal-600' : 'bg-gray-300'}`}
              >
                <span
                  className={`absolute top-0.5 left-0.5 h-3 w-3 rounded-full bg-white shadow transition-transform ${floating ? 'translate-x-3' : ''}`}
                />
              </button>
              <span>Floating</span>
            </label>
          </div>
        )}
        <button
          className={`px-2.5 py-1.5 cursor-pointer ${collapsed ? 'rotate-180' : ''} transition-transform`}
          onClick={() => setCollapsed(!collapsed)}
          title={collapsed ? 'Expand index' : 'Minimize index'}
          aria-label={collapsed ? 'Expand index' : 'Minimize index'}
        >
          ⟨
        </button>
      </div>
      {!collapsed && (
        <>
          <PageNavigation />
          <div
            role="separator"
            aria-label="Resize navigation"
            className="absolute top-0 right-0 h-full w-2 cursor-col-resize flex items-center justify-center bg-gray-100 hover:bg-gray-200 transition-colors"
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
