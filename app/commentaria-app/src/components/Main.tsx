import { useEffect, useState } from 'react'
import { AnnotationDetailsTab } from './annotation/details/AnnotationDetailsTab.tsx'
import { AnnotationContentsTab } from './annotation/contents/AnnotationContentsTab.tsx'
import { GalleryViewTab } from './annotation/gallery/GalleryViewTab.tsx'
import { FeatureResultsTab } from './annotation/featureResults/FeatureResultsTab.tsx'
import { FeatureExecutionsTab } from './annotation/featureExecutions/FeatureExecutionsTab.tsx'
import { useAppState } from '../context/useAppState'
import { useQuery } from '@tanstack/react-query'
import { FeaturesService } from '@hub-api'
import { ModelsTable } from './models/ModelsTable.tsx'
import { JobsTable } from './jobs/JobsTable.tsx'
import { BackupsView } from './backups/BackupsView.tsx'
import { LogsView } from './logging/LogsView.tsx'
import { Button } from './core/Button.tsx'
import { CreateDatasetModal } from './dataset/CreateDatasetModal.tsx'
import { DatasetDetails } from './dataset/DatasetDetails.tsx'
import { useAuthStore } from '../store/authStore.ts'
import { TabButton } from './core/TabButton.tsx'
import { AnnotationsTable } from './annotations/AnnotationsTable.tsx'

const diagramUrl = new URL('../assets/diagram.png', import.meta.url).href

export function Main() {
  const [isCreateDatasetOpen, setIsCreateDatasetOpen] = useState(false)
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const { state, setState } = useAppState()
  const activeTab = state.annotationTab
  const featuresQuery = useQuery({
    queryKey: ['features', 'definitions', state.datasetId],
    queryFn: () =>
      FeaturesService.getDatasetsFeatures({
        dataSetId: state.datasetId!,
      }),
    enabled: Boolean(state.datasetId && state.annotationId),
    refetchOnWindowFocus: false,
  })
  const hasDatasetFeatures = (featuresQuery.data?.length ?? 0) > 0
  const showFeatureExecutionsTab = featuresQuery.isLoading || hasDatasetFeatures

  useEffect(() => {
    if (
      (activeTab === 'featureExecutions' || activeTab === 'featureResults') &&
      !featuresQuery.isLoading &&
      !hasDatasetFeatures
    ) {
      setState({ annotationTab: 'details' })
    }
  }, [activeTab, featuresQuery.isLoading, hasDatasetFeatures, setState])

  if (state.viewMode === 'models') {
    return <ModelsTable />
  }
  if (state.viewMode === 'annotations') {
    return <AnnotationsTable />
  }
  if (state.viewMode === 'jobs') {
    return <JobsTable />
  }
  if (state.viewMode === 'backups') {
    return <BackupsView />
  }
  if (state.viewMode === 'logs') {
    return <LogsView />
  }

  if (!state.datasetId) {
    return (
      <>
        <div className="w-full m-10 font-medium text-center">
          <p>Please select a dataset to get started.</p>
          {isAuthenticated && (
            <div className="mt-4 flex flex-wrap gap-3 justify-center">
              <Button
                onClick={() => setIsCreateDatasetOpen(true)}
                variant="primary"
                className="px-3 py-2 text-sm w-40"
              >
                Create dataset
              </Button>
            </div>
          )}
          <img
            src={diagramUrl}
            alt="Diagram"
            className="mx-auto h-64 w-auto mt-8"
          />
        </div>
        <CreateDatasetModal
          isOpen={isCreateDatasetOpen}
          onClose={() => setIsCreateDatasetOpen(false)}
        />
      </>
    )
  }
  if (!state.annotationId) {
    return <DatasetDetails />
  }

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="flex w-full gap-4 p-3 border-b border-gray-200 bg-white items-center justify-center">
        <TabButton
          onSelected={() => setState({ annotationTab: 'details' })}
          title="Annotation Details"
          isActive={activeTab === 'details'}
        />
        <TabButton
          onSelected={() => setState({ annotationTab: 'gallery' })}
          title="Gallery View"
          isActive={activeTab === 'gallery'}
        />
        <TabButton
          onSelected={() => setState({ annotationTab: 'text' })}
          title="Page View"
          isActive={activeTab === 'text'}
        />
        {showFeatureExecutionsTab && (
          <TabButton
            onSelected={() => setState({ annotationTab: 'featureResults' })}
            title="Feature Results"
            isActive={activeTab === 'featureResults'}
          />
        )}
        {showFeatureExecutionsTab && (
          <TabButton
            onSelected={() => setState({ annotationTab: 'featureExecutions' })}
            title="Feature Executions"
            isActive={activeTab === 'featureExecutions'}
          />
        )}
      </div>
      <div
        className={`flex-1 min-h-0 m-3 ${activeTab === 'featureResults' || activeTab === 'gallery' ? 'flex overflow-hidden' : 'overflow-auto'}`}
      >
        {activeTab === 'details' && <AnnotationDetailsTab />}
        {activeTab === 'text' && <AnnotationContentsTab />}
        {activeTab === 'gallery' && <GalleryViewTab />}
        {activeTab === 'featureResults' && showFeatureExecutionsTab && (
          <FeatureResultsTab />
        )}
        {activeTab === 'featureExecutions' && showFeatureExecutionsTab && (
          <FeatureExecutionsTab />
        )}
      </div>
    </div>
  )
}
