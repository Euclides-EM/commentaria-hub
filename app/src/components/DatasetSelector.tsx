import { useDatasetsQuery } from '../queries/datasets'

interface DatasetSelectorProps {
  selectedDatasetId: string
  onDatasetChange: (id: string) => void
}

export function DatasetSelector({
  selectedDatasetId,
  onDatasetChange,
}: DatasetSelectorProps) {
  const { data: datasets, isLoading, error } = useDatasetsQuery()

  if (isLoading) {
    return (
      <div>
        <label
          htmlFor="datasetId"
          className="block text-xs opacity-80 mb-1 ml-0.5"
        >
          Dataset
        </label>
        <select
          id="datasetId"
          className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
          disabled
        >
          <option>Loading datasets...</option>
        </select>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <label
          htmlFor="datasetId"
          className="block text-xs opacity-80 mb-1 ml-0.5"
        >
          Dataset
        </label>
        <select
          id="datasetId"
          className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
          disabled
        >
          <option>Failed to load datasets</option>
        </select>
      </div>
    )
  }

  return (
    <div>
      <label
        htmlFor="datasetId"
        className="block text-xs opacity-80 mb-1 ml-0.5"
      >
        Dataset
      </label>
      <select
        id="datasetId"
        className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
        value={selectedDatasetId}
        onChange={(e) => onDatasetChange(e.target.value)}
      >
        <option value="">Select dataset...</option>
        {datasets?.map((dataset) => (
          <option key={dataset.id} value={dataset.id}>
            {dataset.name || dataset.id}
          </option>
        ))}
      </select>
    </div>
  )
}
