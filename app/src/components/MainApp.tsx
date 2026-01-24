import { AnnotationDetailsTab } from './annotation/AnnotationDetailsTab.tsx'
import { AnnotationContentsTab } from './annotation/AnnotationContentsTab.tsx'
import { useAppState } from '../context/useAppState'
import useLocalStorageState from 'use-local-storage-state'

type Tab = 'details' | 'text'

const TabButton = ({
  onSelected,
  title,
  isActive,
}: {
  onSelected: () => void
  title: string
  isActive: boolean
}) => {
  return (
    <button
      className={`px-3 py-1 rounded w-45 text-sm ${
        isActive
          ? 'bg-gray-500  text-white !cursor-default'
          : 'bg-gray-200 hover:bg-gray-300'
      }`}
      onClick={() => onSelected()}
    >
      {isActive && '> '}
      {title}
    </button>
  )
}

export function MainApp() {
  const [activeTab, setActiveTab] = useLocalStorageState<Tab>(
    'annotation-tab',
    { defaultValue: 'details' },
  )
  const { state } = useAppState()

  if (!state.datasetId || !state.annotationId) {
    return (
      <div className="w-full m-10 font-medium text-center">
        Please select dataset and annotation
      </div>
    )
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex w-full gap-4 p-3 border-b border-gray-200 bg-white items-center justify-center">
        <TabButton
          onSelected={() => setActiveTab('details')}
          title="Annotation Details"
          isActive={activeTab === 'details'}
        />
        <TabButton
          onSelected={() => setActiveTab('text')}
          title="Annotation Contents"
          isActive={activeTab === 'text'}
        />
      </div>
      <div className="flex-1 overflow-auto">
        {activeTab === 'details' && <AnnotationDetailsTab />}
        {activeTab === 'text' && <AnnotationContentsTab />}
      </div>
    </div>
  )
}
