import { useCallback, useEffect, useMemo } from 'react'
import {
  useDatasetImageKeysQuery,
  useDatasetsQuery,
} from '../queries/datasets.ts'
import { useAnnotationsQuery } from '../queries/annotations.ts'
import { parsePageEntries } from '../utils/pages.ts'
import { findMatchingImage, hasAnnotationPages } from '../utils/editions.ts'
import type { PageOrKey } from './AppStateContext'

const getDefaultPageOrKey = (availablePages: string[]): string => {
  if (!availablePages.length) return ''
  if (availablePages[0] !== '1') {
    return availablePages[0]
  }
  return availablePages[Math.floor(availablePages.length / 2)]
}

interface UseAnnotationSelectionOptions {
  datasetId: string
  annotationId: string
  currentPageOrKey: PageOrKey
  onAnnotationIdResolved: (annotationId: string) => void
  onPageOrKeyResolved: (pageOrKey: string) => void
}

export function useAnnotationSelection({
  datasetId,
  annotationId,
  currentPageOrKey,
  onAnnotationIdResolved,
  onPageOrKeyResolved,
}: UseAnnotationSelectionOptions) {
  const { data: datasets, refetch: refetchDatasets } = useDatasetsQuery()
  const { data: annotations, refetch: refetchAnnotations } =
    useAnnotationsQuery(datasetId)

  const dataset = useMemo(
    () => datasets?.find((d) => datasetId && d.id === datasetId) || null,
    [datasets, datasetId],
  )

  const annotation = useMemo(
    () =>
      annotations?.find((a) => annotationId && a.id === annotationId) || null,
    [annotations, annotationId],
  )
  const hasPages = hasAnnotationPages(annotation)
  const annotationPageEntries = useMemo(
    () => (annotation ? parsePageEntries(annotation.pages || '') : []),
    [annotation],
  )
  const shouldLoadImageKeys = !!annotation && !hasPages
  const { data: imageKeys = [] } = useDatasetImageKeysQuery(
    datasetId,
    shouldLoadImageKeys,
    annotationPageEntries.length > 0 ? annotationPageEntries : null,
  )
  const availablePageOrKeys = useMemo(() => {
    if (!annotation) {
      return []
    }
    if (annotationPageEntries.length > 0) {
      return [...new Set(annotationPageEntries)].sort((a, b) =>
        a.localeCompare(b, undefined, { numeric: true }),
      )
    }
    return imageKeys.map((image) => image.key)
  }, [annotation, annotationPageEntries, imageKeys])

  const refetch = useCallback(() => {
    refetchDatasets()
    refetchAnnotations()
  }, [refetchDatasets, refetchAnnotations])

  useEffect(() => {
    if (annotations?.length === 1) {
      onAnnotationIdResolved(annotations[0].id!)
    }
  }, [annotations, onAnnotationIdResolved])

  useEffect(() => {
    if (!annotation || !availablePageOrKeys.length) {
      return
    }
    if (availablePageOrKeys.includes(String(currentPageOrKey))) {
      return
    }
    if (!hasPages) {
      const matchedImage = findMatchingImage(
        String(currentPageOrKey),
        imageKeys,
      )
      if (matchedImage?.key) {
        onPageOrKeyResolved(matchedImage.key)
        return
      }
    }
    onPageOrKeyResolved(getDefaultPageOrKey(availablePageOrKeys))
  }, [
    annotation,
    availablePageOrKeys,
    hasPages,
    imageKeys,
    onPageOrKeyResolved,
    currentPageOrKey,
  ])

  return { dataset, annotation, refetch }
}
