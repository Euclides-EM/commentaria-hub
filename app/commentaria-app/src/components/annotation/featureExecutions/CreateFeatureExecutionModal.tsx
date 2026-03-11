import { type FormEvent, useMemo, useState } from 'react'
import type {
  feature_ExecutionSkipIf,
  feature_Feature,
  model_Edition,
} from '@hub-api'
import { Button } from '../../core/Button.tsx'
import { SearchInput } from '../../core/SearchInput.tsx'
import { ErrorMessage } from '../../core/ErrorMessage.tsx'
import { formatEditionLabel } from '../../../utils/editions.ts'

interface CreateFeatureExecutionModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (values: {
    selectedFeatureIds: string[]
    selectedKeys: string[]
    skipIf: feature_ExecutionSkipIf[]
    pushToOrigin: boolean
  }) => Promise<void>
  features: feature_Feature[]
  editionItems: model_Edition[]
  skipIfOptions: feature_ExecutionSkipIf[]
  skipIfLabels: Record<feature_ExecutionSkipIf, string>
  loadingFeatures: boolean
  loadingEditions: boolean
  isSubmitting: boolean
  errorMessage: string | null
}

const sortByNewestRevision = (feature: feature_Feature) => {
  const revisions = [...(feature.revisions ?? [])]
  revisions.sort((left, right) => {
    const leftTime = left.created_at ? new Date(left.created_at).getTime() : 0
    const rightTime = right.created_at
      ? new Date(right.created_at).getTime()
      : 0
    return rightTime - leftTime
  })
  return revisions[0]
}

const splitSearchQuery = (query: string) =>
  Array.from(query.toLowerCase().matchAll(/"([^"]+)"|(\S+)/g))
    .map((match) => match[1] ?? match[2] ?? '')
    .filter(Boolean)

export function CreateFeatureExecutionModal({
  isOpen,
  onClose,
  onSubmit,
  features,
  editionItems,
  skipIfOptions,
  skipIfLabels,
  loadingFeatures,
  loadingEditions,
  isSubmitting,
  errorMessage,
}: CreateFeatureExecutionModalProps) {
  if (!isOpen) return null

  return (
    <OpenCreateFeatureExecutionModal
      onClose={onClose}
      onSubmit={onSubmit}
      features={features}
      editionItems={editionItems}
      skipIfOptions={skipIfOptions}
      skipIfLabels={skipIfLabels}
      loadingFeatures={loadingFeatures}
      loadingEditions={loadingEditions}
      isSubmitting={isSubmitting}
      errorMessage={errorMessage}
    />
  )
}

