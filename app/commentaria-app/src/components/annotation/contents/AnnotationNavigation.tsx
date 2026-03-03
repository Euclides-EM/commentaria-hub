import { useEffect, useRef, useState } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { PageNavigation } from './navigation/PageNavigation.tsx'

export const AnnotationNavigation = () => {
  const [collapsed, setCollapsed] = useLocalStorageState('sidebarCollapsed', {
    defaultValue: false,
    storageSync: false,
  })
  const [sidebarWidth, setSidebarWidth] = useLocalStorageState('sidebarWidth', {
    defaultValue: 380,
    storageSync: false,
  })
  const [isResizing, setIsResizing] = useState(false)
  const asideRef = useRef<HTMLElement | null>(null)

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

  return (
    <aside
      ref={asideRef}
      className="border-r border-gray-200 flex flex-col overflow-hidden bg-white transition-all duration-200 flex-1 relative"
      style={{
        width: collapsed ? '44px' : `${sidebarWidth}px`,
        minWidth: collapsed ? '44px' : `${sidebarWidth}px`,
        maxWidth: collapsed ? '44px' : `${sidebarWidth}px`,
      }}
    >
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        {!collapsed && <div className="font-semibold">Navigation</div>}
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
