import { useEffect, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

interface MultiSelectDropdownProps<T> {
  allItems: T[]
  selectedItems: T[] | null
  setSelectedItems: (items: T[] | null) => void
  itemsLabel: string
  getItemLabel: (item: T) => string
  getPickerLabel?: (args: {
    allItems: T[]
    selectedItems: T[] | null
    itemsLabel: string
    getItemLabel: (item: T) => string
  }) => string
  getItemKey?: (item: T) => string
  bulkActionItems?: T[]
  bulkActionLabel?: string
  showSeparatorBeforeItem?: (item: T) => boolean
  minWidth?: string
  disabled?: boolean
}

export function MultiSelectDropdown<T>({
  allItems,
  selectedItems,
  setSelectedItems,
  itemsLabel,
  getItemLabel,
  getPickerLabel,
  getItemKey,
  bulkActionItems,
  bulkActionLabel,
  showSeparatorBeforeItem,
  minWidth = '160px',
  disabled = false,
}: MultiSelectDropdownProps<T>) {
  const [isOpen, setIsOpen] = useState(false)
  const isDisabled = disabled || allItems.length === 0
  const buttonRef = useRef<HTMLButtonElement | null>(null)
  const [menuRect, setMenuRect] = useState<DOMRect | null>(null)

  useEffect(() => {
    if (!isOpen) return

    const updateRect = () => {
      const rect = buttonRef.current?.getBoundingClientRect()
      setMenuRect(rect ?? null)
    }

    updateRect()
    window.addEventListener('resize', updateRect)
    window.addEventListener('scroll', updateRect, true)
    return () => {
      window.removeEventListener('resize', updateRect)
      window.removeEventListener('scroll', updateRect, true)
    }
  }, [isOpen])

  const menuStyle = useMemo(() => {
    if (!menuRect) return undefined
    const maxWidth = Math.max(240, menuRect.width)
    const width = Math.min(window.innerWidth - 16, maxWidth)
    const left = Math.min(menuRect.left, window.innerWidth - width - 8)
    return {
      position: 'fixed' as const,
      display: 'inline-table' as const,
      top: menuRect.bottom + 4,
      left,
      width,
      zIndex: 50,
    }
  }, [menuRect])

  const effectiveBulkActionItems = useMemo(
    () => bulkActionItems ?? allItems,
    [allItems, bulkActionItems],
  )

  const handleBulkSelection = (selectAll: boolean) => {
    const currentSelected = selectedItems == null ? allItems : selectedItems
    const preservedItems = currentSelected.filter(
      (item) => !effectiveBulkActionItems.includes(item),
    )
    const nextItems = selectAll
      ? [
          ...preservedItems,
          ...effectiveBulkActionItems.filter(
            (item) => !preservedItems.includes(item),
          ),
        ]
      : preservedItems

    setSelectedItems(
      nextItems.length === allItems.length ? allItems : nextItems,
    )
  }

  const pickerLabel = () => {
    if (getPickerLabel) {
      return getPickerLabel({
        allItems,
        selectedItems,
        itemsLabel,
        getItemLabel,
      })
    }
    if (selectedItems == null || selectedItems.length === allItems.length) {
      return `All ${itemsLabel}`
    }
    if (selectedItems.length === 0) {
      return 'None'
    }
    if (selectedItems.length === 1) {
      return getItemLabel(selectedItems[0])
    }
    return `${selectedItems.length} ${itemsLabel}`
  }

  const toggleItem = (item: T) => {
    if (selectedItems == null || selectedItems.includes(item)) {
      setSelectedItems(
        (selectedItems == null ? allItems : selectedItems).filter(
          (s) => s !== item,
        ),
      )
    } else {
      const nextStages = [...selectedItems, item]
      setSelectedItems(
        nextStages.length === allItems.length ? allItems : nextStages,
      )
    }
  }

  return (
    <div className="relative" style={{ minWidth }}>
      <button
        className={`flex items-center justify-between w-full px-2 py-1 text-sm bg-white border border-gray-400 rounded-md transition-colors ${isDisabled ? 'cursor-not-allowed opacity-60' : 'hover:border-gray-500 focus:border-blue-500 focus:outline-none focus:ring-3 focus:ring-blue-100'}`}
        style={{ height: '32px' }}
        onClick={() => {
          if (isDisabled) return
          setIsOpen(!isOpen)
        }}
        ref={buttonRef}
        disabled={isDisabled}
      >
        <span className="text-gray-700">{pickerLabel()}</span>
        <svg
          className={`w-4 h-4 text-gray-600 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>

      {isOpen &&
        !isDisabled &&
        menuStyle &&
        createPortal(
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setIsOpen(false)}
            />
            <div
              className="bg-white border border-gray-300 rounded-md shadow-lg z-50 max-h-64 overflow-y-auto"
              style={menuStyle}
            >
              {allItems.length === 0 ? (
                <div className="px-3 py-2 text-sm text-gray-500">
                  No options
                </div>
              ) : (
                <>
                  <button
                    className="pl-[32px] w-full px-3 py-2 text-gray-600 break-words hover:bg-gray-50 cursor-pointer text-sm text-start"
                    onClick={() => handleBulkSelection(true)}
                  >
                    {bulkActionLabel
                      ? `Select all ${bulkActionLabel}`
                      : 'Select all'}
                  </button>
                  <button
                    className="pl-[32px] w-full px-3 py-2 text-gray-600 break-words hover:bg-gray-50 cursor-pointer text-sm text-start"
                    onClick={() => handleBulkSelection(false)}
                  >
                    {bulkActionLabel
                      ? `Deselect all ${bulkActionLabel}`
                      : 'Clear all'}
                  </button>
                  {allItems.map((item) => (
                    <label
                      key={getItemKey ? getItemKey(item) : String(item)}
                      className={`flex items-center gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer text-sm ${
                        showSeparatorBeforeItem?.(item)
                          ? 'border-t border-gray-200 mt-1 pt-3'
                          : ''
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={
                          selectedItems == null || selectedItems.includes(item)
                        }
                        onChange={() => toggleItem(item)}
                        className="text-blue-600"
                      />
                      <span className="text-gray-700 break-words">
                        {getItemLabel(item)}
                      </span>
                    </label>
                  ))}
                </>
              )}
            </div>
          </>,
          document.body,
        )}
    </div>
  )
}
