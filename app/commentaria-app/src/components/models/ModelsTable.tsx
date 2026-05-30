import { useEffect, useMemo, useState } from 'react'
import type {
  common_OCRModelType,
  annotation_Annotation,
  model_Model,
} from '@hub-api'
import { AnnotationsService, ApiError } from '@hub-api'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { DeleteAnnotationModal } from '../modal/DeleteAnnotationModal'
import {
  useCreateModelMutation,
  useDeleteModelMutation,
  useModelsQuery,
  useTrainModelMutation,
  useUpdateModelMutation,
} from '../../queries/models'
import { SearchInput } from '../core/SearchInput'
import useLocalStorageState from 'use-local-storage-state'
import { ModelEditModal } from './ModelEditModal'
import { useAppState } from '../../context/useAppState'
import { useAuthStore } from '../../store/authStore.ts'
import { Timestamp } from '../core/Timestamp'
import { ModelImportModal } from './ModelImportModal'
import { ModelTrainModal } from './ModelTrainModal'
import { useDatasetsQuery } from '../../queries/datasets'
import { useQueries } from '@tanstack/react-query'

const getDisplayValue = (value?: string) => value?.trim() || '—'

type SortKey = 'name' | 'type' | 'algorithm' | 'base' | 'updated'
type SortDirection = 'asc' | 'desc'

type SortConfig = {
  key: SortKey
  direction: SortDirection
}

const getSortValue = (model: model_Model, key: SortKey) => {
  switch (key) {
    case 'name':
      return (model.name || model.id || '').toLowerCase()
    case 'type':
      return (model.type || '').toLowerCase()
    case 'algorithm':
      return (model.algorithm_family || '').toLowerCase()
    case 'base':
      return (model.base_model_id || '').toLowerCase()
    case 'updated': {
      const raw = model.updated_at || model.created_at
      const time = raw ? new Date(raw).getTime() : 0
      return Number.isNaN(time) ? 0 : time
    }
    default:
      return ''
  }
}

type BaseAnnotationLookup = Record<string, annotation_Annotation[]>

function AnnotationsCell({
  references,
}: {
  references: model_Model['base_annotations']
}) {
  const { setState } = useAppState()
  const [isExpanded, setIsExpanded] = useState(false)
  const { data: datasets } = useDatasetsQuery()
  const annotations = useMemo(() => references ?? [], [references])
  const baseCount = annotations.length
  const datasetNameById = useMemo(() => {
    const lookup = new Map<string, string>()
    datasets?.forEach((dataset) => {
      if (dataset.id) {
        lookup.set(dataset.id, dataset.name || dataset.id)
      }
    })
    return lookup
  }, [datasets])
  const datasetIds = useMemo(() => {
    const seen = new Set<string>()
    annotations.forEach((ref) => {
      if (ref.dataset_id) {
        seen.add(ref.dataset_id)
      }
    })
    return Array.from(seen)
  }, [annotations])

  const annotationQueries = useQueries({
    queries: datasetIds.map((datasetId) => ({
      queryKey: ['annotations', datasetId] as const,
      queryFn: () =>
        AnnotationsService.getDatasetsAnnotations({
          dataSetId: datasetId,
        }),
      enabled: isExpanded,
    })),
  })

  const annotationsByDataset = useMemo<BaseAnnotationLookup>(() => {
    return annotationQueries.reduce<BaseAnnotationLookup>(
      (acc, query, index) => {
        const datasetId = datasetIds[index]
        if (datasetId && query.data) {
          acc[datasetId] = query.data
        }
        return acc
      },
      {},
    )
  }, [annotationQueries, datasetIds])

  const isLoading =
    isExpanded && annotationQueries.some((query) => query.isLoading)
  const hasError =
    isExpanded && annotationQueries.some((query) => query.isError)

  const handleSelectAnnotation = (
    datasetId?: string,
    annotationId?: string,
  ) => {
    if (!datasetId || !annotationId) return
    setState({ datasetId, annotationId })
  }

  if (baseCount === 0) {
    return <td className="px-4 py-3 text-center text-xs text-gray-400">None</td>
  }

  return (
    <td className={`px-2 py-2 ${isExpanded ? 'align-top' : 'align-middle'}`}>
      {!isExpanded && (
        <div className="flex h-full items-center justify-center">
          <button
            type="button"
            onClick={() => setIsExpanded(true)}
            className="inline-flex items-center justify-center rounded-full border border-gray-200 px-2 py-0.5 text-xs font-semibold text-gray-600 hover:border-gray-300 hover:text-gray-800"
          >
            {baseCount}
          </button>
        </div>
      )}
      {isExpanded && (
        <div className="flex flex-col items-center gap-2">
          <button
            type="button"
            onClick={() => setIsExpanded(false)}
            className="inline-flex w-fit items-center gap-2 rounded-full border border-gray-200 px-2 py-0.5 text-[11px] font-semibold text-gray-600 hover:border-gray-300 hover:text-gray-800"
          >
            <span>{baseCount}</span>
            <span className="text-[10px] uppercase tracking-wide text-gray-400">
              Hide
            </span>
          </button>
          {isLoading && (
            <span className="text-xs text-gray-400">Loading annotations…</span>
          )}
          {hasError && (
            <span className="text-xs text-red-500">
              Unable to load annotations.
            </span>
          )}
          {!isLoading && !hasError && (
            <div className="flex w-full flex-col gap-2">
              {annotations.map((ref, index) => {
                const datasetId = ref.dataset_id
                const annotationId = ref.id
                const annotation =
                  datasetId && annotationId
                    ? annotationsByDataset[datasetId]?.find(
                        (item) => item.id === annotationId,
                      )
                    : null
                const datasetLabel = datasetId
                  ? datasetNameById.get(datasetId) || datasetId
                  : 'Dataset'
                const annotationLabel =
                  annotation?.name ||
                  annotation?.id ||
                  annotationId ||
                  'Annotation'
                return (
                  <Button
                    key={`${datasetId || 'dataset'}-${annotationId || index}`}
                    type="button"
                    onClick={() =>
                      handleSelectAnnotation(datasetId, annotationId)
                    }
                    variant="primary"
                    className="flex w-full items-center justify-start px-2 py-1 text-xs"
                  >
                    <span className="text-gray-500">{datasetLabel}</span>
                    <span className="ml-3">{annotationLabel}</span>
                  </Button>
                )
              })}
            </div>
          )}
        </div>
      )}
    </td>
  )
}

