import { useEffect, useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type feature_Feature, FeaturesService } from '@hub-api'
import type { FeatureEditState } from './FeatureCard.tsx'

export function useDatasetFeatures(datasetId: string) {
  const queryClient = useQueryClient()
  const featuresQueryKey = ['features', 'definitions', datasetId]

  const [featureEdits, setFeatureEdits] = useState<
    Record<string, FeatureEditState>
  >({})
  const [expandedFeatures, setExpandedFeatures] = useState<
    Record<string, boolean>
  >({})
  const [actionError, setActionError] = useState<string | null>(null)
  const [busyFeatureId, setBusyFeatureId] = useState<string | null>(null)
  const [editingFeatures, setEditingFeatures] = useState<
    Record<string, boolean>
  >({})
  const [searchQuery, setSearchQuery] = useState('')

  const featuresQuery = useQuery({
    queryKey: featuresQueryKey,
    queryFn: () =>
      FeaturesService.getDatasetsFeatures({
        dataSetId: datasetId,
        expand: ['revisions'],
      }),
    refetchOnWindowFocus: false,
  })

  const sortedFeatures = useMemo(
    () =>
      [...(featuresQuery.data ?? [])].sort((a, b) =>
        (a.name || '').localeCompare(b.name || '', undefined, {
          sensitivity: 'base',
        }),
      ),
    [featuresQuery.data],
  )

  const filteredFeatures = useMemo(() => {
    const query = searchQuery.trim().toLowerCase()
    if (!query) return sortedFeatures
    return sortedFeatures.filter((feature) => {
      const name = feature.name?.toLowerCase() ?? ''
      const description = feature.description?.toLowerCase() ?? ''
      return name.includes(query) || description.includes(query)
    })
  }, [searchQuery, sortedFeatures])

  useEffect(() => {
    setFeatureEdits((previous) => {
      const next: Record<string, FeatureEditState> = {}
      for (const feature of featuresQuery.data ?? []) {
        if (!feature.id) continue
        if (editingFeatures[feature.id] && previous[feature.id]) {
          next[feature.id] = previous[feature.id]
          continue
        }
        next[feature.id] = {
          name: feature.name || '',
          description: feature.description || '',
          color: feature.color || '',
        }
      }
      return next
    })
  }, [featuresQuery.data, editingFeatures])

  const updateFeatureMutation = useMutation({
    mutationFn: ({
      featureId,
      feature,
    }: {
      featureId: string
      feature: feature_Feature
    }) =>
      FeaturesService.putDatasetsFeatures({
        dataSetId: datasetId,
        featureId,
        feature,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: featuresQueryKey })
    },
  })

  const deleteFeatureMutation = useMutation({
    mutationFn: (featureId: string) =>
      FeaturesService.deleteDatasetsFeatures({
        dataSetId: datasetId,
        featureId,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: featuresQueryKey })
    },
  })

  const queryError =
    featuresQuery.error instanceof Error
      ? featuresQuery.error.message
      : featuresQuery.error
        ? 'Failed to load features.'
        : null
  const error = actionError ?? queryError
  const loading = featuresQuery.isLoading

  const handleUpdateFeature = async (
    feature: feature_Feature,
    event?: React.FormEvent<HTMLFormElement>,
  ) => {
    event?.preventDefault()
    if (!feature.id) return
    const form = featureEdits[feature.id]
    if (!form?.name?.trim()) {
      setActionError('Feature name is required.')
      return
    }
    setBusyFeatureId(feature.id)
    setActionError(null)
    try {
      await updateFeatureMutation.mutateAsync({
        featureId: feature.id,
        feature: {
          name: form.name.trim(),
          description: form.description.trim() || undefined,
          is_root: feature.is_root ?? false,
          is_default: feature.is_default ?? false,
        },
      })
      setEditingFeatures((prev) => ({
        ...prev,
        [feature.id as string]: false,
      }))
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : 'Failed to update feature.',
      )
    } finally {
      setBusyFeatureId(null)
    }
  }

  const handleDeleteFeature = async (feature: feature_Feature) => {
    if (!feature.id) return
    if (!window.confirm(`Delete "${feature.name || 'this feature'}"?`)) return
    setBusyFeatureId(feature.id)
    setActionError(null)
    try {
      await deleteFeatureMutation.mutateAsync(feature.id)
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : 'Failed to delete feature.',
      )
    } finally {
      setBusyFeatureId(null)
    }
  }

  const handleCancelEdit = (feature: feature_Feature) => {
    if (!feature.id) return
    setFeatureEdits((prev) => ({
      ...prev,
      [feature.id as string]: {
        name: feature.name || '',
        description: feature.description || '',
        color: feature.color || '',
      },
    }))
    setEditingFeatures((prev) => ({
      ...prev,
      [feature.id as string]: false,
    }))
  }

  const startEditing = (featureId: string) => {
    setEditingFeatures((prev) => ({ ...prev, [featureId]: true }))
  }

  const toggleExpand = (featureId: string) => {
    setExpandedFeatures((prev) => ({
      ...prev,
      [featureId]: !prev[featureId],
    }))
  }

  const updateEditField = (
    featureId: string,
    update: Partial<FeatureEditState>,
  ) => {
    setFeatureEdits((prev) => ({
      ...prev,
      [featureId]: { ...prev[featureId], ...update },
    }))
  }

  return {
    sortedFeatures,
    filteredFeatures,
    featureEdits,
    expandedFeatures,
    editingFeatures,
    busyFeatureId,
    error,
    loading,
    searchQuery,
    setSearchQuery,
    handleUpdateFeature,
    handleDeleteFeature,
    handleCancelEdit,
    startEditing,
    toggleExpand,
    updateEditField,
  }
}
