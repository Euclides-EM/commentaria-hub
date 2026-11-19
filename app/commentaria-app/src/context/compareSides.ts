export type CompareSide = 'left' | 'right'

export const COMPARE_SIDE_KEYS = {
  left: {
    datasetId: 'leftDatasetId',
    annotationId: 'leftAnnotationId',
    page: 'leftPage',
  },
  right: {
    datasetId: 'rightDatasetId',
    annotationId: 'rightAnnotationId',
    page: 'rightPage',
  },
} as const

export type ComparePane = 'scan' | 'contents'

export const COMPARE_PANES: ComparePane[] = ['scan', 'contents']
