interface TabButtonProps {
  onSelected: () => void
  title: string
  isActive: boolean
}

export function TabButton({ onSelected, title, isActive }: TabButtonProps) {
  return (
    <button
      className={`px-3 py-1 rounded w-45 text-sm ${
        isActive
          ? 'bg-gray-500  text-white !cursor-default'
          : 'bg-gray-200 hover:bg-gray-300'
      }`}
      onClick={() => onSelected()}
    >
      {isActive && '> '}
      {title}
    </button>
  )
}
