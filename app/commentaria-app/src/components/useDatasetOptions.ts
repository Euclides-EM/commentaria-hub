import { useMemo } from 'react'
import useLocalStorageState from 'use-local-storage-state'
import { useDatasetsQuery } from '../queries/datasets.ts'
import {
  DATASET_COMPLETENESS_ITEMS,
  getDatasetCompleteness,
  type DatasetCompleteness,
} from '../utils/datasets.ts'

export function useDatasetOptions() {
  const { data: datasets, isLoading: datasetsLoading } = useDatasetsQuery()
  const [selectedDatasetCompleteness, setSelectedDatasetCompleteness] =
    useLocalStorageState<DatasetCompleteness[] | null>(
      'datasetCompletenessFilter',
      { defaultValue: null },
    )

  const datasetOptions = useMemo(() => {
    if (!datasets) return []
    const selectedCompleteness =
      selectedDatasetCompleteness ?? DATASET_COMPLETENESS_ITEMS
    return datasets
      .filter((d) => {
        if (!d.id) return false
        const completeness = getDatasetCompleteness(d.pages)
        return selectedCompleteness.includes(completeness)
      })
      .map((d) => ({
        value: d.id as string,
        label: d.name || (d.id as string),
      }))
      .sort((a, b) => a.label.localeCompare(b.label))
  }, [datasets, selectedDatasetCompleteness])

  return {
    datasets,
    datasetsLoading,
    datasetOptions,
    selectedDatasetCompleteness,
    setSelectedDatasetCompleteness,
  }
}
