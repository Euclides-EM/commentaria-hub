import type { StylesConfig } from 'react-select'

type SelectStylesConfig = {
  controlWidth: number | string
}

export const selectStyles = <
  OptionType extends { value: unknown; label: string },
>(
  config: Partial<SelectStylesConfig> = {},
): StylesConfig<OptionType, false> => ({
  control: (base, state) => ({
    ...base,
    minHeight: 32,
    height: 32,
    width: config.controlWidth,
    fontSize: '14px',
    border: `1px solid ${state.isFocused ? '#14b8a6' : '#9ca3af'}`,
    borderRadius: '6px',
    backgroundColor: 'white',
    boxShadow: state.isFocused ? '0 0 0 3px rgba(20, 184, 166, 0.15)' : 'none',
    '&:hover': {
      borderColor: state.isFocused ? '#14b8a6' : '#6b7280',
    },
  }),
  valueContainer: (base) => ({
    ...base,
    padding: '2px 8px',
    display: 'flex',
    alignItems: 'center',
    height: '30px',
  }),
  singleValue: (base) => ({
    ...base,
    color: '#374151',
    lineHeight: '1.5',
    margin: 0,
  }),
  input: (base) => ({
    ...base,
    margin: 0,
    paddingTop: 0,
    paddingBottom: 0,
  }),
  indicatorsContainer: (base) => ({
    ...base,
    height: '30px',
  }),
  indicatorSeparator: () => ({
    display: 'none',
  }),
  dropdownIndicator: (base) => ({
    ...base,
    padding: '4px 4px',
    color: '#6b7280',
    '&:hover': {
      color: '#374151',
    },
  }),
  clearIndicator: (base) => ({
    ...base,
    padding: '4px 4px',
    color: '#6b7280',
    '&:hover': {
      color: '#374151',
    },
  }),
  menu: (base) => ({
    ...base,
    minWidth: config.controlWidth || '200px',
    width: 'max-content',
    fontSize: '14px',
    border: '1px solid #d1d5db',
    boxShadow:
      '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
  }),
  option: (base, state) => ({
    ...base,
    fontSize: '14px',
    padding: '8px 12px',
    backgroundColor: state.isSelected
      ? '#14b8a6'
      : state.isFocused
        ? '#f3f4f6'
        : 'white',
    color: state.isSelected ? 'white' : '#374151',
    '&:hover': {
      backgroundColor: state.isSelected ? '#14b8a6' : '#f3f4f6',
    },
  }),
  placeholder: (base) => ({
    ...base,
    color: '#9ca3af',
    lineHeight: '1.5',
    margin: 0,
  }),
})
