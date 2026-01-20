type Props = {
  title: string
  isOn: boolean
  toggle: () => void
}

export const ToggleButton = ({ title, isOn, toggle }: Props) => (
  <button
    className="text-sm cursor-pointer rounded-md px-2 py-1 border border-gray-300 bg-gray-100"
    onClick={() => toggle()}
  >
    {isOn ? 'Hide' : 'Show'} {title}
  </button>
)
