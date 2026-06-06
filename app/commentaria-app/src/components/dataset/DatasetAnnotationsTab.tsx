import { AnnotationsTable } from '../annotations/AnnotationsTable.tsx'

interface DatasetAnnotationsTabProps {
  datasetId: string
}

export function DatasetAnnotationsTab({
  datasetId,
}: DatasetAnnotationsTabProps) {
  return <AnnotationsTable datasetIds={[datasetId]} />
}
