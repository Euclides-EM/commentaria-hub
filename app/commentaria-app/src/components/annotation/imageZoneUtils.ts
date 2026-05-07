import type { TeiSurfaceZone } from './contents/tei/tei.ts'
import {
  filterSurfaceZones,
  type HighlightZoneFilter,
} from './highlightControls.ts'

export type ImageDisplayBox = {
  left: number
  top: number
  width: number
  height: number
  naturalWidth: number
  naturalHeight: number
}

export type VisibleZone = TeiSurfaceZone & {
  left: number
  top: number
  width: number
  height: number
  isActive: boolean
}

export function filterValidZones(
  surfaceZones: TeiSurfaceZone[],
  filters: HighlightZoneFilter[],
): TeiSurfaceZone[] {
  return filterSurfaceZones(surfaceZones, filters).filter(
    (zone) =>
      Number.isFinite(zone.ulx) &&
      Number.isFinite(zone.uly) &&
      Number.isFinite(zone.lrx) &&
      Number.isFinite(zone.lry) &&
      Number.isFinite(zone.refUlx) &&
      Number.isFinite(zone.refUly) &&
      Number.isFinite(zone.refLrx) &&
      Number.isFinite(zone.refLry) &&
      zone.lrx > zone.ulx &&
      zone.lry > zone.uly &&
      zone.refLrx > zone.refUlx &&
      zone.refLry > zone.refUly,
  )
}

export function computeVisibleZones(
  highlightableZones: TeiSurfaceZone[],
  imageDisplayBox: ImageDisplayBox,
  activeMatchIdSet: Set<string>,
): VisibleZone[] {
  if (!imageDisplayBox.width || !imageDisplayBox.height) return []
  return highlightableZones
    .map((zone) => {
      const useNaturalBounds =
        !zone.hasSurfaceBounds &&
        imageDisplayBox.naturalWidth > 0 &&
        imageDisplayBox.naturalHeight > 0
      const refUlx = useNaturalBounds ? 0 : zone.refUlx
      const refUly = useNaturalBounds ? 0 : zone.refUly
      const refLrx = useNaturalBounds
        ? imageDisplayBox.naturalWidth
        : zone.refLrx
      const refLry = useNaturalBounds
        ? imageDisplayBox.naturalHeight
        : zone.refLry
      const refWidth = refLrx - refUlx
      const refHeight = refLry - refUly
      const left =
        imageDisplayBox.left +
        ((zone.ulx - refUlx) / refWidth) * imageDisplayBox.width
      const top =
        imageDisplayBox.top +
        ((zone.uly - refUly) / refHeight) * imageDisplayBox.height
      const width = ((zone.lrx - zone.ulx) / refWidth) * imageDisplayBox.width
      const height =
        ((zone.lry - zone.uly) / refHeight) * imageDisplayBox.height
      const isActive = zone.matchIds.some((id) => activeMatchIdSet.has(id))
      return { ...zone, left, top, width, height, isActive }
    })
    .filter(
      (z) =>
        z.width > 0 &&
        z.height > 0 &&
        z.left < imageDisplayBox.left + imageDisplayBox.width &&
        z.top < imageDisplayBox.top + imageDisplayBox.height &&
        z.left + z.width > imageDisplayBox.left &&
        z.top + z.height > imageDisplayBox.top,
    )
}
