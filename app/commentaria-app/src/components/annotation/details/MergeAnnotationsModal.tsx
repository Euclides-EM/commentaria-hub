import { useMemo, useState } from 'react'
import { useQueries } from '@tanstack/react-query'
import {
  type annotation_Annotation,
  type annotation_Reference,
  AnnotationsService,
} from '@hub-api'
import { useDatasetsQuery } from '../../../queries/datasets'
import { Button } from '../../core/Button'
import { ErrorMessage } from '../../core/ErrorMessage'
import { LoadingSpinner } from '../../core/LoadingSpinner'
import { SearchInput } from '../../core/SearchInput'

type AnnotationOption = {
  annotation: annotation_Annotation
  datasetId: string
  datasetName: string
}

interface MergeAnnotationsModalProps {
  isOpen: boolean
  currentAnnotation: annotation_Annotation
  isMerging: boolean
  error?: string | null
  onClose: () => void
  onConfirm: (annotationsToMerge: annotation_Reference[]) => void
}

export function MergeAnnotationsModal({
  isOpen,
  currentAnnotation,
  isMerging,
  error,
  onClose,
  onConfirm,
}: MergeAnnotationsModalProps) {
  if (!isOpen) {
    return null
  }

  return (
    <MergeAnnotationsModalContent
      key={`${currentAnnotation.dataset_id}:${currentAnnotation.id}`}
      currentAnnotation={currentAnnotation}
      isMerging={isMerging}
      error={error}
      onClose={onClose}
      onConfirm={onConfirm}
    />
  )
}

interface MergeAnnotationsModalContentProps {
  currentAnnotation: annotation_Annotation
  isMerging: boolean
  error?: string | null
  onClose: () => void
  onConfirm: (annotationsToMerge: annotation_Reference[]) => void
}

