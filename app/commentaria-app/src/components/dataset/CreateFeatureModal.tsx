import { type FormEvent, useEffect, useMemo, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { FeaturesService } from '@hub-api'
import { HexColorPicker } from 'react-colorful'
import Select from 'react-select'
import { useFeaturePropertiesQuery } from '../../queries/datasets.ts'
import { selectStyles } from '../../styles/selectStyles.ts'
import {
  getFeaturePropertyDisplayName,
  normalizeFeatureProperties,
} from '../../utils/featureProperties.ts'
import { Button } from '../core/Button.tsx'
import { ErrorMessage } from '../core/ErrorMessage.tsx'

interface CreateFeatureModalProps {
  isOpen: boolean
  onClose: () => void
  datasetId: string
}

type FeaturePropertyOption = {
  value: string
  label: string
}

const createRandomPastelColor = () => {
  const hue = Math.floor(Math.random() * 360)
  const saturation = 60 + Math.random() * 20
  const lightness = 78 + Math.random() * 10
  const s = saturation / 100
  const l = lightness / 100
  const c = (1 - Math.abs(2 * l - 1)) * s
  const x = c * (1 - Math.abs(((hue / 60) % 2) - 1))
  const m = l - c / 2
  let r = 0
  let g = 0
  let b = 0

  if (hue < 60) {
    r = c
    g = x
  } else if (hue < 120) {
    r = x
    g = c
  } else if (hue < 180) {
    g = c
    b = x
  } else if (hue < 240) {
    g = x
    b = c
  } else if (hue < 300) {
    r = x
    b = c
  } else {
    r = c
    b = x
  }

  const toHex = (value: number) =>
    Math.round((value + m) * 255)
      .toString(16)
      .padStart(2, '0')

  return `#${toHex(r)}${toHex(g)}${toHex(b)}`.toLowerCase()
}

const normalizeHexInput = (value: string) => {
  const cleaned = value.trim().replace(/^#/, '')
  if (!cleaned) return ''
  return `#${cleaned.toLowerCase()}`
}

export function CreateFeatureModal({
  isOpen,
  onClose,
  datasetId,
}: CreateFeatureModalProps) {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [color, setColor] = useState(createRandomPastelColor())
  const [properties, setProperties] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const { data: availableProperties = [], isLoading: isLoadingProperties } =
    useFeaturePropertiesQuery(isOpen)
  const propertyOptions = useMemo<FeaturePropertyOption[]>(
    () =>
      availableProperties.map((property) => ({
        value: property,
        label: getFeaturePropertyDisplayName(property),
      })),
    [availableProperties],
  )

  useEffect(() => {
    if (isOpen) {
      setName('')
      setDescription('')
      setColor(createRandomPastelColor())
      setProperties([])
      setError(null)
      setLoading(false)
    }
  }, [isOpen])

  const handleSubmit = async (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault()
    if (!name.trim()) {
      setError('Feature name is required.')
      return
    }
    try {
      setError(null)
      setLoading(true)
      await FeaturesService.postDatasetsFeatures({
        dataSetId: datasetId,
        feature: {
          name: name.trim(),
          description: description.trim() || undefined,
          color: color || undefined,
          is_default: false,
          properties: normalizeFeatureProperties(properties),
          is_list: true,
        },
      })
      await queryClient.invalidateQueries({
        queryKey: ['features', 'definitions', datasetId],
      })
      onClose()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create feature.')
    } finally {
      setLoading(false)
    }
  }

  if (!isOpen) return null

  return (
    <div
      className="fixed inset-0 bg-black/20 backdrop-blur-sm flex items-center justify-center z-50 text-start"
      onClick={loading ? undefined : onClose}
    >
      <form
        className="bg-white rounded-lg max-w-xl w-full max-h-[90vh] flex flex-col m-4"
        onClick={(e) => e.stopPropagation()}
        onSubmit={handleSubmit}
      >
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold">Create a feature</h2>
        </div>

        <div className="flex-1 overflow-auto p-6 text-sm">
          <div className="grid gap-8 md:grid-cols-[minmax(0,1fr)_240px]">
            <div className="space-y-4">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Name
                </label>
                <input
                  type="text"
                  autoComplete="on"
                  value={name}
                  required
                  onChange={(e) => setName(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-md"
                  disabled={loading}
                />
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Description (optional)
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full p-2 border border-gray-300 rounded-md"
                  rows={3}
                  disabled={loading}
                />
              </div>

              <div className="space-y-2">
                <label className="block text-sm font-medium text-gray-700">
                  Properties
                </label>
                <Select<FeaturePropertyOption, true>
                  isMulti
                  value={propertyOptions.filter((option) =>
                    properties.includes(option.value),
                  )}
                  onChange={(options) =>
                    setProperties(
                      normalizeFeatureProperties(
                        (options || []).map((option) => option.value),
                      ),
                    )
                  }
                  options={propertyOptions}
                  closeMenuOnSelect={false}
                  hideSelectedOptions={false}
                  isLoading={isLoadingProperties}
                  isDisabled={loading || isLoadingProperties}
                  placeholder={
                    isLoadingProperties
                      ? 'Loading feature properties...'
                      : availableProperties.length === 0
                        ? 'No feature properties available'
                        : 'Select properties'
                  }
                  noOptionsMessage={() => 'No feature properties available'}
                  styles={selectStyles<FeaturePropertyOption, true>({
                    controlWidth: '100%',
                    isMulti: true,
                  })}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                />
              </div>

              <ErrorMessage message={error} />
            </div>

            <div className="space-y-2">
              <label className="block text-sm font-medium text-gray-700">
                Color
              </label>
              <div className="flex flex-col items-start gap-3">
                <span
                  className="w-5 h-5 rounded shrink-0 border border-gray-300"
                  style={{
                    backgroundColor: color || '#f2f2f2',
                    boxShadow: 'inset 0 0 0 1px rgba(255,255,255,0.4)',
                  }}
                />
                <HexColorPicker
                  color={color || '#f2f2f2'}
                  onChange={setColor}
                  style={{ width: 220, height: 170 }}
                />
                <input
                  className="p-2 border border-gray-300 rounded-md text-sm w-[120px] font-mono"
                  value={color.replace(/^#/, '')}
                  onChange={(e) => setColor(normalizeHexInput(e.target.value))}
                  disabled={loading}
                  aria-label="Feature color hex value"
                />
              </div>
            </div>
          </div>
        </div>

        <div className="px-6 py-4 border-t border-gray-200 flex justify-end gap-2">
          <Button
            type="button"
            onClick={onClose}
            className="px-3 py-1.5 text-sm"
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            className="px-3 py-1.5 text-sm"
            disabled={loading}
          >
            {loading ? 'Creating...' : 'Create'}
          </Button>
        </div>
      </form>
    </div>
  )
}
