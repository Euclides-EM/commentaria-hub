import { useMemo, useState } from 'react'
import { Button } from '../core/Button'
import { ErrorMessage } from '../core/ErrorMessage'

type GroupOption = {
  id: string
  label: string
}

interface AddToAnnotationGroupsModalProps {
  isOpen: boolean
  selectedCount: number
  groups: GroupOption[]
  isSubmitting: boolean
  error?: string | null
  onClose: () => void
  onSubmit: (groupIds: string[]) => Promise<void>
}

export function AddToAnnotationGroupsModal({
  isOpen,
  selectedCount,
  groups,
  isSubmitting,
  error,
  onClose,
  onSubmit,
}: AddToAnnotationGroupsModalProps) {
  if (!isOpen) {
    return null
  }

  return (
    <AddToAnnotationGroupsModalContent
      selectedCount={selectedCount}
      groups={groups}
      isSubmitting={isSubmitting}
      error={error}
      onClose={onClose}
      onSubmit={onSubmit}
    />
  )
}

type AddToAnnotationGroupsModalContentProps = Omit<
  AddToAnnotationGroupsModalProps,
  'isOpen'
>

function AddToAnnotationGroupsModalContent({
  selectedCount,
  groups,
  isSubmitting,
  error,
  onClose,
  onSubmit,
}: AddToAnnotationGroupsModalContentProps) {
  const [selectedGroupIds, setSelectedGroupIds] = useState<string[]>([])

  const selectedSet = useMemo(
    () => new Set(selectedGroupIds),
    [selectedGroupIds],
  )

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50"
      onClick={isSubmitting ? undefined : onClose}
    >
      <div
        className="bg-white rounded-lg w-full max-w-md m-4 shadow-lg"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-5 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Add to groups</h2>
        </div>
        <div className="px-5 py-4 space-y-3 text-sm text-gray-700">
          <p>
            Add <span className="font-semibold">{selectedCount}</span> selected{' '}
            {selectedCount === 1 ? 'annotation' : 'annotations'} to one or more
            groups.
          </p>
          <div className="max-h-56 overflow-auto rounded-md border border-gray-200 bg-white divide-y divide-gray-100">
            {groups.map((group) => (
              <label
                key={group.id}
                className="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={selectedSet.has(group.id)}
                  onChange={(e) => {
                    setSelectedGroupIds((current) =>
                      e.target.checked
                        ? [...current, group.id]
                        : current.filter((id) => id !== group.id),
                    )
                  }}
                  className="h-4 w-4"
                  disabled={isSubmitting}
                />
                <span>{group.label}</span>
              </label>
            ))}
          </div>
          <ErrorMessage message={error} />
        </div>
        <div className="px-5 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm"
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="primary"
            onClick={() => void onSubmit(selectedGroupIds)}
            className="px-3 py-1.5 text-sm"
            disabled={isSubmitting || selectedGroupIds.length === 0}
          >
            {isSubmitting ? 'Adding...' : 'Add'}
          </Button>
        </div>
      </div>
    </div>
  )
}
