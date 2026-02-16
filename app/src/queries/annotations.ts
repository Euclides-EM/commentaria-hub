import { useQuery } from '@tanstack/react-query'
import { AnnotationsService } from '../api'

const annotationsQueryKey = (datasetId: string) =>
  ['annotations', datasetId] as const

export function useAnnotationsQuery(datasetId: string) {
  return useQuery({
    queryKey: annotationsQueryKey(datasetId),
    queryFn: async () =>
      AnnotationsService.getDatasetsAnnotations({
        dataSetId: datasetId,
      }),
    enabled: !!datasetId,
  })
}

const annotationIndexQueryKey = (datasetId: string, annotationId: string) =>
  ['annotations', datasetId, annotationId, 'index'] as const

export function useAnnotationIndexQuery(
  datasetId: string,
  annotationId: string,
) {
  return useQuery({
    queryKey: annotationIndexQueryKey(datasetId, annotationId),
    queryFn: () =>
      AnnotationsService.getDatasetsAnnotationsIndex({
        dataSetId: datasetId,
        id: annotationId,
        categories: '',
      }),
    enabled: !!datasetId && !!annotationId,
  })
}

export const annotationTeiQueryKey = (
  datasetId: string,
  annotationId: string,
  pageNum: number,
) => ['annotations', datasetId, annotationId, 'tei', pageNum] as const

export function useAnnotationTeiQuery(
  datasetId: string,
  annotationId: string,
  pageNum: number,
  enabled: boolean,
) {
  return useQuery({
    queryKey: annotationTeiQueryKey(datasetId, annotationId, pageNum),
    queryFn: () =>
      AnnotationsService.getDatasetsAnnotationsTei({
        dataSetId: datasetId,
        id: annotationId,
        pageNumOrKey: pageNum.toString(),
      }),
    enabled: !!datasetId && !!annotationId && enabled,
  })
}

export function useAnnotationSearch(
  datasetId: string,
  annotationId: string,
  regex: string,
  categories: string[] = [],
) {
  return useQuery({
    queryKey: [
      'annotations',
      datasetId,
      annotationId,
      'search',
      { regex, categories },
    ] as const,
    queryFn: () =>
      AnnotationsService.getDatasetsAnnotationsSearch({
        dataSetId: datasetId,
        id: annotationId,
        regex,
        category: categories.length > 0 ? categories : undefined,
      }),
    enabled: !!datasetId && !!annotationId && regex.length > 0,
  })
}

export function useAnnotationCategories(
  datasetId: string,
  annotationId: string,
) {
  return useQuery({
    queryKey: ['annotations', datasetId, annotationId, 'categories'] as const,
    queryFn: () =>
      AnnotationsService.getDatasetsAnnotationsCategories({
        dataSetId: datasetId,
        id: annotationId,
      }),
    enabled: !!datasetId && !!annotationId,
  })
}
