import { useState, useEffect } from 'react'
import type {
  annotationrule_Segment,
  annotationrule_SlicePages,
  annotationrule_Stretch,
  annotationrule_AddMargin,
  annotationrule_LinesDetect,
  annotationrule_RemoveCategories,
  annotationrule_RemoveOverlap,
  annotationrule_ReassignTextLinesByTolerance,
  annotationrule_TextBlockCorrections,
} from '../api'

type AnnotationRule =
  | annotationrule_Segment
  | annotationrule_SlicePages
  | annotationrule_Stretch
  | annotationrule_AddMargin
  | annotationrule_LinesDetect
  | annotationrule_RemoveCategories
  | annotationrule_RemoveOverlap
  | annotationrule_ReassignTextLinesByTolerance
  | annotationrule_TextBlockCorrections

interface RuleEditModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (
    payload: AnnotationRule,
    action: 'overwrite' | 'create_new',
  ) => void
  initialPayload: AnnotationRule
  ruleType: AnnotationRule['type'] | undefined
}

export function RuleEditModal({
  isOpen,
  onClose,
  onSubmit,
  initialPayload,
  ruleType,
}: RuleEditModalProps) {
  const [payload, setPayload] = useState('')
  const [action, setAction] = useState<'overwrite' | 'create_new'>('overwrite')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (isOpen && initialPayload) {
      // Remove _index from the payload and keep type non-editable
      const { _index, type, ...editablePayload } = initialPayload as any
      setPayload(JSON.stringify(editablePayload, null, 2))
      setError(null)
    }
  }, [isOpen, initialPayload])

  const handleSubmit = () => {
    try {
      const parsedPayload = JSON.parse(payload)
      // Add the type back to the payload
      const fullPayload = { ...parsedPayload, type: ruleType } as AnnotationRule
      onSubmit(fullPayload, action)
      onClose()
    } catch (e) {
      setError('Invalid JSON format')
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50">
      <div className="bg-white rounded-lg max-w-2xl w-full max-h-[80vh] flex flex-col m-4">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">
            Edit Rule:{' '}
            {ruleType
              ?.split('_')
              .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
              .join(' ')}
          </h2>
        </div>

        <div className="flex-1 overflow-auto p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Rule Parameters (type: {ruleType})
              </label>
              <textarea
                value={payload}
                onChange={(e) => {
                  setPayload(e.target.value)
                  setError(null)
                }}
                className="w-full h-64 p-3 border border-gray-300 rounded-md font-mono text-sm focus:ring-blue-500 focus:border-blue-500"
                spellCheck={false}
              />
              {error && (
                <div className="mt-2 text-sm text-red-600">{error}</div>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Action
              </label>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="radio"
                    value="overwrite"
                    checked={action === 'overwrite'}
                    onChange={() => setAction('overwrite')}
                    className="mr-2"
                  />
                  <span className="text-sm">
                    <span className="font-medium">Overwrite</span> - Replace
                    current annotation
                  </span>
                </label>
                <label className="flex items-center">
                  <input
                    type="radio"
                    value="create_new"
                    checked={action === 'create_new'}
                    onChange={() => setAction('create_new')}
                    className="mr-2"
                  />
                  <span className="text-sm">
                    <span className="font-medium">Create New</span> - Create a
                    new annotation
                  </span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
          >
            Cancel
          </button>
          <button
            onClick={handleSubmit}
            className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
          >
            Run Rule
          </button>
        </div>
      </div>
    </div>
  )
}
