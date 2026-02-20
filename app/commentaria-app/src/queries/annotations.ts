import { useQuery } from '@tanstack/react-query'
import { AnnotationsService } from '@hub-api'

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
  pageOrKey: number | string,
) => ['annotations', datasetId, annotationId, 'tei', pageOrKey] as const

export function useAnnotationTeiQuery(
  datasetId: string,
  annotationId: string,
  pageOrKey: number | string,
  enabled: boolean,
) {
  return useQuery({
    queryKey: annotationTeiQueryKey(datasetId, annotationId, pageOrKey),
    queryFn: () =>
      AnnotationsService.getDatasetsAnnotationsTei({
        dataSetId: datasetId,
        id: annotationId,
        page: String(pageOrKey),
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

const annotationImageKeysQueryKey = (datasetId: string) =>
  ['annotations', datasetId, 'image-keys'] as const

const parseImageKeys = (raw: string): string[] => {
  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    parsed = undefined
  }

  if (Array.isArray(parsed)) {
    return [...new Set(parsed.filter((item) => typeof item === 'string'))]
  }

  const hrefMatches = [...raw.matchAll(/href=(?:"([^"]+)"|'([^']+)')/g)]
  const fromHref = hrefMatches
    .map((match) => match[1] || match[2] || '')
    .map((href) => {
      const withoutQuery = href.split('?')[0].split('#')[0]
      const normalized = withoutQuery.endsWith('/')
        ? withoutQuery.slice(0, -1)
        : withoutQuery
      const lastSegment = normalized.split('/').pop()
      return decodeURIComponent(lastSegment || '')
    })
    .filter((item) => item && item !== '.' && item !== '..')

  if (fromHref.length > 0) {
    return [...new Set(fromHref)].sort((a, b) =>
      a.localeCompare(b, undefined, { numeric: true }),
    )
  }

  return [
    ...new Set(
      raw
        .split('\n')
        .map((line) => line.trim())
        .filter((line) => line.length > 0),
    ),
  ].sort((a, b) => a.localeCompare(b, undefined, { numeric: true }))
}

export function useAnnotationImageKeysQuery(datasetId: string) {
  return useQuery({
    queryKey: annotationImageKeysQueryKey(datasetId),
    queryFn: async () => {
      const backendUrl =
        import.meta.env.VITE_BACKEND_URL || 'http://localhost:8085'
      const response = await fetch(
        `${backendUrl}/store/data/${datasetId}/imgs/`,
      )
      if (!response.ok) {
        throw new Error(`Failed to fetch image keys (${response.status})`)
      }
      const raw = await response.text()
      return parseImageKeys(raw)
    },
    enabled: !!datasetId,
  })
}
