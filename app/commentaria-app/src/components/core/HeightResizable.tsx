import { type CSSProperties, type ReactNode, useEffect, useRef } from 'react'
import useLocalStorageState from 'use-local-storage-state'

interface HeightResizableProps {
  storageKey: string
  defaultHeight: number
  minHeight?: number
  maxHeight?: number
  className?: string
  style?: CSSProperties
  children: ReactNode
}

export function HeightResizable({
  storageKey,
  defaultHeight,
  minHeight = 200,
  maxHeight = 4000,
  className = '',
  style,
  children,
}: HeightResizableProps) {
  const [height, setHeight] = useLocalStorageState(storageKey, {
    defaultValue: defaultHeight,
    storageSync: false,
  })
  const dragRef = useRef<{ startY: number; startHeight: number } | null>(null)

  useEffect(() => {
    const onPointerMove = (event: PointerEvent) => {
      const drag = dragRef.current
      if (!drag) return
      const next = drag.startHeight + event.clientY - drag.startY
      setHeight(Math.round(Math.min(maxHeight, Math.max(minHeight, next))))
    }

    const onPointerUp = () => {
      dragRef.current = null
    }

    window.addEventListener('pointermove', onPointerMove)
    window.addEventListener('pointerup', onPointerUp)
    return () => {
      window.removeEventListener('pointermove', onPointerMove)
      window.removeEventListener('pointerup', onPointerUp)
    }
  }, [maxHeight, minHeight, setHeight])

  return (
    <div
      className={`relative flex flex-col pb-2 ${className}`}
      style={{ ...style, height: `${height}px` }}
    >
      {children}
      <div
        role="separator"
        aria-orientation="horizontal"
        aria-label="Resize height"
        className="absolute bottom-0 left-0 w-full h-2 cursor-row-resize flex items-center justify-center hover:bg-gray-200 transition-colors rounded"
        onPointerDown={(event) => {
          event.preventDefault()
          dragRef.current = { startY: event.clientY, startHeight: height }
        }}
      >
        <div className="w-10 h-0.5 rounded-full bg-gray-300" />
      </div>
    </div>
  )
}
