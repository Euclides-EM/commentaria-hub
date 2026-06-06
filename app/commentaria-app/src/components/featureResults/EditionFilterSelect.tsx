import Select from 'react-select'
import { wrappedOptionSelectStyles } from '../../styles/selectStyles'

export type EditionFilterOption = {
  value: string
  label: string
}

type EditionFilterSelectProps = {
  value: EditionFilterOption | null
  onChange: (option: EditionFilterOption | null) => void
  options: EditionFilterOption[]
}

const editionFilterSelectStyles =
  wrappedOptionSelectStyles<EditionFilterOption>({
    controlWidth: 320,
  })

export function EditionFilterSelect({
  value,
  onChange,
  options,
}: EditionFilterSelectProps) {
  if (options.length === 0) {
    return null
  }

  return (
    <Select<EditionFilterOption>
      value={value}
      onChange={(option) => onChange(option)}
      options={options}
      placeholder="Filter by edition..."
      styles={editionFilterSelectStyles}
      menuPortalTarget={document.body}
      menuPosition="fixed"
      isClearable
    />
  )
}
