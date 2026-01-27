import { useEffect, useMemo, useState } from 'react'
import type { model_Model } from '../../api'
import { ApiError } from '../../api'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'
import { LoadingSpinner } from '../core/LoadingSpinner'
import { DeleteAnnotationModal } from '../modal/DeleteAnnotationModal'
import { useDeleteModelMutation, useModelsQuery } from '../../queries/models'
import { SearchInput } from '../core/SearchInput'
import useLocalStorageState from 'use-local-storage-state'
import { ModelEditModal } from './ModelEditModal'
import { useQueryClient } from '@tanstack/react-query'
import { useAppState } from '../../context/useAppState'
import { useAuthStore } from '../../store/authStore.ts'

const formatDate = (value?: string) => {
  if (!value) {
    return '—'
  }
  const date = new Date(value)
  return date.toLocaleString()
}

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

export function ModelsTable() {
  const queryClient = useQueryClient()
  const { data: models, isLoading, error } = useModelsQuery()
  const deleteMutation = useDeleteModelMutation()
  const [modelToDelete, setModelToDelete] = useState<model_Model | null>(null)
  const [modelToEdit, setModelToEdit] = useState<model_Model | null>(null)
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
    const arrow = isActive ? (sortConfig.direction === 'asc' ? '▲' : '▼') : ''
    return (
      <button
        type="button"
        onClick={() => toggleSort(key)}
        className={`inline-flex items-center gap-1 ${isActive ? 'text-gray-800' : 'text-gray-500 hover:text-gray-700'}`}
      >
        <span>{label}</span>
        <span className="text-[10px]">{arrow}</span>
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
  }) => {
    if (!modelToEdit?.id) {
      return
    }
    setModelToEdit(null)
    queryClient.invalidateQueries({ queryKey: ['models'] })
    void updates
  }

  const handleDelete = () => {
    if (!modelToDelete?.id) {
      return
    }
    deleteMutation.mutate({ id: modelToDelete.id })
  }

  const handleDeleteCancel = () => {
    deleteMutation.reset()
    setModelToDelete(null)
  }

  useEffect(() => {
    if (deleteMutation.isSuccess) {
      setModelToDelete(null)
      deleteMutation.reset()
    }
  }, [deleteMutation.isSuccess, deleteMutation.reset])

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
        <div>
          <h2 className="text-lg font-semibold text-gray-900">Models</h2>
          <p className="text-xs text-gray-500">
            {filteredRows.length}
            {searchQuery && ` of ${rows.length}`}{' '}
            {rows.length === 1 ? 'model' : 'models'}
          </p>
        </div>
        <div className="w-full max-w-xs">
          <SearchInput
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Search models..."
          />
        </div>
      </div>

      <div className="flex-1 overflow-auto p-6">
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
          <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
            <table className="min-w-full text-sm">
              <thead className="bg-gray-50 text-xs uppercase text-gray-500">
                <tr>
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
                    {renderSortHeader('Updated', 'updated')}
                  </th>
                  {isAuthenticated && (
                    <th className="px-4 py-3 text-right">Actions</th>
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
                      <td className="px-4 py-3 text-gray-700">
                        {formatDate(model.updated_at || model.created_at)}
                      </td>
                      {isAuthenticated && (
                        <td className="px-4 py-3">
                          <div className="flex items-center justify-end gap-2">
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
        )}
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
      <ModelEditModal
        isOpen={!!modelToEdit}
        model={modelToEdit}
        allModels={rows}
        onClose={handleEditClose}
        onSubmit={handleEditSubmit}
      />
    </div>
  )
}
