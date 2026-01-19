import { useQuery } from '@tanstack/react-query'
import { AnnotationsService } from '../api'

export const annotationsQueryKey = (
  datasetId: string,
  onlyTranscribed?: boolean,
) => ['annotations', datasetId, { onlyTranscribed }] as const

export function useAnnotationsQuery(
  datasetId: string,
  onlyTranscribed: boolean = false,
) {
  return useQuery({
    queryKey: annotationsQueryKey(datasetId, onlyTranscribed),
    queryFn: async () => {
      const data = await AnnotationsService.getDatasetsAnnotations({
        dataSetId: datasetId,
      })
      return onlyTranscribed ? data.filter((ann) => ann.ocred) : data
    },
    enabled: !!datasetId,
  })
}

export const annotationIndexQueryKey = (
  datasetId: string,
  annotationId: string,
) => ['annotations', datasetId, annotationId, 'index'] as const

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
) {
  return useQuery({
    queryKey: annotationTeiQueryKey(datasetId, annotationId, pageNum),
    queryFn: () =>
      AnnotationsService.getDatasetsAnnotationsTei({
        dataSetId: datasetId,
        id: annotationId,
        pageNum: pageNum.toString(),
      }),
    enabled: !!datasetId && !!annotationId,
  })
}
