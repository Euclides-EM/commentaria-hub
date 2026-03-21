import Select, { type ActionMeta } from 'react-select'
import { RangeInput } from '../../../core/RangeInput.tsx'
import type { FeatureOption } from './TeiPane.types.ts'
import {
  featureSelectStyles,
  formatFeatureOptionLabel,
} from './teiPaneUtils.tsx'

const RESET_TO_DEFAULTS_OPTION: FeatureOption = {
  value: '__reset_to_defaults__',
  label: 'Reset to defaults',
  description: '',
  isAction: true,
}

const SELECT_ALL_OPTION: FeatureOption = {
  value: '__select_all__',
  label: 'Select all',
  description: '',
  isAction: true,
}

type TeiDisplayControlsProps = {
  showMinCertControl: boolean
  minCert: number
  setMinCert: (value: number) => void
  showTeiLineHighlights: boolean
  setShowTeiLineHighlights: (value: boolean) => void
  alignLines: boolean
  setAlignLines: (value: boolean) => void
  showCertaintyVisualization: boolean
  setShowCertaintyVisualization: (value: boolean) => void
  allFeatureOptions: FeatureOption[]
  selectedFeatureOptions: FeatureOption[]
  isFeatureSelectExpanded: boolean
  setIsFeatureSelectExpanded: (value: boolean) => void
  setVisibleFeatureIds: (value: string[]) => void
  onResetVisibleFeatureIds: () => void
  isFeaturesLoading: boolean
}

export function TeiDisplayControls({
  showMinCertControl,
  minCert,
  setMinCert,
  showTeiLineHighlights,
  setShowTeiLineHighlights,
  alignLines,
  setAlignLines,
  showCertaintyVisualization,
  setShowCertaintyVisualization,
  allFeatureOptions,
  selectedFeatureOptions,
  isFeatureSelectExpanded,
  setIsFeatureSelectExpanded,
  setVisibleFeatureIds,
  onResetVisibleFeatureIds,
  isFeaturesLoading,
}: TeiDisplayControlsProps) {
  const featureSelectOptions = [
    RESET_TO_DEFAULTS_OPTION,
    SELECT_ALL_OPTION,
    ...allFeatureOptions,
  ]

  const handleFeatureSelectChange = (
    options: readonly FeatureOption[],
    actionMeta: ActionMeta<FeatureOption>,
  ) => {
    const selectedOption = actionMeta.option
    if (selectedOption?.value === RESET_TO_DEFAULTS_OPTION.value) {
      onResetVisibleFeatureIds()
      return
    }
    if (selectedOption?.value === SELECT_ALL_OPTION.value) {
      setVisibleFeatureIds(allFeatureOptions.map((option) => option.value))
      return
    }
    setVisibleFeatureIds(
      options
        .filter((option) => !option.isAction)
        .map((option) => option.value),
    )
  }

  return (
    <>
      <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
        <input
          type="checkbox"
          checked={showTeiLineHighlights}
          onChange={(event) => setShowTeiLineHighlights(event.target.checked)}
          className="rounded border-gray-300"
        />
        <span>Line highlights</span>
      </label>
      <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
        <input
          type="checkbox"
          checked={alignLines}
          onChange={(event) => setAlignLines(event.target.checked)}
          className="rounded border-gray-300"
        />
        <span>Align lines</span>
      </label>
      {showMinCertControl && (
        <>
          <RangeInput
            label="Min certainty"
            value={minCert}
            min={0.8}
            max={1}
            step={0.001}
            title="Hide tokens below certainty threshold"
            onChange={(value) => setMinCert(Math.round(value * 1000) / 1000)}
          />
          <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
            <input
              type="checkbox"
              checked={showCertaintyVisualization}
              onChange={(event) =>
                setShowCertaintyVisualization(event.target.checked)
              }
              className="rounded border-gray-300"
            />
            <span>Certainty heatmap</span>
          </label>
        </>
      )}
      {allFeatureOptions.length > 0 && (
        <>
          <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
            <input
              type="checkbox"
              checked={isFeatureSelectExpanded}
              onChange={(event) =>
                setIsFeatureSelectExpanded(event.target.checked)
              }
              className="rounded border-gray-300"
            />
            <span>Features select</span>
          </label>
          {isFeatureSelectExpanded && (
            <div className="flex items-center gap-1.5 min-w-65">
              <div className="flex-1 min-w-65">
                <Select<FeatureOption, true>
                  isMulti
                  value={selectedFeatureOptions}
                  onChange={handleFeatureSelectChange}
                  options={featureSelectOptions}
                  closeMenuOnSelect={false}
                  hideSelectedOptions
                  isLoading={isFeaturesLoading}
                  placeholder="Select features"
                  styles={featureSelectStyles}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                  formatOptionLabel={(option, { context }) =>
                    formatFeatureOptionLabel(option, context)
                  }
                  isOptionSelected={(option, selectValue) =>
                    option.isAction
                      ? false
                      : selectValue.some(
                          (selected) => selected.value === option.value,
                        )
                  }
                />
              </div>
            </div>
          )}
        </>
      )}
    </>
  )
}