function MergeAnnotationsModalContent({
  currentAnnotation,
  isMerging,
  error,
  onClose,
  onConfirm,
}: MergeAnnotationsModalContentProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedKeys, setSelectedKeys] = useState<string[]>([])
  const [selectionError, setSelectionError] = useState<string | null>(null)
  const { data: datasets = [] } = useDatasetsQuery()

  const datasetIds = useMemo(
    () => datasets.flatMap((dataset) => (dataset.id ? [dataset.id] : [])),
    [datasets],
  )

  const annotationQueries = useQueries({
    queries: datasetIds.map((datasetId) => ({
      queryKey: ['annotations', datasetId] as const,
      queryFn: () =>
        AnnotationsService.getDatasetsAnnotations({
          dataSetId: datasetId,
        }),
      enabled: datasetIds.length > 0,
    })),
  })

  const options = useMemo<AnnotationOption[]>(() => {
    const datasetNameById = new Map<string, string>()
    datasets.forEach((dataset) => {
      if (dataset.id) {
        datasetNameById.set(dataset.id, dataset.name || dataset.id)
      }
    })

    const allOptions: AnnotationOption[] = []
    annotationQueries.forEach((query, index) => {
      const datasetId = datasetIds[index]
      if (!datasetId || !query.data) {
        return
      }

      query.data.forEach((annotation) => {
        if (
          !annotation.id ||
          (annotation.id === currentAnnotation.id &&
            annotation.dataset_id === currentAnnotation.dataset_id)
        ) {
          return
        }

        allOptions.push({
          annotation,
          datasetId,
          datasetName: datasetNameById.get(datasetId) || datasetId,
        })
      })
    })

    return allOptions.sort((a, b) => {
      const datasetCompare = a.datasetName.localeCompare(b.datasetName)
      if (datasetCompare !== 0) {
        return datasetCompare
      }

      return (a.annotation.name || a.annotation.id || '').localeCompare(
        b.annotation.name || b.annotation.id || '',
      )
    })
  }, [
    annotationQueries,
    currentAnnotation.dataset_id,
    currentAnnotation.id,
    datasetIds,
    datasets,
  ])

  const filteredOptions = useMemo(() => {
    const trimmedQuery = searchQuery.trim().toLowerCase()
    if (!trimmedQuery) {
      return options
    }

    return options.filter(({ annotation, datasetId, datasetName }) =>
      [
        annotation.name,
        annotation.description,
        annotation.id,
        datasetId,
        datasetName,
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(trimmedQuery),
    )
  }, [options, searchQuery])

  const selectedAnnotations = useMemo(
    () =>
      options.filter((option) =>
        selectedKeys.includes(`${option.datasetId}:${option.annotation.id}`),
      ),
    [options, selectedKeys],
  )

  const hasCrossDatasetSelection = selectedAnnotations.some(
    (option) => option.datasetId !== currentAnnotation.dataset_id,
  )

  const isLoading = annotationQueries.some((query) => query.isLoading)
  const queryError = annotationQueries.find((query) => query.error)?.error

  const toggleSelection = (option: AnnotationOption) => {
    const key = `${option.datasetId}:${option.annotation.id}`
    setSelectionError(null)
    setSelectedKeys((currentKeys) =>
      currentKeys.includes(key)
        ? currentKeys.filter((currentKey) => currentKey !== key)
        : [...currentKeys, key],
    )
  }

  const handleConfirm = () => {
    if (selectedAnnotations.length === 0) {
      setSelectionError('Select at least one annotation to merge.')
      return
    }

    onConfirm(
      selectedAnnotations.map((option) => ({
        dataset_id: option.datasetId,
        id: option.annotation.id,
      })),
    )
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={isMerging ? undefined : onClose}
    >
      <div
        className="bg-white rounded-lg w-full max-w-3xl max-h-[85vh] flex flex-col m-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Merge into</h2>
          <p className="mt-1 text-sm text-gray-600">
            Select one or more annotations to merge into{' '}
            <span className="font-semibold">
              {currentAnnotation.name || currentAnnotation.id}
            </span>
            .
          </p>
        </div>

        <div className="px-5 py-4 border-b border-gray-200 space-y-3">
          <SearchInput
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="Search all annotations across datasets..."
          />
          {hasCrossDatasetSelection && (
            <div className="rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-900">
              Warning: the current selection includes annotations from other
              datasets.
            </div>
          )}
          <ErrorMessage message={selectionError} />
          <ErrorMessage error={queryError} />
          <ErrorMessage message={error} />
        </div>

        <div className="flex-1 overflow-auto px-5 py-4">
          {isLoading ? (
            <LoadingSpinner size="sm" message="Loading annotations..." />
          ) : filteredOptions.length === 0 ? (
            <div className="py-8 text-center text-sm text-gray-500">
              {options.length === 0
                ? 'No annotations available for merging.'
                : 'No annotations match the current search.'}
            </div>
          ) : (
            <div className="space-y-2">
              {filteredOptions.map((option) => {
                const annotationId = option.annotation.id || ''
                const key = `${option.datasetId}:${annotationId}`
                const isSelected = selectedKeys.includes(key)

                return (
                  <label
                    key={key}
                    className="flex items-start gap-3 rounded-md border border-gray-200 px-3 py-3 hover:bg-gray-50"
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      disabled={isMerging}
                      onChange={() => toggleSelection(option)}
                      className="mt-0.5 h-4 w-4"
                    />
                    <div className="min-w-0 flex-1">
                      <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                        <span className="text-sm font-semibold text-gray-900">
                          {option.annotation.name || annotationId}
                        </span>
                        <span className="rounded bg-gray-100 px-1.5 py-0.5 text-[11px] font-medium text-gray-700">
                          {option.datasetName}
                        </span>
                        {option.datasetId !== currentAnnotation.dataset_id && (
                          <span className="rounded bg-amber-100 px-1.5 py-0.5 text-[11px] font-medium text-amber-800">
                            Other dataset
                          </span>
                        )}
                      </div>
                      <div className="mt-1 break-all font-mono text-xs text-gray-500">
                        {annotationId}
                      </div>
                      {option.annotation.description && (
                        <div className="mt-1 whitespace-pre-wrap break-words text-sm text-gray-600">
                          {option.annotation.description}
                        </div>
                      )}
                    </div>
                  </label>
                )
              })}
            </div>
          )}
        </div>

        <div className="px-5 py-4 border-t border-gray-200 flex items-center justify-between gap-3">
          <div className="text-sm text-gray-600">
            {selectedAnnotations.length} selected
          </div>
          <div className="flex justify-end gap-2">
            <Button
              onClick={onClose}
              className="px-3 py-1.5 text-sm font-semibold"
              disabled={isMerging}
            >
              Cancel
            </Button>
            <Button
              onClick={handleConfirm}
              variant="primary"
              className="px-3 py-1.5 text-sm font-semibold"
              disabled={isMerging || !!queryError || isLoading}
            >
              {isMerging ? 'Merging...' : 'Merge selected'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
