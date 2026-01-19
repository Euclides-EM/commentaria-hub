import { useAnnotationsQuery } from '../queries/annotations'

interface AnnotationSelectorProps {
  datasetId: string
  onlyTranscribed: boolean
  selectedAnnotationId: string
  onAnnotationChange: (id: string) => void
}

export function AnnotationSelector({
  datasetId,
  onlyTranscribed,
  selectedAnnotationId,
  onAnnotationChange,
}: AnnotationSelectorProps) {
  const {
    data: annotations,
    isLoading,
    error,
  } = useAnnotationsQuery(datasetId, onlyTranscribed)

  if (!datasetId) {
    return (
      <div>
        <label
          htmlFor="annotationId"
          className="block text-xs opacity-80 mb-1 ml-0.5"
        >
          Annotation
        </label>
        <select
          id="annotationId"
          className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
          value=""
          disabled
        >
          <option value="">Select dataset first...</option>
        </select>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div>
        <label
          htmlFor="annotationId"
          className="block text-xs opacity-80 mb-1 ml-0.5"
        >
          Annotation
        </label>
        <select
          id="annotationId"
          className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
          disabled
        >
          <option>Loading annotations...</option>
        </select>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <label
          htmlFor="annotationId"
          className="block text-xs opacity-80 mb-1 ml-0.5"
        >
          Annotation
        </label>
        <select
          id="annotationId"
          className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
          disabled
        >
          <option>Failed to load annotations</option>
        </select>
      </div>
    )
  }

  return (
    <div>
      <label
        htmlFor="annotationId"
        className="block text-xs opacity-80 mb-1 ml-0.5"
      >
        Annotation
      </label>
      <select
        id="annotationId"
        className="w-full border border-gray-300 rounded-lg px-2.5 py-2 font-mono text-xs box-border"
        value={selectedAnnotationId}
        onChange={(e) => onAnnotationChange(e.target.value)}
      >
        <option value="">Select annotation...</option>
        {annotations?.map((annotation) => (
          <option key={annotation.id} value={annotation.id}>
            {annotation.name || annotation.id}
          </option>
        ))}
      </select>
    </div>
  )
}
