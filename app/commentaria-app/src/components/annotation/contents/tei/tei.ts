export type {
  TeiEditableHighlight,
  TeiHighlightCategory,
  TeiHighlightCategoryConfig,
  TeiHighlightConfig,
  TeiManualHighlight,
  TeiOriginalEditableLine,
  TeiParagraphSelection,
  TeiSurfaceZone,
  TeiTranslation,
  TeiViewMode,
} from './teiTypes.ts'
export { getTeiParagraphSelection } from './teiSelection.ts'
export {
  getTeiHighlightCategories,
  getTeiZoneToServerTextBlockId,
  toServerTextBlockId,
} from './teiHighlights.ts'
export {
  getTeiEditableHighlights,
  getTeiOriginalEditableLines,
  getTeiTranslations,
  hasTeiCertaintyDegrees,
} from './teiExtractors.ts'
export { getTeiSurfaceZones } from './teiSurfaceZones.ts'

import { parseXml } from './teiDom.ts'
import { renderTeiHtml } from './teiRendering.ts'
import type { TeiHighlightConfig, TeiViewMode } from './teiTypes.ts'

export const parseTeisXml = (xml: string | undefined): Map<string, string> => {
  const result = new Map<string, string>()
  if (!xml?.trim()) return result
  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(xml.trim(), 'text/xml')
    const root = doc.documentElement
    if (!root) return result
    const serializer = new XMLSerializer()
    for (const child of Array.from(root.children)) {
      const key =
        child.getAttribute('key') ||
        child.getAttribute('n') ||
        child.getAttribute('pageNumOrKey') ||
        child.getAttribute('page')
      if (!key) continue
      const teiEl = child.querySelector('TEI') || child
      result.set(key, serializer.serializeToString(teiEl))
    }
  } catch {
    // return empty map
  }
  return result
}

export const teiToHtml = (
  tei: string,
  minCert: number,
  searchResultHighlight: string | null,
  maskChar: string = '@',
  viewMode: TeiViewMode = 'original',
  alignLines: boolean = false,
  showCertaintyVisualization: boolean = false,
  highlightConfig?: TeiHighlightConfig,
) =>
  renderTeiHtml(
    parseXml(tei.trim()),
    minCert,
    searchResultHighlight,
    maskChar,
    viewMode,
    alignLines,
    showCertaintyVisualization,
    highlightConfig,
  )