function BaseAnnotationsCell({ model }: { model: model_Model }) {
  return <AnnotationsCell references={model.base_annotations} />
}

function UsedAnnotationsCell({ model }: { model: model_Model }) {
  return <AnnotationsCell references={model.used_in_annotations} />
}

export function ModelsTable() {
  const { data: models, isLoading, error } = useModelsQuery()
  const createMutation = useCreateModelMutation()
  const deleteMutation = useDeleteModelMutation()
  const trainMutation = useTrainModelMutation()
  const updateMutation = useUpdateModelMutation()
  const [modelToDelete, setModelToDelete] = useState<model_Model | null>(null)
  const [modelToEdit, setModelToEdit] = useState<model_Model | null>(null)
  const [isImportOpen, setIsImportOpen] = useState(false)
  const [isTrainOpen, setIsTrainOpen] = useState(false)
  const [copiedId, setCopiedId] = useState<string | null>(null)
  const { modelSearchPrefill, setModelSearchPrefill } = useAppState()
  const isAuthenticated = !!useAuthStore((store) => store.token)
  const [searchQuery, setSearchQuery] = useLocalStorageState<string>(
    'modelsSearch',
    {
      defaultValue: '',
    },
  )
  const [sortConfig, setSortConfig] = useLocalStorageState<SortConfig>(
    'modelsSort',
    {
      defaultValue: {
        key: 'name',
        direction: 'asc',
      },
    },
  )

  const deleteErrorMessage =
    deleteMutation.error instanceof ApiError
      ? deleteMutation.error.body
      : deleteMutation.error
        ? String(deleteMutation.error)
        : null
  const updateErrorMessage =
    updateMutation.error instanceof ApiError
      ? updateMutation.error.body
      : updateMutation.error
        ? String(updateMutation.error)
        : null
  const createErrorMessage =
    createMutation.error instanceof ApiError
      ? createMutation.error.body
      : createMutation.error
        ? String(createMutation.error)
        : null
  const trainErrorMessage =
    trainMutation.error instanceof ApiError
      ? trainMutation.error.body
      : trainMutation.error
        ? String(trainMutation.error)
        : null

  const queryErrorMessage =
    error instanceof ApiError ? error.body : error ? String(error) : null

  const rows = useMemo(() => models ?? [], [models])
  const filteredRows = useMemo(() => {
    const query = searchQuery.trim().toLowerCase()
    if (!query) {
      return rows
    }
    return rows.filter((model) => {
      const haystack = [
        model.id,
        model.name,
        model.description,
        model.type,
        model.algorithm_family,
        model.base_model_id,
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
      return haystack.includes(query)
    })
  }, [rows, searchQuery])

  const sortedRows = useMemo(() => {
    const data = [...filteredRows]
    data.sort((a, b) => {
      const aValue = getSortValue(a, sortConfig.key)
      const bValue = getSortValue(b, sortConfig.key)
      if (aValue < bValue) return sortConfig.direction === 'asc' ? -1 : 1
      if (aValue > bValue) return sortConfig.direction === 'asc' ? 1 : -1
      return 0
    })
    return data
  }, [filteredRows, sortConfig.direction, sortConfig.key])

  const toggleSort = (key: SortKey) => {
    setSortConfig((current) => {
      if (current.key === key) {
        return {
          key,
          direction: current.direction === 'asc' ? 'desc' : 'asc',
        }
      }
      return {
        key,
        direction: 'asc',
      }
    })
  }

  const renderSortHeader = (label: string, key: SortKey) => {
    const isActive = sortConfig.key === key
    const arrow = isActive ? (sortConfig.direction === 'asc' ? '▲' : '▼') : null
    return (
      <button
        type="button"
        onClick={() => toggleSort(key)}
        className={`inline-flex items-center gap-1 ${isActive ? 'text-gray-800' : 'text-gray-500 hover:text-gray-700'}`}
      >
        <span>{label}</span>
        {arrow && <span className="text-[10px]">{arrow}</span>}
      </button>
    )
  }

  const handleEditOpen = (model: model_Model) => {
    setModelToEdit(model)
  }

  const handleEditClose = () => {
    setModelToEdit(null)
  }

  const handleEditSubmit = (updates: {
    name: string
    description?: string
    type: string
    algorithm_family?: string
    base_model_id?: string
    base_annotations?: model_Model['base_annotations']
  }) => {
    if (!modelToEdit?.id) {
      return
    }
    updateMutation.mutate(
      {
        id: modelToEdit.id,
        model: {
          name: updates.name,
          description: updates.description,
          type: updates.type as common_OCRModelType,
          algorithm_family: updates.algorithm_family as 'yolo' | undefined,
          base_model_id: updates.base_model_id,
          base_annotations:
            updates.base_annotations ?? modelToEdit.base_annotations,
          categories: modelToEdit.categories,
        },
      },
      {
        onSuccess: () => {
          setModelToEdit(null)
        },
      },
    )
  }

  const handleImportSubmit = (payload: {
    file: File
    name: string
    description?: string
    baseModelId?: string
    baseAnnotations?: string
  }) => {
    createMutation.mutate(payload, {
      onSuccess: () => {
        setIsImportOpen(false)
      },
    })
  }

  const handleDelete = () => {
    if (!modelToDelete?.id) {
      return
    }
    deleteMutation.mutate(
      { id: modelToDelete.id },
      {
        onSuccess: () => {
          setModelToDelete(null)
          deleteMutation.reset()
        },
      },
    )
  }

  const handleDeleteCancel = () => {
    deleteMutation.reset()
    setModelToDelete(null)
  }

  const handleImportClose = () => {
    setIsImportOpen(false)
    createMutation.reset()
  }

  const handleTrainClose = () => {
    setIsTrainOpen(false)
    trainMutation.reset()
  }

  const handleTrainSubmit = (model: model_Model) => {
    trainMutation.mutate(model, {
      onSuccess: () => {
        setIsTrainOpen(false)
      },
    })
  }

  const handleCopyId = (id: string | undefined) => {
    if (!id) return
    void navigator.clipboard.writeText(id).then(() => {
      setCopiedId(id)
      setTimeout(() => setCopiedId(null), 2000)
    })
  }

  useEffect(() => {
    if (!modelSearchPrefill) {
      return
    }
    setSearchQuery(modelSearchPrefill)
    setModelSearchPrefill(null)
  }, [modelSearchPrefill, setModelSearchPrefill, setSearchQuery])

  return (
    <div className="w-full h-full flex flex-col">
      <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-white gap-4">
        <div className="flex items-center gap-6">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Models</h2>
            <p className="text-xs text-gray-500">
              {filteredRows.length}
              {searchQuery && ` of ${rows.length}`}{' '}
              {rows.length === 1 ? 'model' : 'models'}
            </p>
          </div>
          <SearchInput
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Search models..."
            className="w-88 max-w-full"
          />
        </div>
        {isAuthenticated && (
          <div className="flex items-center gap-2">
            <Button
              className="px-3 py-1.5 text-sm font-semibold"
              onClick={() => setIsTrainOpen(true)}
            >
              Train a model
            </Button>
            <Button
              variant="primary"
              className="px-3 py-1.5 text-sm font-semibold"
              onClick={() => setIsImportOpen(true)}
            >
              Import a model
            </Button>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-hidden px-2 py-4">
        <div className="flex flex-col h-full">
          {isLoading && <LoadingSpinner message="Loading models..." />}
          {!isLoading && queryErrorMessage && (
            <div className="mb-4">
              <ErrorMessage message={queryErrorMessage} />
            </div>
          )}
          {!isLoading && rows.length === 0 && !queryErrorMessage && (
            <div className="text-sm text-gray-500">No models available.</div>
          )}

          {!isLoading && rows.length > 0 && filteredRows.length === 0 && (
            <div className="text-sm text-gray-500">
              No models match "{searchQuery.trim()}".
            </div>
          )}

          {!isLoading && sortedRows.length > 0 && (
            <div className="flex-1 min-h-0">
              <div className="h-full overflow-auto rounded-lg border border-gray-200 bg-white">
                <table className="min-w-full text-sm">
                  <thead className="bg-gray-50 text-xs text-gray-500">
                    <tr>
                      <th className="px-4 py-3 text-left">Base Annotations</th>
                      <th className="px-4 py-3 text-left">
                        {renderSortHeader('Model', 'name')}
                      </th>
                      <th className="px-4 py-3 text-left">
                        {renderSortHeader('Type', 'type')}
                      </th>
                      <th className="px-4 py-3 text-left">
                        {renderSortHeader('Algorithm', 'algorithm')}
                      </th>
                      <th className="px-4 py-3 text-left">
                        {renderSortHeader('Base Model', 'base')}
                      </th>
                      <th className="px-4 py-3 text-left">
                        Used in Annotations
                      </th>
                      <th className="px-4 py-3 text-left">
                        {renderSortHeader('Updated', 'updated')}
                      </th>
                      {isAuthenticated && (
                        <th className="px-4 py-3 text-right"></th>
                      )}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {sortedRows.map((model, index) => {
                      return (
                        <tr
                          key={model.id || model.name || `model-${index}`}
                          className="hover:bg-gray-50"
                        >
                          <BaseAnnotationsCell model={model} />
                          <td className="px-4 py-3">
                            <div>
                              <div className="font-medium text-gray-900">
                                {getDisplayValue(model.name || model.id)}
                              </div>
                              {model.description && (
                                <div className="text-xs text-gray-500 mt-1 max-w-xs whitespace-pre-line">
                                  {model.description.replace(/\\n/g, '\n')}
                                </div>
                              )}
                            </div>
                          </td>
                          <td className="px-4 py-3 text-gray-700">
                            {getDisplayValue(model.type)}
                          </td>
                          <td className="px-4 py-3 text-gray-700">
                            {getDisplayValue(model.algorithm_family)}
                          </td>
                          <td className="px-4 py-3 text-gray-700">
                            <span className="block max-w-[160px] truncate">
                              {getDisplayValue(model.base_model_id)}
                            </span>
                          </td>
                          <UsedAnnotationsCell model={model} />
                          <td className="px-4 py-3 text-gray-700 text-xs whitespace-nowrap">
                            <Timestamp
                              hideFullDate
                              date={model.updated_at || model.created_at}
                            />
                          </td>
                          {isAuthenticated && (
                            <td className="px-4 py-3 whitespace-nowrap">
                              <div className="flex items-center justify-end gap-2">
                                <Button
                                  className="px-2 py-1 text-xs"
                                  onClick={() => handleCopyId(model.id)}
                                >
                                  {copiedId === model.id
                                    ? 'Copied!'
                                    : 'Copy ID'}
                                </Button>
                                <Button
                                  className="px-2 py-1 text-xs"
                                  onClick={() => handleEditOpen(model)}
                                >
                                  Edit
                                </Button>
                                <Button
                                  variant="danger"
                                  className="px-2 py-1 text-xs"
                                  onClick={() => setModelToDelete(model)}
                                >
                                  Delete
                                </Button>
                              </div>
                            </td>
                          )}
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      </div>

      <DeleteAnnotationModal
        isOpen={!!modelToDelete}
        annotationLabel={
          modelToDelete?.name || modelToDelete?.id || 'this model'
        }
        title="Delete model"
        loadingMessage="Deleting model..."
        error={deleteErrorMessage}
        isDeleting={deleteMutation.isPending}
        onCancel={handleDeleteCancel}
        onConfirm={handleDelete}
      />
      {modelToEdit && (
        <ModelEditModal
          model={modelToEdit}
          allModels={rows}
          onClose={handleEditClose}
          onSubmit={handleEditSubmit}
          isSaving={updateMutation.isPending}
          errorMessage={updateErrorMessage}
        />
      )}
      {isImportOpen && (
        <ModelImportModal
          isOpen={isImportOpen}
          models={rows}
          onClose={handleImportClose}
          onSubmit={handleImportSubmit}
          isSaving={createMutation.isPending}
          errorMessage={createErrorMessage}
        />
      )}
      {isTrainOpen && (
        <ModelTrainModal
          isOpen={isTrainOpen}
          models={rows}
          onClose={handleTrainClose}
          onSubmit={handleTrainSubmit}
          isSaving={trainMutation.isPending}
          errorMessage={trainErrorMessage}
        />
      )}
    </div>
  )
}