function OpenCreateFeatureExecutionModal({
  onClose,
  onSubmit,
  features,
  editionItems,
  skipIfOptions,
  skipIfLabels,
  loadingFeatures,
  loadingEditions,
  isSubmitting,
  errorMessage,
}: Omit<CreateFeatureExecutionModalProps, 'isOpen'>) {
  const [featureSearch, setFeatureSearch] = useState('')
  const [editionSearch, setEditionSearch] = useState('')
  const [selectedFeatureIds, setSelectedFeatureIds] = useState<Set<string>>(
    new Set(),
  )
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set())
  const [skipIf, setSkipIf] = useState<Set<feature_ExecutionSkipIf>>(new Set())
  const [pushToOrigin, setPushToOrigin] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)

  const filteredFeatures = useMemo(() => {
    const query = featureSearch.trim().toLowerCase()
    if (!query) return features
    return features.filter((feature) => {
      const name = feature.name?.toLowerCase() ?? ''
      const description = feature.description?.toLowerCase() ?? ''
      return name.includes(query) || description.includes(query)
    })
  }, [featureSearch, features])

  const filteredEditionItems = useMemo(() => {
    const queryParts = splitSearchQuery(editionSearch.trim())
    if (queryParts.length === 0) {
      return editionItems
    }
    return editionItems.filter((item) => {
      const haystack = [
        item.key,
        item.year,
        item.editor?.join(', '),
        item.cities?.join(', '),
        item.shortTitle,
        item.title,
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
      return queryParts.every((queryPart) => haystack.includes(queryPart))
    })
  }, [editionItems, editionSearch])

  const toggleFeatureSelection = (featureId: string) => {
    setSelectedFeatureIds((previous) => {
      const next = new Set(previous)
      if (next.has(featureId)) next.delete(featureId)
      else next.add(featureId)
      return next
    })
  }

  const toggleEditionSelection = (key: string) => {
    setSelectedKeys((previous) => {
      const next = new Set(previous)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  const toggleSkipIf = (value: feature_ExecutionSkipIf) => {
    setSkipIf((previous) => {
      const next = new Set(previous)
      if (next.has(value)) next.delete(value)
      else next.add(value)
      return next
    })
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (selectedFeatureIds.size === 0) {
      setValidationError('Select at least one feature.')
      return
    }
    if (selectedKeys.size === 0) {
      setValidationError('Select at least one edition.')
      return
    }
    setValidationError(null)
    await onSubmit({
      selectedFeatureIds: Array.from(selectedFeatureIds),
      selectedKeys: Array.from(selectedKeys),
      skipIf: Array.from(skipIf),
      pushToOrigin,
    })
  }

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={isSubmitting ? undefined : onClose}
    >
      <form
        className="bg-white rounded-lg max-w-5xl w-full max-h-[90vh] flex flex-col m-4"
        onClick={(event) => event.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200 flex items-center justify-between gap-2">
          <h2 className="text-lg font-semibold">Execute Features</h2>
          <div className="text-xs text-gray-500">
            {selectedFeatureIds.size} features, {selectedKeys.size} editions
            selected
          </div>
        </div>

        <div className="flex-1 overflow-auto p-6 space-y-5">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
            <section className="space-y-2">
              <div className="text-sm font-medium text-gray-700">Features</div>
              <div className="flex items-center gap-2">
                <SearchInput
                  value={featureSearch}
                  onChange={setFeatureSearch}
                  placeholder="Search features"
                  className="w-full"
                />
                <Button
                  type="button"
                  className="px-2 py-1 text-xs shrink-0"
                  onClick={() =>
                    setSelectedFeatureIds(
                      new Set(
                        filteredFeatures
                          .map((feature) => feature.id || '')
                          .filter(Boolean),
                      ),
                    )
                  }
                  disabled={loadingFeatures}
                >
                  Select all
                </Button>
                <Button
                  type="button"
                  className="px-2 py-1 text-xs shrink-0"
                  onClick={() => setSelectedFeatureIds(new Set())}
                  disabled={loadingFeatures}
                >
                  Clear
                </Button>
              </div>
              <div className="border border-gray-200 rounded-md max-h-72 overflow-auto divide-y divide-gray-100">
                {loadingFeatures ? (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    Loading features...
                  </div>
                ) : filteredFeatures.length === 0 ? (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    No features found.
                  </div>
                ) : (
                  filteredFeatures.map((feature) => {
                    const featureId = feature.id || ''
                    if (!featureId) return null
                    const latestRevision = sortByNewestRevision(feature)
                    return (
                      <label
                        key={featureId}
                        className="flex items-center gap-2 px-3 py-2 text-xs text-gray-700"
                      >
                        <input
                          type="checkbox"
                          checked={selectedFeatureIds.has(featureId)}
                          onChange={() => toggleFeatureSelection(featureId)}
                        />
                        <div className="flex items-center gap-2 min-w-0">
                          <span
                            className="inline-block h-2.5 w-2.5 shrink-0 rounded-full border border-gray-300"
                            style={{
                              backgroundColor: feature.color || '#d1d5db',
                            }}
                          />
                          <span className="font-medium">
                            {feature.name || 'Untitled'}
                          </span>
                        </div>
                        {!latestRevision && (
                          <span className="text-red-500">(no revisions)</span>
                        )}
                      </label>
                    )
                  })
                )}
              </div>
            </section>

            <section className="space-y-2">
              <div className="text-sm font-medium text-gray-700">Editions</div>
              <div className="flex items-center gap-2">
                <SearchInput
                  value={editionSearch}
                  onChange={setEditionSearch}
                  placeholder="Search editions"
                  className="w-full"
                />
                <Button
                  type="button"
                  className="px-2 py-1 text-xs shrink-0"
                  onClick={() =>
                    setSelectedKeys(
                      new Set(filteredEditionItems.map((item) => item.key!)),
                    )
                  }
                  disabled={loadingEditions}
                >
                  Select all
                </Button>
                <Button
                  type="button"
                  className="px-2 py-1 text-xs shrink-0"
                  onClick={() => setSelectedKeys(new Set())}
                  disabled={loadingEditions}
                >
                  Clear
                </Button>
              </div>
              <div className="border border-gray-200 rounded-md max-h-72 overflow-auto divide-y divide-gray-100">
                {loadingEditions ? (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    Loading editions...
                  </div>
                ) : filteredEditionItems.length === 0 ? (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    No editions found.
                  </div>
                ) : (
                  filteredEditionItems.map((item) => (
                    <label
                      key={item.key}
                      className="flex items-start gap-2 px-3 py-2 text-xs text-gray-700"
                    >
                      <input
                        type="checkbox"
                        checked={selectedKeys.has(item.key!)}
                        onChange={() => toggleEditionSelection(item.key!)}
                      />
                      <span>{formatEditionLabel(item)}</span>
                    </label>
                  ))
                )}
              </div>
            </section>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
            <section className="space-y-2">
              <div className="text-sm font-medium text-gray-700">Skip if</div>
              <div className="flex flex-wrap gap-3">
                {skipIfOptions.map((option) => (
                  <label
                    key={option}
                    className="inline-flex items-center gap-2 text-xs text-gray-700"
                  >
                    <input
                      type="checkbox"
                      checked={skipIf.has(option)}
                      onChange={() => toggleSkipIf(option)}
                    />
                    <span>{skipIfLabels[option] || option}</span>
                  </label>
                ))}
              </div>
            </section>

            <section className="space-y-2">
              <label className="inline-flex items-center gap-2 text-xs text-gray-700">
                <input
                  type="checkbox"
                  checked={pushToOrigin}
                  onChange={(event) => setPushToOrigin(event.target.checked)}
                  disabled={isSubmitting}
                />
                <span>Push to origin annotation if applicable</span>
              </label>
            </section>
          </div>

          <ErrorMessage message={validationError || errorMessage} />
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm"
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            className="px-3 py-1.5 text-sm"
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Submitting...' : 'Submit Execution'}
          </Button>
        </div>
      </form>
    </div>
  )
}
