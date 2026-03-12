import Select from 'react-select'
import { createPortal } from 'react-dom'
import type { FeatureModalState, FeatureOption } from './TeiPane.types.ts'
import {
  featureModalStyles,
  formatFeatureOptionLabel,
} from './teiPaneUtils.tsx'

type FeatureHighlightModalProps = {
  state: FeatureModalState | null
  isOpen: boolean
  modalFeatureId: string
  allFeatureOptions: FeatureOption[]
  onChangeFeatureId: (value: string) => void
  onClose: () => void
  onSubmit: () => void
}

export const FeatureHighlightModal = ({
  state,
  isOpen,
  modalFeatureId,
  allFeatureOptions,
  onChangeFeatureId,
  onClose,
  onSubmit,
}: FeatureHighlightModalProps) => {
  if (!state || !isOpen) {
    return null
  }

  const modalFeatureOption =
    allFeatureOptions.find((option) => option.value === modalFeatureId) || null

  return createPortal(
    <div className="fixed inset-0 z-[12500] flex items-center justify-center bg-black/40 p-4">
      <div className="w-full max-w-md rounded-xl bg-white border border-gray-200 p-4 shadow-xl">
        <div className="text-sm font-semibold text-gray-900 mb-2">
          Highlight a feature
        </div>
        <div className="text-xs text-gray-600 mb-2">
          "{state.selection.surface}"
        </div>
        <Select<FeatureOption, false>
          value={modalFeatureOption}
          onChange={(option) => {
            onChangeFeatureId(option?.value || '')
          }}
          options={allFeatureOptions}
          isClearable={false}
          styles={featureModalStyles}
          menuPortalTarget={document.body}
          menuPosition="fixed"
          formatOptionLabel={(option, { context }) =>
            formatFeatureOptionLabel(option, context)
          }
        />
        <div className="mt-4 flex items-center justify-end gap-2">
          <button
            type="button"
            className="px-3 py-1.5 rounded border border-gray-300 text-gray-700 hover:bg-gray-50 text-sm"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            type="button"
            className="px-3 py-1.5 rounded border border-teal-300 text-teal-700 hover:bg-teal-50 text-sm"
            onClick={onSubmit}
            disabled={!modalFeatureId}
          >
            Apply
          </button>
        </div>
      </div>
    </div>,
    document.body,
  )
}
