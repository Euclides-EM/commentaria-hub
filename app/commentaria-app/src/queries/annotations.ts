import { useQuery } from '@tanstack/react-query'
import {
  AnnotationsService,
  EditionsService,
  type annotation_SearchWithin,
} from '@hub-api'

const MAX_BULK_TEI_PAGES = 25

const chunkPages = (pages: string[], size: number) => {
  const chunks: string[][] = []
  for (let i = 0; i < pages.length; i += size) {
    chunks.push(pages.slice(i, i + size))
  }
  return chunks
}

const mergeTeisXmlResponses = (responses: string[]) => {
  const parser = new DOMParser()
  const serializer = new XMLSerializer()
  const merged = document.implementation.createDocument('', 'teis', null)
  const root = merged.documentElement

  responses.forEach((response) => {
    const xml = response.trim()
    if (!xml) return
    const doc = parser.parseFromString(xml, 'text/xml')
    const sourceRoot = doc.documentElement
    if (!sourceRoot) return
    Array.from(sourceRoot.children).forEach((child) => {
      root.appendChild(merged.importNode(child, true))
    })
  })

  return serializer.serializeToString(root)
}

const annotationsQueryKey = (datasetId: string) =>
  ['annotations', datasetId] as const

export function useAnnotationsQuery(datasetId: string) {
  return useQuery({
    queryKey: annotationsQueryKey(datasetId),
    queryFn: async () =>
      datasetId
        ? AnnotationsService.getDatasetsAnnotations({
            dataSetId: datasetId,
          })
        : Promise.resolve([]),
    enabled: !!datasetId,
    refetchInterval: 10 * 1000,
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
        pageNumOrKey: String(pageOrKey),
        fallbackToOrigin: true,
      }),
    enabled: !!datasetId && !!annotationId && !!pageOrKey && enabled,
  })
}

export const editionTeiQueryKey = (
  editionId: string,
  pageNum: number | string,
) => ['editions', editionId, 'tei', pageNum] as const

export function useEditionTeiQuery(
  editionId: string | null | undefined,
  pageNum: number | string,
  enabled: boolean,
) {
  return useQuery({
    queryKey: editionTeiQueryKey(editionId!, pageNum),
    queryFn: () =>
      EditionsService.getEditionsTei({
        editionId: editionId || '',
        pageNum: String(pageNum),
      }),
    enabled: !!editionId && !!pageNum && enabled,
    retry: false,
  })
}

export function useAnnotationSearch(
  datasetId: string,
  annotationId: string,
  regex: string,
  categories: string[] = [],
  searchWithin: annotation_SearchWithin[] = [],
) {
  return useQuery({
    queryKey: [
      'annotations',
      datasetId,
      annotationId,
      'search',
      { regex, categories, searchWithin },
    ] as const,
    queryFn: () =>
      AnnotationsService.getDatasetsAnnotationsSearch({
        dataSetId: datasetId,
        id: annotationId,
        regex,
        category: categories.length > 0 ? categories : undefined,
        searchWithin,
        fallbackToOrigin: true,
      }),
    enabled: !!datasetId && !!annotationId && regex.length > 0,
  })
}

export function useAnnotationTeisQuery(
  datasetId: string,
  annotationId: string,
  pages: string[],
  enabled: boolean,
) {
  return useQuery({
    queryKey: ['annotations', datasetId, annotationId, 'teis', pages],
    queryFn: async () => {
      const pageChunks = chunkPages(pages, MAX_BULK_TEI_PAGES)
      const responses = await Promise.all(
        pageChunks.map((pageNumOrKey) =>
          AnnotationsService.getDatasetsAnnotationsTeis({
            dataSetId: datasetId,
            id: annotationId,
            pageNumOrKey,
            fallbackToOrigin: true,
          }),
        ),
      )
      return mergeTeisXmlResponses(responses)
    },
    enabled: !!datasetId && !!annotationId && pages.length > 0 && enabled,
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
