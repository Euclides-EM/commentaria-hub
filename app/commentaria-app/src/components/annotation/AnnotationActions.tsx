import { useState } from 'react'
import { Button } from '../core/Button.tsx'
import { ImportAnnotationsModal } from './ImportAnnotationsModal.tsx'
import { CreateAnnotationModal } from './CreateAnnotationModal.tsx'
import { useAuthStore } from '../../store/authStore.ts'

interface AnnotationEmptyStateProps {
  dataSetId: string
  onAnnotationSelected?: (annotationId: string) => void
}

export function AnnotationActions({
  dataSetId,
  onAnnotationSelected,
}: AnnotationEmptyStateProps) {
  const [isImportUrlOpen, setIsImportUrlOpen] = useState(false)
  const [isImportZipOpen, setIsImportZipOpen] = useState(false)
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const isAuthenticated = !!useAuthStore((store) => store.token)

  if (!isAuthenticated) {
    return null
  }

  return (
    <div className="w-[calc(100%-1.5rem)] max-w-[80vw] mx-auto mt-6 mb-3 font-medium text-left">
      <p>Annotation import</p>
      <div className="mt-4 flex flex-wrap gap-3 justify-start">
        <Button
          onClick={() => setIsImportUrlOpen(true)}
          variant="primary"
          className="px-3 py-2 text-sm w-40"
        >
          Import from URL
        </Button>
        <Button
          onClick={() => setIsImportZipOpen(true)}
          variant="primary"
          className="px-3 py-2 text-sm w-40"
        >
          Import from ZIP
        </Button>
        <Button
          onClick={() => setIsCreateOpen(true)}
          variant="primary"
          className="px-3 py-2 text-sm w-40"
        >
          Create new
        </Button>
      </div>
      <ImportAnnotationsModal
        isOpen={isImportUrlOpen}
        mode="url"
        dataSetId={dataSetId}
        onClose={() => setIsImportUrlOpen(false)}
        onImported={(annotationId) => {
          onAnnotationSelected?.(annotationId)
        }}
      />
      <ImportAnnotationsModal
        isOpen={isImportZipOpen}
        mode="zip"
        dataSetId={dataSetId}
        onClose={() => setIsImportZipOpen(false)}
        onImported={(annotationId) => {
          onAnnotationSelected?.(annotationId)
        }}
      />
      <CreateAnnotationModal
        isOpen={isCreateOpen}
        dataSetId={dataSetId}
        onClose={() => setIsCreateOpen(false)}
        onCreated={(annotationId) => {
          onAnnotationSelected?.(annotationId)
        }}
      />
    </div>
  )
}
