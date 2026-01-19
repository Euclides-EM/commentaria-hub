import { useMemo } from 'react'
import { useDatasetPageImageQuery } from './datasets'
import { useAnnotationTeiQuery } from './annotations'

export interface PageData {
  imageUrl: string
  teiData: string
  imgStatus: string
  teiStatus: string
}

export function usePageDataQuery(
  datasetId: string,
  annotationId: string,
  pageNum: number,
) {
  const imageQuery = useDatasetPageImageQuery(datasetId, pageNum)
  const teiQuery = useAnnotationTeiQuery(datasetId, annotationId, pageNum)

  const pageData = useMemo<PageData>(() => {
    const imageUrl = imageQuery.data ? URL.createObjectURL(imageQuery.data) : ''
    const imgStatus = imageQuery.isLoading
      ? 'Loading'
      : imageQuery.error
        ? 'Error'
        : imageQuery.data
          ? 'Loaded'
          : ''
    const teiStatus = teiQuery.isLoading
      ? 'Loading'
      : teiQuery.error
        ? 'Error'
        : teiQuery.data
          ? 'Loaded'
          : ''

    return {
      imageUrl,
      teiData: teiQuery.data || '',
      imgStatus,
      teiStatus,
    }
  }, [
    imageQuery.data,
    imageQuery.isLoading,
    imageQuery.error,
    teiQuery.data,
    teiQuery.isLoading,
    teiQuery.error,
  ])

  return {
    pageData,
    isLoading: imageQuery.isLoading || teiQuery.isLoading,
    error: imageQuery.error || teiQuery.error,
    refetch: () => {
      imageQuery.refetch()
      teiQuery.refetch()
    },
  }
}
