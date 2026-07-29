export type TeiViewMode = 'original' | `translation:${number}`

export type TeiTranslation = {
  id: `translation:${number}`
  label: string
}

export type TeiHighlightCategory = {
  id: string
  label: string
}

export type TeiHighlightCategoryConfig = {
  label?: string
  color?: string
  description?: string
  renderMode?: 'fill' | 'outline'
}

export type TeiManualHighlight = {
  id: string
  paragraphIndex: number
  start: number
  end: number
  featureId: string
  surface?: string
  normalized?: string
  institution?: string
  ancientPersona?: string
}

export type TeiHighlightConfig = {
  selectedCategoryIds?: string[]
  categoryConfigById?: Record<string, TeiHighlightCategoryConfig>
  categoryToFeatureId?: Record<string, string>
  manualHighlights?: TeiManualHighlight[]
  ignoreTeiHighlights?: boolean
  hiddenTeiHighlightIds?: string[]
}

export type TeiParagraphSelection = {
  start: number
  end: number
  surface: string
}

export type TeiSurfaceZone = {
  id: string
  matchIds: string[]
  hoverMatchIds: string[]
  zoneType: 'line' | 'block'
  zoneCategory: string
  ulx: number
  uly: number
  lrx: number
  lry: number
  refUlx: number
  refUly: number
  refLrx: number
  refLry: number
  hasSurfaceBounds: boolean
}

export type TeiTranslationSource = {
  label: string
  element: Element
}

export type TeiHighlightSpan = {
  id: string
  fromAnchorId: string
  toAnchorId: string
  categoryId: string
  categoryLabel: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
}

export type ParagraphAnchorLocation = {
  paragraphIndex: number
  offset: number
}

export type ParagraphHighlightSpan = {
  id: string
  start: number
  end: number
  tooltipStart?: number
  tooltipEnd?: number
  featureId: string
  categoryId: string
  categoryLabel: string
  description: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
  fromAnchorId?: string
  toAnchorId?: string
  color: string
  renderMode: 'fill' | 'outline'
}

export type TeiEditableHighlight = {
  id: string
  paragraphIndex: number
  start: number
  end: number
  featureId: string
  categoryId: string
  categoryLabel: string
  surface: string
  normalized: string
  institution: string
  ancientPersona: string
  fromAnchorId: string
  toAnchorId: string
}

export type TeiOriginalEditableLine = {
  id: string
  blockId: string
  text: string
}

export type ReadingOptions = {
  showPB: boolean
  minCert: number
  maskChar: string
  alignLines: boolean
  showCertaintyVisualization?: boolean
  certaintyDegreeByTargetId?: Map<string, number>
}

export type TextWithAnchors = {
  text: string
  anchors: Record<string, number>
}

export type ParagraphLineRange = {
  start: number
  end: number
  matchIds: string[]
  certaintyDegree: number | null
}

export type ParagraphTextWithAnchors = {
  text: string
  anchors: Record<string, number>
  lineRanges: ParagraphLineRange[]
  blockType?: string
  table?: ParagraphTable
}

export type ParagraphTableCell = {
  start: number
  end: number
}

export type ParagraphTable = {
  rows: ParagraphTableCell[][]
  columnCount: number
}

export type LineMatchMode = 'none' | 'original-id' | 'corresp'

export type LineTextWithAnchors = TextWithAnchors & {
  matchIds: string[]
  certaintyDegree: number | null
}
