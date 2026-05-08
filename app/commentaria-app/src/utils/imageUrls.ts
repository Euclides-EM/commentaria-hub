export type DatasetImageVariant = 'original' | 'preview' | 'thumb'

export function buildDatasetImageUrl(
  datasetId: string,
  imageIdentifier: string,
  variant: DatasetImageVariant,
  imageVersion?: number,
) {
  const url = new URL(
    `${import.meta.env.VITE_BACKEND_URL}/api/v1/datasets/${encodeURIComponent(datasetId)}/images/${encodeURIComponent(imageIdentifier)}`,
  )

  if (variant !== 'original') {
    url.searchParams.set('variant', variant)
  }
  if (typeof imageVersion === 'number') {
    url.searchParams.set('v', String(imageVersion))
  }

  return url.toString()
}
