import { useEffect, useState } from 'react'
import {
  type AnnotationRule,
} from '../utils/rules.ts'
import {
  ApiError,
  type annotationrule_MetadataDetails,
  type annotationrule_Type,
} from '../api'
import Select from 'react-select'


interface RuleEditModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (
    payload: AnnotationRule,
    action: 'overwrite' | 'create_new',
  ) => Promise<void>
  initialPayload?: AnnotationRule // Make optional for manual creation
  ruleMetadata?: annotationrule_MetadataDetails[] // Add for manual creation
}

export function RuleEditModal({
  isOpen,
  onClose,
  onSubmit,
  initialPayload,
  ruleMetadata,
}: RuleEditModalProps) {
  const [payload, setPayload] = useState('')
  const [action, setAction] = useState<'overwrite' | 'create_new'>('overwrite')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [selectedRuleType, setSelectedRuleType] = useState<annotationrule_Type | undefined>(
    initialPayload?.type || ruleMetadata?.[0]?.type,
  )

  useEffect(() => {
    if (isOpen) {
      setError(null)
      let currentPayload: AnnotationRule | object = {}

      if (initialPayload) {
        // Editing an existing suggested rule
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const { type, ...editablePayload } = initialPayload
        currentPayload = editablePayload
        setSelectedRuleType(initialPayload.type)
      } else if (selectedRuleType && ruleMetadata) {
        // Creating a new manual rule, use default payload from metadata
        const metadata = ruleMetadata.find(meta => meta.type === selectedRuleType);
        if (metadata?.default) {
          currentPayload = metadata.default;
        }
      }
      setPayload(JSON.stringify(currentPayload, null, 2))
    }
  }, [isOpen, initialPayload, selectedRuleType, ruleMetadata])

  const handleSubmit = async () => {
    if (!selectedRuleType) {
      setError('Please select a rule type.')
      return;
    }

    try {
      setError(null)
      setLoading(true)
      const parsedPayload = JSON.parse(payload)
      const fullPayload = {
        ...parsedPayload,
        type: selectedRuleType,
      } as AnnotationRule
      await onSubmit(fullPayload, action)
      onClose()
    } catch (e) {
      console.error('Failed to run rule:', e)
      setError(e instanceof ApiError ? e.body : String(e))
    } finally {
      setLoading(false)
    }
  }

  const ruleOptions = ruleMetadata?.map(meta => ({
    value: meta.type,
    label: meta.type
      ?.split('_')
      .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
      .join(' ') || String(meta.type),
  })) || [];

  if (!isOpen) {
    return null
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={loading ? undefined : onClose}
    >
      <div
        className="bg-white rounded-lg max-w-2xl w-full max-h-[80vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">
            {initialPayload ? 'Edit Rule' : 'Run Manual Rule'}
          </h2>
          {!initialPayload && ruleMetadata && (
            <div className="mt-2">
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Select Rule Type
              </label>
              <Select
                options={ruleOptions}
                value={ruleOptions.find(option => option.value === selectedRuleType)}
                onChange={(option) => setSelectedRuleType(option?.value)}
                isDisabled={loading}
                className="text-sm"
              />
            </div>
          )}
           {initialPayload && (
            <p className="text-sm text-gray-600 mt-1">
              Type:{' '}
              {initialPayload.type
                ?.split('_')
                .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
                .join(' ')}
            </p>
          )}
        </div>

        <div className="flex-1 overflow-auto p-6">
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Parameters (type: {selectedRuleType})
              </label>
              <textarea
                value={payload}
                onChange={(e) => {
                  setPayload(e.target.value)
                  setError(null)
                }}
                className="w-full h-64 p-3 border border-gray-300 rounded-md font-mono text-sm focus:ring-blue-500 focus:border-blue-500"
                spellCheck={false}
                disabled={loading}
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
                    disabled={loading}
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
                    disabled={loading}
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
          {loading ? (
            <div className="flex items-center gap-3">
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600"></div>
              <span className="text-sm text-gray-600">Applying rule...</span>
            </div>
          ) : (
            <>
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
            </>
          )}
        </div>
      </div>
    </div>
  )
}
