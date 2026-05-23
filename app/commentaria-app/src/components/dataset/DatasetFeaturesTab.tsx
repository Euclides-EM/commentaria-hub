import { useState } from 'react'
import { type feature_Revision } from '@hub-api'
import { useAppState } from '../../context/useAppState.ts'
import { useAuthStore } from '../../store/authStore.ts'
import { normalizeFeatureProperties } from '../../utils/featureProperties.ts'
import { Button } from '../core/Button.tsx'
import { SearchInput } from '../core/SearchInput.tsx'
import { CreateFeatureModal } from './CreateFeatureModal.tsx'
import { CreateRevisionModal } from './CreateRevisionModal.tsx'
import { FeatureCard } from './FeatureCard.tsx'
import { useDatasetFeatures } from './useDatasetFeatures.ts'

export function DatasetFeaturesTab() {
  const { state } = useAppState()
  const datasetId = state.datasetId
  const isAuthenticated = !!useAuthStore((store) => store.token)

  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [revisionModal, setRevisionModal] = useState<{
    featureId: string
    latestRevision?: feature_Revision
  } | null>(null)

  const {
    sortedFeatures,
    filteredFeatures,
    featureEdits,
    expandedFeatures,
    editingFeatures,
    busyFeatureId,
    error,
    availableProperties,
    isLoadingProperties,
    loading,
    searchQuery,
    setSearchQuery,
    handleUpdateFeature,
    handleDeleteFeature,
    handleCancelEdit,
    startEditing,
    toggleExpand,
    updateEditField,
  } = useDatasetFeatures(datasetId)

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col bg-white m-3 mb-0 w-[calc(100%-1.5rem)] max-w-[80vw] mx-auto">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Dataset Features</div>
        {isAuthenticated && (
          <Button
            onClick={() => setIsCreateModalOpen(true)}
            variant="primary"
            className="px-2 py-1 text-xs"
          >
            Create
          </Button>
        )}
      </div>
      <div className="flex-1 min-h-0 overflow-auto p-4">
        <div className="flex flex-col gap-6">
          <div className="flex items-center gap-4">
            <SearchInput
              placeholder="Search features"
              className="w-full max-w-[520px]"
              value={searchQuery}
              onChange={setSearchQuery}
            />
            <div className="text-xs text-gray-500 shrink-0">
              {searchQuery.trim()
                ? `Listing ${filteredFeatures.length} of ${sortedFeatures.length} features`
                : `Listing ${sortedFeatures.length} features`}
            </div>
          </div>

          {error && <div className="text-sm text-red-600">{error}</div>}

          {loading ? (
            <div className="text-sm text-gray-500">Loading features...</div>
          ) : filteredFeatures.length === 0 ? (
            <div className="text-sm text-gray-500">
              {searchQuery.trim()
                ? 'No features match your search.'
                : 'No features were created yet.'}
            </div>
          ) : (
            <div className="flex flex-col gap-5">
              {filteredFeatures.map((feature) => {
                const featureId = feature.id ?? ''
                const edits = featureId ? featureEdits[featureId] : undefined
                const isEditing = featureId ? editingFeatures[featureId] : false
                const isDirty =
                  edits &&
                  (edits.name !== (feature.name || '') ||
                    edits.description !== (feature.description || '') ||
                    edits.color !== (feature.color || '') ||
                    edits.properties.join('|') !==
                      normalizeFeatureProperties(feature.properties ?? []).join(
                        '|',
                      ))

                return (
                  <FeatureCard
                    key={featureId || feature.name}
                    feature={feature}
                    edits={edits}
                    isEditing={isEditing}
                    isExpanded={
                      featureId ? !!expandedFeatures[featureId] : false
                    }
                    isSaving={busyFeatureId === featureId}
                    isDirty={!!isDirty}
                    isAuthenticated={isAuthenticated}
                    availableProperties={availableProperties}
                    isLoadingProperties={isLoadingProperties}
                    onSubmit={(event) => handleUpdateFeature(feature, event)}
                    onEdit={() => featureId && startEditing(featureId)}
                    onCancelEdit={() => handleCancelEdit(feature)}
                    onDelete={() => handleDeleteFeature(feature)}
                    onToggleExpand={() => featureId && toggleExpand(featureId)}
                    onEditField={(update) =>
                      featureId && updateEditField(featureId, update)
                    }
                    onNewRevision={(latestRevision) =>
                      setRevisionModal({ featureId, latestRevision })
                    }
                  />
                )
              })}
            </div>
          )}
        </div>
      </div>
      <CreateFeatureModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        datasetId={datasetId}
      />
      <CreateRevisionModal
        isOpen={revisionModal !== null}
        onClose={() => setRevisionModal(null)}
        featureId={revisionModal?.featureId ?? ''}
        latestRevision={revisionModal?.latestRevision}
      />
    </section>
  )
}
