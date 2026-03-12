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
