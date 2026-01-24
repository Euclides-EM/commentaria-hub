interface RangeInputProps {
  label: string
  value: number
  min: number
  max: number
  step: number
  onChange: (value: number) => void
  title?: string
  className?: string
}

const clamp = (value: number, min: number, max: number) =>
  Math.min(max, Math.max(min, value))

export function RangeInput({
  label,
  value,
  min,
  max,
  step,
  onChange,
  title,
  className,
}: RangeInputProps) {
  return (
    <div
      className={`flex items-center gap-2 px-1.5 py-1 border border-gray-200 rounded-lg bg-white ${className || ''}`}
      title={title}
    >
      <label className="text-xs opacity-75">{label}</label>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        className="w-40"
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
      />
      <input
        type="number"
        min={min}
        max={max}
        step={step}
        className="text-xs opacity-75 font-mono p-1 rounded-lg border border-gray-300"
        value={value}
        onChange={(e) => onChange(clamp(Number(e.target.value), min, max))}
      />
    </div>
  )
}
