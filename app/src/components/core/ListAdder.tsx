import type { ReactNode } from 'react'
import { Button } from './Button'

type ListAdderProps<T extends { id: number }> = {
  label: string
  items: T[]
  onAdd: () => void
  renderItem: (item: T) => ReactNode
  emptyLabel?: string
  addLabel?: string
  isDisabled?: boolean
}

export function ListAdder<T extends { id: number }>({
  label,
  items,
  onAdd,
  renderItem,
  emptyLabel = 'No items added.',
  addLabel = '+ Add',
  isDisabled = false,
}: ListAdderProps<T>) {
  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <label className="block text-sm font-medium text-gray-700">
          {label}
        </label>
        <Button
          type="button"
          className="px-2 py-1 text-xs"
          onClick={onAdd}
          disabled={isDisabled}
        >
          {addLabel}
        </Button>
      </div>
      {items.length === 0 ? (
        <div className="text-xs text-gray-500">{emptyLabel}</div>
      ) : (
        <div className="space-y-2">
          {items.map((item) => (
            <div key={item.id}>{renderItem(item)}</div>
          ))}
        </div>
      )}
    </div>
  )
}
