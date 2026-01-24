import { useEffect, useMemo, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

interface MultiSelectDropdownProps<T> {
  allItems: T[]
  selectedItems: T[]
  onToggleItem: (item: T) => void
  getButtonLabel: (selectedItems: T[], allItems: T[]) => string
  getItemLabel: (item: T) => string
  getItemKey?: (item: T) => string
  minWidth?: string
  disabled?: boolean
}

export function MultiSelectDropdown<T>({
  allItems,
  selectedItems,
  onToggleItem,
  getButtonLabel,
  getItemLabel,
  getItemKey,
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
      top: menuRect.bottom + 4,
      left,
      width,
      zIndex: 50,
    }
  }, [menuRect])

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
        <span className="text-gray-700">
          {getButtonLabel(selectedItems, allItems)}
        </span>
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
                allItems.map((item) => (
                  <label
                    key={getItemKey ? getItemKey(item) : String(item)}
                    className="flex items-start gap-2 px-3 py-2 hover:bg-gray-50 cursor-pointer text-sm"
                  >
                    <input
                      type="checkbox"
                      checked={
                        selectedItems.length === 0 ||
                        selectedItems.includes(item)
                      }
                      onChange={() => onToggleItem(item)}
                      className="text-blue-600"
                    />
                    <span className="text-gray-700 break-words">
                      {getItemLabel(item)}
                    </span>
                  </label>
                ))
              )}
            </div>
          </>,
          document.body,
        )}
    </div>
  )
}
