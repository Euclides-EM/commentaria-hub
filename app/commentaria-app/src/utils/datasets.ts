export type DatasetCompleteness = 'full' | 'partial'

export const DATASET_COMPLETENESS_ITEMS: DatasetCompleteness[] = [
  'full',
  'partial',
]

export const isFullDatasetPageRange = (pages?: string) => {
  const pageRange = pages?.trim() ?? ''
  // The backend omits `pages` when the dataset uses the entire facsimile.
  return pageRange === '' || /^1\s*-\s*[1-9]\d*$/.test(pageRange)
}

export const getDatasetCompleteness = (pages?: string): DatasetCompleteness =>
  isFullDatasetPageRange(pages) ? 'full' : 'partial'
