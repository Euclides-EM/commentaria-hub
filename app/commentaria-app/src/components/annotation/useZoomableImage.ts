import {
  type MouseEvent as ReactMouseEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import type { ImageDisplayBox } from './imageZoneUtils.ts'

export function useZoomableImage(imageUrl: string, zoom: number) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const imgRef = useRef<HTMLImageElement | null>(null)
  const [zoomedForUrl, setZoomedForUrl] = useState<string | null>(null)
  const [cursorLocal, setCursorLocal] = useState({ x: 0, y: 0 })
  const [containerSize, setContainerSize] = useState({ width: 0, height: 0 })
  const [naturalSizeByUrl, setNaturalSizeByUrl] = useState<{
    url: string
    width: number
    height: number
  } | null>(null)

  const isZoomed = zoomedForUrl === imageUrl
  const naturalSize = useMemo(
    () =>
      naturalSizeByUrl?.url === imageUrl
        ? { width: naturalSizeByUrl.width, height: naturalSizeByUrl.height }
        : { width: 0, height: 0 },
    [imageUrl, naturalSizeByUrl],
  )

  useEffect(() => {
    const container = containerRef.current
    if (!container) return
    const ro = new ResizeObserver(([entry]) => {
      setContainerSize({
        width: entry.contentRect.width,
        height: entry.contentRect.height,
      })
    })
    ro.observe(container)
    const onWindowResize = () =>
      setContainerSize({
        width: container.clientWidth,
        height: container.clientHeight,
      })
    window.addEventListener('resize', onWindowResize)
    return () => {
      ro.disconnect()
      window.removeEventListener('resize', onWindowResize)
    }
  }, [])

  useEffect(() => {
    const img = imgRef.current
    if (!img) return
    const handleLoad = () =>
      setNaturalSizeByUrl({
        url: imageUrl,
        width: img.naturalWidth,
        height: img.naturalHeight,
      })
    img.addEventListener('load', handleLoad)
    const raf = requestAnimationFrame(() => {
      if (img.complete && img.naturalWidth) {
        handleLoad()
      }
    })
    return () => {
      img.removeEventListener('load', handleLoad)
      cancelAnimationFrame(raf)
    }
  }, [imageUrl])

  const imageDisplayBox: ImageDisplayBox = useMemo(() => {
    const { width: cw, height: ch } = containerSize
    const { width: nw, height: nh } = naturalSize
    const empty: ImageDisplayBox = {
      left: 0,
      top: 0,
      width: 0,
      height: 0,
      naturalWidth: nw,
      naturalHeight: nh,
    }
    if (!cw || !ch || !nw || !nh) return empty
    const containerRatio = cw / ch
    const imageRatio = nw / nh
    let renderedWidth = cw
    let renderedHeight = ch
    let offsetLeft = 0
    let offsetTop = 0
    if (imageRatio > containerRatio) {
      renderedHeight = cw / imageRatio
      offsetTop = (ch - renderedHeight) / 2
    } else {
      renderedWidth = ch * imageRatio
      offsetLeft = (cw - renderedWidth) / 2
    }
    return {
      left: offsetLeft,
      top: offsetTop,
      width: renderedWidth,
      height: renderedHeight,
      naturalWidth: nw,
      naturalHeight: nh,
    }
  }, [containerSize, naturalSize])

  const s = zoom / 100
  const tx = isZoomed ? cursorLocal.x * (1 - s) : 0
  const ty = isZoomed ? cursorLocal.y * (1 - s) : 0
  const zoomTransform = isZoomed
    ? `translate(${tx}px, ${ty}px) scale(${s})`
    : 'none'

  const handleContainerClick = (event: ReactMouseEvent<HTMLDivElement>) => {
    if (!isZoomed) {
      const rect = event.currentTarget.getBoundingClientRect()
      setCursorLocal({
        x: event.clientX - rect.left,
        y: event.clientY - rect.top,
      })
      setZoomedForUrl(imageUrl)
    } else {
      setZoomedForUrl(null)
    }
  }

  const updateCursor = (cx: number, cy: number) => {
    setCursorLocal({ x: cx, y: cy })
  }

  const getLocalCursor = (cx: number, cy: number) => ({
    localX: isZoomed ? (cx - tx) / s : cx,
    localY: isZoomed ? (cy - ty) / s : cy,
  })

  return {
    containerRef,
    imgRef,
    isZoomed,
    zoomTransform,
    imageDisplayBox,
    handleContainerClick,
    updateCursor,
    getLocalCursor,
  }
}
