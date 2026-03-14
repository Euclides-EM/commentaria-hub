import Select from 'react-select'
import { RangeInput } from '../../../core/RangeInput.tsx'
import type { FeatureOption } from './TeiPane.types.ts'
import {
  featureSelectStyles,
  formatFeatureOptionLabel,
} from './teiPaneUtils.tsx'

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
  isFeaturesLoading,
}: TeiDisplayControlsProps) {
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
              <Select<FeatureOption, true>
                isMulti
                value={selectedFeatureOptions}
                onChange={(options) =>
                  setVisibleFeatureIds((options || []).map((o) => o.value))
                }
                options={allFeatureOptions}
                closeMenuOnSelect={false}
                hideSelectedOptions={false}
                isLoading={isFeaturesLoading}
                placeholder="Select features"
                styles={featureSelectStyles}
                menuPortalTarget={document.body}
                menuPosition="fixed"
                formatOptionLabel={(option, { context }) =>
                  formatFeatureOptionLabel(option, context)
                }
              />
            </div>
          )}
        </>
      )}
    </>
  )
}
