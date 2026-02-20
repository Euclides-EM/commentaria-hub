import Select, {
  components,
  MultiValueGenericProps,
  GroupBase,
  MultiValue,
  SingleValue,
} from "react-select";
import CreatableSelect from "react-select/creatable";
import { TOOLTIP_FEATURES_HIGHLIGHT } from "../../map/MapTooltips.tsx";

type OptionLabelProps = {
  label: string;
  tooltip?: string;
};

const OptionLabel = ({ label, tooltip }: OptionLabelProps) => (
  <span
    data-tooltip-id={TOOLTIP_FEATURES_HIGHLIGHT}
    data-tooltip-content={tooltip}
  >
    {label}
  </span>
);

type MultiSelectProps = {
  name: string;
  options: string[];
  onChange: (values: string[]) => void;
  onBlur?: () => void;
  defaultValues?: string[];
  colors?: Record<string, string>;
  tooltips?: Record<string, string>;
  labelFn?: (opt: string) => string;
  value?: string[];
  className?: string;
  isCreatable?: boolean;
  placeholder?: string;
};

export const MultiSelect = ({
  name,
  options,
  onChange,
  onBlur,
  defaultValues,
  colors,
  tooltips,
  labelFn,
  value,
  className,
  isCreatable = false,
  placeholder,
}: MultiSelectProps) => {
  const SelectComponent = isCreatable ? CreatableSelect : Select;
  const displayLabel = (opt: string) => (labelFn ? labelFn(opt) : opt);

  return (
    <SelectComponent
      isMulti
      name={name}
      components={{
        MultiValueLabel: (
          props: MultiValueGenericProps<
            { value: string; label: string | JSX.Element },
            boolean,
            GroupBase<{ value: string; label: string | JSX.Element }>
          >,
        ) => (
          <components.MultiValueLabel
            {...props}
            innerProps={
              {
                ...props.innerProps,
                onMouseDown: (e: React.MouseEvent) => e.stopPropagation(),
              } as React.HTMLProps<HTMLDivElement>
            }
          />
        ),
      }}
      value={value?.map((v) => ({
        value: v,
        label: tooltips ? (
          <OptionLabel label={displayLabel(v)} tooltip={tooltips[v]} />
        ) : (
          displayLabel(v)
        ),
      }))}
      defaultValue={defaultValues?.map((v) => ({
        value: v,
        label: tooltips ? (
          <OptionLabel label={displayLabel(v)} tooltip={tooltips[v]} />
        ) : (
          displayLabel(v)
        ),
      }))}
      options={options.map((option) => ({
        value: option,
        label: tooltips ? (
          <OptionLabel
            label={displayLabel(option)}
            tooltip={tooltips[option]}
          />
        ) : (
          displayLabel(option)
        ),
      }))}
      className={`basic-multi-select ${className}`}
      classNamePrefix="select"
      onBlur={onBlur}
      onChange={(
        selected:
          | MultiValue<{ value: string; label: string | JSX.Element }>
          | SingleValue<{ value: string; label: string | JSX.Element }>,
      ) => {
        if (selected && Array.isArray(selected)) {
          onChange(selected.map((option) => option.value));
        } else {
          onChange([]);
        }
      }}
      placeholder={placeholder || `Select ${name}`}
      styles={{
        menu: (base) => ({
          ...base,
          zIndex: 9999,
        }),
        menuPortal: (base) => ({
          ...base,
          zIndex: 9999,
        }),
        option: (base, { data }) => ({
          ...base,
          color: "black",
          backgroundColor: colors?.[data.value] || base.backgroundColor,
        }),
        multiValue: (base, { data }) => ({
          ...base,
          backgroundColor: colors?.[data.value] || base.backgroundColor,
        }),
        multiValueLabel: (base) => ({
          ...base,
          color: "black",
          userSelect: "text",
          cursor: "text",
        }),
      }}
      menuPortalTarget={document.body}
    />
  );
};
