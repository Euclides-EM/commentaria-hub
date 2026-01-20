import { useQuery } from '@tanstack/react-query'
import { AnnotationsService } from '../api'

type AnnotationFilter = 'transcribed' | 'ground_truth'

export const annotationsQueryKey = (
  datasetId: string,
  filters?: AnnotationFilter[],
) => ['annotations', datasetId, { filters }] as const

export function useAnnotationsQuery(
  datasetId: string,
  filters: AnnotationFilter[] = [],
) {
  return useQuery({
    queryKey: annotationsQueryKey(datasetId, filters),
    queryFn: async () => {
      const data = await AnnotationsService.getDatasetsAnnotations({
        dataSetId: datasetId,
      })

      if (filters.length === 0) {
        return data
      }

      return data.filter((ann) => {
        if (
          filters.includes('transcribed') &&
          filters.includes('ground_truth')
        ) {
          return ann.ocred && ann.ground_truth
        }
        if (filters.includes('transcribed')) {
          return ann.ocred
        }
        if (filters.includes('ground_truth')) {
          return ann.ground_truth
        }
        return true
      })
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
