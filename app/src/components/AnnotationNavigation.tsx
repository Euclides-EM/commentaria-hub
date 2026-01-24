import useLocalStorageState from 'use-local-storage-state'
import { PageNavigation } from './PageNavigation.tsx'

export const AnnotationNavigation = () => {
  const [collapsed, setCollapsed] = useLocalStorageState('sidebarCollapsed', {
    defaultValue: false,
  })

  return (
    <aside
      className={`${collapsed ? 'w-11 min-w-11 max-w-11' : 'w-80 min-w-80 max-w-80'} border-r border-gray-200 flex flex-col overflow-hidden bg-white transition-all duration-200 flex-1`}
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
        </>
      )}
    </aside>
  )
}
