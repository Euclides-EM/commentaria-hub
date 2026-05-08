import Select from "react-select";
import CreatableSelect from "react-select/creatable";

type SingleSelectProps = {
  name: string;
  options: Array<{ value: string | number; label: string; color?: string }>;
  value: string | number | null;
  onBlur?: () => void;
  onChange: (value: string | number | null) => void;
  placeholder?: string;
  className?: string;
  isCreatable?: boolean;
  formatOptionLabel?: (option: {
    value: string | number;
    label: string;
    color?: string;
  }) => React.ReactNode;
  stylesOverride?: Record<string, unknown>;
};

export const SingleSelect = ({
  name,
  options,
  value,
  onChange,
  onBlur,
  placeholder,
  className,
  isCreatable,
  formatOptionLabel,
  stylesOverride,
}: SingleSelectProps) => {
  const SelectComponent = isCreatable ? CreatableSelect : Select;
  return (
    <SelectComponent
      name={name}
      value={options.find((option) => option.value === value) || null}
      options={options}
      className={`basic-single-select ${className}`}
      classNamePrefix="select"
      onBlur={onBlur}
      onChange={(selected) => onChange(selected?.value || null)}
      placeholder={placeholder || `Select ${name}`}
      isClearable
      formatOptionLabel={formatOptionLabel}
      styles={{
        menu: (base) => ({
          ...base,
          zIndex: 9999,
        }),
        menuPortal: (base) => ({
          ...base,
          zIndex: 9999,
        }),
        placeholder: (provided) => ({
          ...provided,
          fontSize: "14px",
        }),
        option: (base) => ({
          ...base,
          color: "black",
          fontSize: "14px",
        }),
        singleValue: (base) => ({
          ...base,
          color: "black",
          fontSize: "14px",
        }),
        ...(stylesOverride || {}),
      }}
      menuPortalTarget={document.body}
    />
  );
};
