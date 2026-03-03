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

export type TeiSurfaceZone = {
  id: string
  matchIds: string[]
  hoverMatchIds: string[]
  zoneType: 'line' | 'block'
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

type TeiTranslationSource = {
  label: string
  element: Element
}

type TeiHighlightSpan = {
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

type ParagraphAnchorLocation = {
  paragraphIndex: number
  offset: number
}

type ParagraphHighlightSpan = {
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

type ReadingOptions = {
  showPB: boolean
  minCert: number
  maskChar: string
  alignLines: boolean
}

type TextWithAnchors = {
  text: string
  anchors: Record<string, number>
}

type ParagraphTextWithAnchors = {
  text: string
  anchors: Record<string, number>
  lineRanges: ParagraphLineRange[]
}

type ParagraphLineRange = {
  start: number
  end: number
  matchIds: string[]
}

type LineMatchMode = 'none' | 'original-id' | 'corresp'

type LineTextWithAnchors = TextWithAnchors & {
  matchIds: string[]
}

const escapeHtml = (text: string) => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

const escapeHtmlAttr = (text: string) =>
  escapeHtml(text).replaceAll('"', '&quot;').replaceAll("'", '&#39;')

const parseXml = (xmlString: string) => {
  const parser = new DOMParser()
  const doc = parser.parseFromString(xmlString, 'application/xml')
  const pe = doc.getElementsByTagName('parsererror')[0]
  if (pe) {
    throw new Error('XML parse error (often unescaped & or mismatched tags).')
  }
  return doc
}

const isElement = (node: ChildNode, localName: string): node is Element =>
  node &&
  node.nodeType === Node.ELEMENT_NODE &&
  (node as Element).localName === localName

function findFirstCertaintyDegree(segEl: Element) {
  const certs = segEl.getElementsByTagNameNS('*', 'certainty')
  if (!certs || !certs.length) return null

  const degreeStr = certs[0].getAttribute('degree')
  const degree = degreeStr == null ? NaN : parseFloat(degreeStr)
  return Number.isFinite(degree) ? degree : null
}

function textContentExcludingCertainty(node: ChildNode) {
  let out = ''
  for (let i = 0; i < node.childNodes.length; i++) {
    const child = node.childNodes[i]
    if (child.nodeType === Node.TEXT_NODE) {
      out += child.nodeValue || ''
      continue
    }
    if (child.nodeType !== Node.ELEMENT_NODE) continue

    if (isElement(child, 'certainty')) continue

    out += textContentExcludingCertainty(child)
  }
  return out
}

function maskText(s: string, maskChar: string) {
  const m = maskChar && String(maskChar).length ? String(maskChar)[0] : '@'
  let out = ''
  for (const ch of s) {
    out += /\s/.test(ch) ? ch : m
  }
  return out
}

const getElementAttr = (el: Element, name: string) =>
  (el.getAttribute(name) || '').trim()

const getXmlId = (el: Element) =>
  getElementAttr(el, 'xml:id') || getElementAttr(el, 'id')

const parseAnaRefs = (value: string | null) =>
  (value || '')
    .split(/\s+/)
    .map((entry) => entry.replace(/^#/, '').trim())
    .filter(Boolean)

const parseCorrespRefs = (value: string | null) =>
  (value || '')
    .split(/\s+/)
    .map((entry) => entry.replace(/^#/, '').trim())
    .filter(Boolean)

export const toServerTextBlockId = (value: string | null | undefined) => {
  const normalized = (value || '').replace(/^#/, '').trim()
  if (!normalized) {
    return ''
  }
  return normalized.replace(/^alto:textblock:/i, '')
}

export const getTeiZoneToServerTextBlockId = (tei: string) => {
  const out: Record<string, string> = {}
  try {
    const doc = parseXml(tei.trim())
    const zones = doc.getElementsByTagNameNS('*', 'zone')
    for (let index = 0; index < zones.length; index++) {
      const zone = zones[index]
      const zoneId = getXmlId(zone)
      if (!zoneId) {
        continue
      }
      const correspRefs = parseCorrespRefs(zone.getAttribute('corresp'))
      if (!correspRefs.length) {
        continue
      }
      const textBlockRef =
        correspRefs.find((entry) => /^alto:textblock:/i.test(entry)) ||
        correspRefs[0]
      const serverId = toServerTextBlockId(textBlockRef)
      if (!serverId) {
        continue
      }
      out[zoneId] = serverId
    }
    return out
  } catch {
    return out
  }
}

const toUniqueSorted = (values: Array<string | null | undefined>) =>
  [
    ...new Set(values.map((value) => (value || '').trim()).filter(Boolean)),
  ].sort()

const isVerbCategory = (categoryId: string, categoryLabel: string) => {
  const normalized = `${categoryId} ${categoryLabel}`
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '')
  return normalized.includes('verb')
}

const getHighlightCategoryMap = (doc: Document) => {
  const categories = new Map<string, string>()
  const interpGroups = doc.getElementsByTagNameNS('*', 'interpGrp')
  for (let i = 0; i < interpGroups.length; i++) {
    const group = interpGroups[i]
    if (getElementAttr(group, 'type') !== 'highlight-categories') {
      continue
    }
    const interps = group.getElementsByTagNameNS('*', 'interp')
    for (let j = 0; j < interps.length; j++) {
      const interp = interps[j]
      const id = getXmlId(interp)
      if (!id) continue
      const label = (interp.textContent || '').trim() || id
      categories.set(id, label)
    }
  }
  return categories
}

const getTeiHighlightSpans = (doc: Document): TeiHighlightSpan[] => {
  const categoryMap = getHighlightCategoryMap(doc)
  const spans: TeiHighlightSpan[] = []
  const spanGroups = doc.getElementsByTagNameNS('*', 'spanGrp')

  for (let i = 0; i < spanGroups.length; i++) {
    const group = spanGroups[i]
    const type = getElementAttr(group, 'type')
    if (type !== 'highlight' && type !== 'highlights') {
      continue
    }

    const groupSpans = group.getElementsByTagNameNS('*', 'span')
    for (let j = 0; j < groupSpans.length; j++) {
      const span = groupSpans[j]
      const fromAnchorId = getElementAttr(span, 'from').replace(/^#/, '')
      const toAnchorId = getElementAttr(span, 'to').replace(/^#/, '')
      if (!fromAnchorId || !toAnchorId) continue

      const anaRefs = parseAnaRefs(span.getAttribute('ana'))
      const categoryId =
        anaRefs.find((ref) => ref.startsWith('cat_')) ||
        anaRefs[0] ||
        'uncategorized'
      const categoryLabel = categoryMap.get(categoryId) || categoryId
      const id = getXmlId(span) || `${categoryId}:${fromAnchorId}-${toAnchorId}`
      const notes = span.getElementsByTagNameNS('*', 'note')
      let normalized = ''
      let surface = ''
      let institution = ''
      let ancientPersona = ''
      for (let k = 0; k < notes.length; k++) {
        const noteAnaRefs = parseAnaRefs(notes[k].getAttribute('ana'))
        const text = (notes[k].textContent || '').trim()
        const isSurface =
          noteAnaRefs.some((ref) => ref === 'prop_surface') ||
          noteAnaRefs.some((ref) => ref.endsWith('surface'))
        if (isSurface && text) {
          surface = text
        }
        const isInstitution =
          noteAnaRefs.some((ref) => ref === 'prop_institution') ||
          noteAnaRefs.some((ref) => ref.endsWith('institution'))
        if (isInstitution && text) {
          institution = text
        }
        const isAncientPersona =
          noteAnaRefs.some((ref) => ref === 'prop_ancient_persona') ||
          noteAnaRefs.some((ref) => ref.endsWith('ancient_persona'))
        if (isAncientPersona && text) {
          ancientPersona = text
        }
        const isNormalized =
          noteAnaRefs.some((ref) => ref === 'prop_normalized') ||
          noteAnaRefs.some((ref) => ref.endsWith('normalized'))
        if (isNormalized && text && !normalized) {
          normalized = text
        }
      }

      spans.push({
        id,
        fromAnchorId,
        toAnchorId,
        categoryId,
        categoryLabel,
        surface,
        normalized,
        institution,
        ancientPersona,
      })
    }
  }

  return spans
}

export const getTeiHighlightCategories = (
  tei: string,
): TeiHighlightCategory[] => {
  try {
    const doc = parseXml(tei.trim())
    const byId = new Map<string, string>()
    const spans = getTeiHighlightSpans(doc)
    for (const span of spans) {
      if (!byId.has(span.categoryId)) {
        byId.set(span.categoryId, span.categoryLabel)
      }
    }
    return [...byId.entries()].map(([id, label]) => ({ id, label }))
  } catch {
    return []
  }
}

const appendTextWithAnchors = (
  node: ChildNode,
  opts: ReadingOptions,
  builder: TextWithAnchors,
) => {
  if (node.nodeType === Node.TEXT_NODE) {
    const raw = node.nodeValue || ''
    builder.text += /[\n\r\t]/.test(raw) ? raw.replace(/\s+/g, ' ') : raw
    return
  }
  if (node.nodeType !== Node.ELEMENT_NODE) {
    return
  }

  const element = node as Element

  if (element.localName === 'anchor') {
    const anchorId = getXmlId(element)
    if (anchorId) {
      builder.anchors[anchorId] = builder.text.length
    }
    return
  }

  if (element.localName === 'lb') {
    builder.text += opts.alignLines ? ' ' : '\n'
    return
  }

  if (element.localName === 'pb') {
    if (opts.showPB) {
      const facs = element.getAttribute('facs') || ''
      const n = element.getAttribute('n') || ''
      builder.text += n || facs ? `Page break ${n || facs}` : 'Page break'
    }
    return
  }

  if (element.localName === 'seg') {
    const degree = findFirstCertaintyDegree(element)
    const rawText = textContentExcludingCertainty(element)
    const minCert = Number.isFinite(opts.minCert) ? opts.minCert : 0
    builder.text +=
      degree != null && degree < minCert
        ? maskText(rawText, opts.maskChar)
        : rawText
    return
  }

  for (let i = 0; i < element.childNodes.length; i++) {
    appendTextWithAnchors(element.childNodes[i], opts, builder)
  }

  if (element.localName === 'p') {
    builder.text += '\n'
  }
}

const toReadingTextWithAnchors = (node: Element, opts: ReadingOptions) => {
  const builder: TextWithAnchors = { text: '', anchors: {} }
  for (let i = 0; i < node.childNodes.length; i++) {
    appendTextWithAnchors(node.childNodes[i], opts, builder)
  }
  return builder
}

const trimTextWithAnchors = (value: TextWithAnchors): TextWithAnchors => {
  const leading = value.text.length - value.text.trimStart().length
  const trailing = value.text.length - value.text.trimEnd().length
  const end = value.text.length - trailing
  const trimmedText = value.text.slice(leading, end)
  const anchors: Record<string, number> = {}

  for (const [id, offset] of Object.entries(value.anchors)) {
    anchors[id] = Math.max(0, Math.min(trimmedText.length, offset - leading))
  }

  return { text: trimmedText, anchors }
}

const getDirectChildrenByName = (parent: Element, localName: string) => {
  const children: Element[] = []
  for (let i = 0; i < parent.children.length; i++) {
    const child = parent.children[i]
    if (child.localName === localName) {
      children.push(child)
    }
  }
  return children
}

const getLineElements = (container: Element): Element[] => {
  const directLines = getDirectChildrenByName(container, 'l')
  if (directLines.length) return directLines
  const lines = container.getElementsByTagNameNS('*', 'l')
  const out: Element[] = []
  for (let i = 0; i < lines.length; i++) {
    out.push(lines[i])
  }
  return out
}

const getMatchIdsForElement = (
  element: Element,
  lineMatchMode: LineMatchMode,
) => {
  if (lineMatchMode === 'original-id') {
    return toUniqueSorted([
      getXmlId(element),
      ...parseCorrespRefs(element.getAttribute('corresp')),
      ...parseCorrespRefs(element.getAttribute('facs')),
    ])
  }
  if (lineMatchMode === 'corresp') {
    return toUniqueSorted([
      ...parseCorrespRefs(element.getAttribute('corresp')),
      ...parseCorrespRefs(element.getAttribute('facs')),
    ])
  }
  return []
}

const joinLineTexts = (
  lines: LineTextWithAnchors[],
  alignLines: boolean,
): ParagraphTextWithAnchors[] => {
  if (!alignLines) {
    let text = ''
    const anchors: Record<string, number> = {}
    const lineRanges: ParagraphLineRange[] = []

    for (let i = 0; i < lines.length; i++) {
      if (i > 0) {
        text += '\n'
      }
      const offset = text.length
      text += lines[i].text
      const end = text.length
      if (lines[i].matchIds.length > 0 && end > offset) {
        lineRanges.push({
          start: offset,
          end,
          matchIds: lines[i].matchIds,
        })
      }
      for (const [id, pos] of Object.entries(lines[i].anchors)) {
        anchors[id] = offset + pos
      }
    }

    return [{ text, anchors, lineRanges }]
  }

  const paragraphs: ParagraphTextWithAnchors[] = []
  let currentText = ''
  let currentAnchors: Record<string, number> = {}
  let currentLineRanges: ParagraphLineRange[] = []
  let previousLineEndedWithMergeDash = false
  const trailingMergeDashPattern =
    /[-\u00AD\u2010\u2011\u2012\u2013\u2014\u2015\u2212\uFE63\uFF0D¬]+$/

  const truncateTrailingMergeDashes = (
    line: LineTextWithAnchors,
  ): LineTextWithAnchors => {
    if (!trailingMergeDashPattern.test(line.text)) {
      return line
    }
    const withoutDashes = line.text.replace(trailingMergeDashPattern, '')
    const nextText = withoutDashes.replace(/\s+$/, '')
    const nextAnchors: Record<string, number> = {}
    for (const [id, pos] of Object.entries(line.anchors)) {
      nextAnchors[id] = Math.min(pos, nextText.length)
    }
    return {
      ...line,
      text: nextText,
      anchors: nextAnchors,
    }
  }

  const pushCurrent = () => {
    if (!currentText) {
      return
    }
    paragraphs.push({
      text: currentText,
      anchors: currentAnchors,
      lineRanges: currentLineRanges,
    })
    currentText = ''
    currentAnchors = {}
    currentLineRanges = []
    previousLineEndedWithMergeDash = false
  }

  for (const rawLine of lines) {
    const lineEndedWithMergeDash = trailingMergeDashPattern.test(rawLine.text)
    const line = truncateTrailingMergeDashes(rawLine)
    if (!line.text) {
      pushCurrent()
      previousLineEndedWithMergeDash = false
      continue
    }

    if (currentText && !previousLineEndedWithMergeDash) {
      currentText += ' '
    }
    const offset = currentText.length
    currentText += line.text
    const end = currentText.length
    if (line.matchIds.length > 0 && end > offset) {
      currentLineRanges.push({
        start: offset,
        end,
        matchIds: line.matchIds,
      })
    }
    for (const [id, pos] of Object.entries(line.anchors)) {
      currentAnchors[id] = offset + pos
    }
    previousLineEndedWithMergeDash = lineEndedWithMergeDash
  }

  pushCurrent()
  return paragraphs.length
    ? paragraphs
    : [{ text: '', anchors: {}, lineRanges: [] }]
}

const renderStructuredDivToParagraphs = (
  div: Element,
  opts: ReadingOptions,
  lineMatchMode: LineMatchMode = 'none',
): ParagraphTextWithAnchors[] => {
  const blocks = getDirectChildrenByName(div, 'ab')
  const containers = blocks.length ? blocks : [div]
  const paragraphs: ParagraphTextWithAnchors[] = []

  for (const container of containers) {
    const lines = getLineElements(container)
    if (lines.length) {
      const renderedLines = lines.map(
        (line) =>
          ({
            ...trimTextWithAnchors(toReadingTextWithAnchors(line, opts)),
            matchIds: getMatchIdsForElement(line, lineMatchMode),
          }) satisfies LineTextWithAnchors,
      )
      const blocks = joinLineTexts(renderedLines, opts.alignLines)
      for (const block of blocks) {
        paragraphs.push(block)
      }
      continue
    }

    const block = trimTextWithAnchors(toReadingTextWithAnchors(container, opts))
    const blockMatchIds = getMatchIdsForElement(container, lineMatchMode)
    paragraphs.push({
      ...block,
      lineRanges:
        blockMatchIds.length > 0 && block.text.length > 0
          ? [{ start: 0, end: block.text.length, matchIds: blockMatchIds }]
          : [],
    })
  }

  return paragraphs
}

const getParagraphHighlightSlice = (
  spans: ParagraphHighlightSpan[],
  sliceStart: number,
  sliceEnd: number,
): ParagraphHighlightSpan[] => {
  const sliced: ParagraphHighlightSpan[] = []
  for (const span of spans) {
    const start = Math.max(sliceStart, span.start)
    const end = Math.min(sliceEnd, span.end)
    if (end <= start) {
      continue
    }
    sliced.push({
      ...span,
      start: start - sliceStart,
      end: end - sliceStart,
      tooltipStart: span.tooltipStart ?? span.start,
      tooltipEnd: span.tooltipEnd ?? span.end,
    })
  }
  return sliced
}

const renderParagraphWithLineRanges = (
  text: string,
  spans: ParagraphHighlightSpan[],
  paragraphIndex: number,
  lineRanges: ParagraphLineRange[],
) => {
  const validRanges = lineRanges
    .map((range) => ({
      start: Math.max(0, Math.min(text.length, range.start)),
      end: Math.max(0, Math.min(text.length, range.end)),
      matchIds: [...new Set(range.matchIds.filter(Boolean))],
    }))
    .filter((range) => range.end > range.start && range.matchIds.length > 0)
    .sort((left, right) => left.start - right.start)

  if (!validRanges.length) {
    return renderParagraphWithHighlights(text, spans, paragraphIndex)
  }

  let html = ''
  let cursor = 0
  for (const range of validRanges) {
    const rangeStart = Math.max(cursor, range.start)
    const rangeEnd = Math.max(rangeStart, range.end)
    if (rangeStart > cursor) {
      const gapSpans = getParagraphHighlightSlice(spans, cursor, rangeStart)
      html += renderParagraphWithHighlights(
        text.slice(cursor, rangeStart),
        gapSpans,
        paragraphIndex,
      )
    }

    if (rangeEnd <= rangeStart) {
      continue
    }
    const lineSpans = getParagraphHighlightSlice(spans, rangeStart, rangeEnd)
    const lineHtml = renderParagraphWithHighlights(
      text.slice(rangeStart, rangeEnd),
      lineSpans,
      paragraphIndex,
    )
    html += `<span data-tei-line-match-ids="${escapeHtmlAttr(range.matchIds.join(' '))}">${lineHtml}</span>`
    cursor = rangeEnd
  }

  if (cursor < text.length) {
    const tailSpans = getParagraphHighlightSlice(spans, cursor, text.length)
    html += renderParagraphWithHighlights(
      text.slice(cursor),
      tailSpans,
      paragraphIndex,
    )
  }

  return html || '&nbsp;'
}

const renderStructuredDiv = (
  div: Element,
  opts: ReadingOptions,
  lineMatchMode: LineMatchMode = 'none',
) => {
  const paragraphs = renderStructuredDivToParagraphs(div, opts, lineMatchMode)
  return paragraphs
    .map((paragraph, index) => {
      const html = renderParagraphWithLineRanges(
        paragraph.text,
        [],
        index,
        paragraph.lineRanges,
      )
      return `<p>${html}</p>`
    })
    .join('')
}

const getXmlLang = (el: Element) =>
  getElementAttr(el, 'xml:lang') || getElementAttr(el, 'lang')

const getTeiTranslationSources = (body: Element): TeiTranslationSource[] => {
  const directDivTranslations: TeiTranslationSource[] = []
  const directDivs = getDirectChildrenByName(body, 'div')
  for (const div of directDivs) {
    if (getElementAttr(div, 'type') !== 'translation') continue
    const lang = getXmlLang(div)
    const n = getElementAttr(div, 'n')
    directDivTranslations.push({
      element: div,
      label: lang || n,
    })
  }
  return directDivTranslations
}

export function getTeiTranslations(tei: string): TeiTranslation[] {
  try {
    const doc = parseXml(tei.trim())
    const body = getBody(doc)
    const sources = getTeiTranslationSources(body)
    return sources.map((source, index) => ({
      id: `translation:${index}`,
      label: source.label || `Translation ${index + 1}`,
    }))
  } catch {
    return []
  }
}

export function hasTeiCertaintyDegrees(tei: string): boolean {
  try {
    const doc = parseXml(tei.trim())
    const certs = doc.getElementsByTagNameNS('*', 'certainty')
    for (let i = 0; i < certs.length; i++) {
      const degree = Number.parseFloat(certs[i].getAttribute('degree') || '')
      if (Number.isFinite(degree)) {
        return true
      }
    }
    return false
  } catch {
    return false
  }
}

function renderTranslationView(
  body: Element,
  translationIndex: number,
  opts: ReadingOptions,
): string {
  const sources = getTeiTranslationSources(body)
  const source = sources[translationIndex]
  if (!source) return '<p></p>'
  return renderStructuredDiv(source.element, opts, 'corresp') || '<p></p>'
}

const getSegmentStyle = (spans: ParagraphHighlightSpan[]) => {
  const fillSpans = spans.filter((span) => span.renderMode !== 'outline')
  const outlineSpans = spans.filter((span) => span.renderMode === 'outline')

  let base = ''
  if (fillSpans.length > 0) {
    const sortedByLengthDesc = [...fillSpans].sort(
      (left, right) => right.end - right.start - (left.end - left.start),
    )
    const layers = sortedByLengthDesc.map((span, index) => {
      const depth = sortedByLengthDesc.length - 1 - index
      const shadowSize = Math.min(6, 2 + depth * 2)
      return { color: span.color, shadowSize }
    })

    // CSS paints the first shadow on top; place narrower spans first.
    const shadows = [...layers]
      .reverse()
      .map((layer) => `0 0 0 ${layer.shadowSize}px ${layer.color}`)
      .join(', ')
    const topColor = layers[layers.length - 1].color
    base = `background-color: ${topColor}; box-shadow: ${shadows}; border-radius: 3px;`
  }

  if (!outlineSpans.length) {
    return base || 'border-radius: 3px;'
  }

  const outlineColor = outlineSpans[0].color
  return `${base} outline: 2px solid ${outlineColor}; outline-offset: 1px; border-radius: 3px;`
}

const renderParagraphWithHighlights = (
  text: string,
  spans: ParagraphHighlightSpan[],
  paragraphIndex: number,
) => {
  if (!text) {
    return '&nbsp;'
  }

  const clampedSpans = spans
    .map((span) => ({
      ...span,
      start: Math.max(0, Math.min(text.length, span.start)),
      end: Math.max(0, Math.min(text.length, span.end)),
    }))
    .filter((span) => span.end > span.start)

  if (!clampedSpans.length) {
    return escapeHtml(text).replaceAll('\n', '<br>')
  }

  const boundaries = new Set<number>([0, text.length])
  for (const span of clampedSpans) {
    boundaries.add(span.start)
    boundaries.add(span.end)
  }
  const orderedBoundaries = [...boundaries].sort((a, b) => a - b)

  let html = ''
  for (let i = 0; i < orderedBoundaries.length - 1; i++) {
    const start = orderedBoundaries[i]
    const end = orderedBoundaries[i + 1]
    if (end <= start) continue

    const segmentText = text.slice(start, end)
    if (!segmentText) continue

    const activeSpans = clampedSpans.filter(
      (span) => span.start < end && span.end > start,
    )

    const escapedSegment = escapeHtml(segmentText).replaceAll('\n', '<br>')
    if (!activeSpans.length) {
      html += escapedSegment
      continue
    }

    const leadingWhitespace = segmentText.match(/^\s+/)?.[0] || ''
    const trailingWhitespace = segmentText.match(/\s+$/)?.[0] || ''
    const trimmedStart = leadingWhitespace.length
    const trimmedEnd = segmentText.length - trailingWhitespace.length
    const highlightedText =
      trimmedEnd > trimmedStart
        ? segmentText.slice(trimmedStart, trimmedEnd)
        : ''
    const escapedLeadingWhitespace = escapeHtml(leadingWhitespace).replaceAll(
      '\n',
      '<br>',
    )
    const escapedTrailingWhitespace = escapeHtml(trailingWhitespace).replaceAll(
      '\n',
      '<br>',
    )
    const escapedHighlightedText = escapeHtml(highlightedText).replaceAll(
      '\n',
      '<br>',
    )
    if (!highlightedText) {
      html += escapedLeadingWhitespace + escapedTrailingWhitespace
      continue
    }

    const style = getSegmentStyle(activeSpans)
    const tooltipItems = activeSpans.map((span) => ({
      id: span.id,
      featureId: span.featureId,
      categoryId: span.categoryId,
      label: span.categoryLabel,
      description: span.description,
      surface: span.surface,
      normalized: span.normalized,
      institution: span.institution,
      ancientPersona: span.ancientPersona,
      paragraphIndex,
      start: span.tooltipStart ?? span.start,
      end: span.tooltipEnd ?? span.end,
      fromAnchorId: span.fromAnchorId || '',
      toAnchorId: span.toAnchorId || '',
      color: span.color,
    }))
    const tooltipItemsAttr = escapeHtmlAttr(
      encodeURIComponent(JSON.stringify(tooltipItems)),
    )
    html += escapedLeadingWhitespace
    html += `<span data-tei-highlight="true" data-tei-highlight-tooltip="${tooltipItemsAttr}" style="${style}">${escapedHighlightedText}</span>`
    html += escapedTrailingWhitespace
  }

  return html || '&nbsp;'
}

const getOriginalParagraphSpans = (
  doc: Document,
  body: Element,
  opts: ReadingOptions,
  highlightConfig?: TeiHighlightConfig,
) => {
  const directDivs = getDirectChildrenByName(body, 'div')
  const transcriptionDivs = directDivs.filter(
    (div) => getElementAttr(div, 'type') === 'transcription',
  )

  if (!transcriptionDivs.length) {
    return {
      paragraphs: [] as ParagraphTextWithAnchors[],
      paragraphSpans: new Map<number, ParagraphHighlightSpan[]>(),
    }
  }

  const paragraphs: ParagraphTextWithAnchors[] = []
  for (let i = 0; i < body.children.length; i++) {
    const child = body.children[i]
    if (
      child.localName === 'div' &&
      getElementAttr(child, 'type') === 'transcription'
    ) {
      const rendered = renderStructuredDivToParagraphs(
        child,
        opts,
        'original-id',
      )
      for (const paragraph of rendered) {
        paragraphs.push(paragraph)
      }
    }
  }

  const anchorLocations: Record<string, ParagraphAnchorLocation> = {}
  for (let i = 0; i < paragraphs.length; i++) {
    for (const [anchorId, offset] of Object.entries(paragraphs[i].anchors)) {
      anchorLocations[anchorId] = { paragraphIndex: i, offset }
    }
  }

  const paragraphSpans = new Map<number, ParagraphHighlightSpan[]>()
  const highlights = highlightConfig?.ignoreTeiHighlights
    ? []
    : getTeiHighlightSpans(doc)
  const selectedSet =
    highlightConfig?.selectedCategoryIds != null
      ? new Set(highlightConfig.selectedCategoryIds)
      : null
  const categoryConfigById = highlightConfig?.categoryConfigById || {}
  const categoryToFeatureId = highlightConfig?.categoryToFeatureId || {}
  const hiddenTeiHighlightIds = new Set(
    highlightConfig?.hiddenTeiHighlightIds || [],
  )
  for (const highlight of highlights) {
    if (hiddenTeiHighlightIds.has(highlight.id)) {
      continue
    }
    const featureId =
      categoryToFeatureId[highlight.categoryId] || highlight.categoryId
    if (selectedSet && !selectedSet.has(featureId)) {
      continue
    }
    const from = anchorLocations[highlight.fromAnchorId]
    const to = anchorLocations[highlight.toAnchorId]
    if (!from || !to || from.paragraphIndex !== to.paragraphIndex) {
      continue
    }
    const start = Math.min(from.offset, to.offset)
    const end = Math.max(from.offset, to.offset)
    if (end <= start) {
      continue
    }

    const current = paragraphSpans.get(from.paragraphIndex) || []
    const categoryConfig =
      categoryConfigById[featureId] || categoryConfigById[highlight.categoryId]
    current.push({
      id: highlight.id,
      start,
      end,
      featureId,
      categoryId: highlight.categoryId,
      categoryLabel: categoryConfig?.label || highlight.categoryLabel,
      description: categoryConfig?.description || '',
      surface: highlight.surface,
      normalized: highlight.normalized,
      institution: highlight.institution,
      ancientPersona: highlight.ancientPersona,
      fromAnchorId: highlight.fromAnchorId,
      toAnchorId: highlight.toAnchorId,
      color: categoryConfig?.color || '#f2f2f2',
      renderMode:
        categoryConfig?.renderMode ||
        (isVerbCategory(highlight.categoryId, highlight.categoryLabel)
          ? 'outline'
          : 'fill'),
    })
    paragraphSpans.set(from.paragraphIndex, current)
  }

  const manualHighlights = highlightConfig?.manualHighlights || []
  for (const highlight of manualHighlights) {
    if (
      highlight.paragraphIndex < 0 ||
      highlight.paragraphIndex >= paragraphs.length
    ) {
      continue
    }
    if (selectedSet && !selectedSet.has(highlight.featureId)) {
      continue
    }
    const paragraph = paragraphs[highlight.paragraphIndex]
    if (!paragraph) {
      continue
    }
    const start = Math.max(0, Math.min(paragraph.text.length, highlight.start))
    const end = Math.max(0, Math.min(paragraph.text.length, highlight.end))
    if (end <= start) {
      continue
    }
    const categoryConfig = categoryConfigById[highlight.featureId]
    const current = paragraphSpans.get(highlight.paragraphIndex) || []
    current.push({
      id: highlight.id,
      start,
      end,
      featureId: highlight.featureId,
      categoryId: highlight.featureId,
      categoryLabel: categoryConfig?.label || highlight.featureId,
      description: categoryConfig?.description || '',
      surface: highlight.surface || paragraph.text.slice(start, end),
      normalized: highlight.normalized || '',
      institution: highlight.institution || '',
      ancientPersona: highlight.ancientPersona || '',
      fromAnchorId: '',
      toAnchorId: '',
      color: categoryConfig?.color || '#f2f2f2',
      renderMode: categoryConfig?.renderMode || 'fill',
    })
    paragraphSpans.set(highlight.paragraphIndex, current)
  }

  return { paragraphs, paragraphSpans }
}

function renderOriginalView(
  doc: Document,
  body: Element,
  opts: ReadingOptions,
  highlightConfig?: TeiHighlightConfig,
) {
  const { paragraphs, paragraphSpans } = getOriginalParagraphSpans(
    doc,
    body,
    opts,
    highlightConfig,
  )
  const parts = paragraphs.map((paragraph, index) => {
    const spans = paragraphSpans.get(index) || []
    return `<p data-tei-paragraph-index="${index}">${renderParagraphWithLineRanges(paragraph.text, spans, index, paragraph.lineRanges)}</p>`
  })

  return parts.join('') || '<p></p>'
}

export const getTeiEditableHighlights = (
  tei: string,
  highlightConfig?: TeiHighlightConfig,
): TeiEditableHighlight[] => {
  try {
    const doc = parseXml(tei.trim())
    const body = getBody(doc)
    const { paragraphSpans } = getOriginalParagraphSpans(
      doc,
      body,
      { showPB: true, minCert: 0.8, maskChar: '@', alignLines: false },
      highlightConfig,
    )
    const out: TeiEditableHighlight[] = []
    for (const [paragraphIndex, spans] of paragraphSpans.entries()) {
      for (const span of spans) {
        if (!span.fromAnchorId || !span.toAnchorId) {
          continue
        }
        out.push({
          id: span.id,
          paragraphIndex,
          start: span.start,
          end: span.end,
          featureId: span.featureId,
          categoryId: span.categoryId,
          categoryLabel: span.categoryLabel,
          surface: span.surface,
          normalized: span.normalized,
          institution: span.institution,
          ancientPersona: span.ancientPersona,
          fromAnchorId: span.fromAnchorId,
          toAnchorId: span.toAnchorId,
        })
      }
    }
    return out
  } catch {
    return []
  }
}

const toTextBlockId = (element: Element, fallback: string) => {
  const facs = parseCorrespRefs(element.getAttribute('facs'))
  if (facs.length > 0) {
    return facs[0]
  }
  const xmlId = getXmlId(element)
  if (xmlId) {
    return xmlId
  }
  const corresp = parseCorrespRefs(element.getAttribute('corresp'))
  if (corresp.length > 0) {
    return corresp[0]
  }
  return fallback
}

const toEditableLineText = (element: Element) =>
  trimTextWithAnchors(
    toReadingTextWithAnchors(element, {
      showPB: false,
      minCert: 0,
      maskChar: '@',
      alignLines: false,
    }),
  ).text

export const getTeiOriginalEditableLines = (
  tei: string,
): TeiOriginalEditableLine[] => {
  try {
    const doc = parseXml(tei.trim())
    const body = getBody(doc)
    const out: TeiOriginalEditableLine[] = []
    const transcriptionDivs = getDirectChildrenByName(body, 'div').filter(
      (div) => getElementAttr(div, 'type') === 'transcription',
    )

    for (let divIndex = 0; divIndex < transcriptionDivs.length; divIndex++) {
      const div = transcriptionDivs[divIndex]
      const blocks = getDirectChildrenByName(div, 'ab')
      const containers = blocks.length ? blocks : [div]

      for (
        let containerIndex = 0;
        containerIndex < containers.length;
        containerIndex++
      ) {
        const container = containers[containerIndex]
        const blockId = toTextBlockId(
          container,
          `block:${divIndex}:${containerIndex}`,
        )
        const lines = getLineElements(container)

        if (!lines.length) {
          out.push({
            id: `${blockId}:0`,
            blockId,
            text: toEditableLineText(container),
          })
          continue
        }

        for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
          out.push({
            id:
              getXmlId(lines[lineIndex]) ||
              `${blockId}:${String(lineIndex + 1)}`,
            blockId,
            text: toEditableLineText(lines[lineIndex]),
          })
        }
      }
    }

    return out
  } catch {
    return []
  }
}

function getBody(doc: Document): Element {
  const body =
    doc.getElementsByTagNameNS('*', 'body')[0] ||
    doc.getElementsByTagNameNS('*', 'text')[0] ||
    doc.documentElement
  return body as Element
}

const parseZoneCoordinate = (zone: Element, attr: string) =>
  Number.parseFloat(getElementAttr(zone, attr))

const parseBoundsFromElement = (element: Element | null) => {
  if (!element) {
    return null
  }
  const ulx = Number.parseFloat(getElementAttr(element, 'ulx'))
  const uly = Number.parseFloat(getElementAttr(element, 'uly'))
  const lrx = Number.parseFloat(getElementAttr(element, 'lrx'))
  const lry = Number.parseFloat(getElementAttr(element, 'lry'))
  if (
    !Number.isFinite(ulx) ||
    !Number.isFinite(uly) ||
    !Number.isFinite(lrx) ||
    !Number.isFinite(lry) ||
    lrx <= ulx ||
    lry <= uly
  ) {
    return null
  }
  return { ulx, uly, lrx, lry }
}

const almostEqual = (left: number, right: number, epsilon = 0.001) =>
  Math.abs(left - right) <= epsilon

const sameBounds = (
  left: { ulx: number; uly: number; lrx: number; lry: number } | null,
  right: { ulx: number; uly: number; lrx: number; lry: number } | null,
) => {
  if (!left && !right) {
    return true
  }
  if (!left || !right) {
    return false
  }
  return (
    almostEqual(left.ulx, right.ulx) &&
    almostEqual(left.uly, right.uly) &&
    almostEqual(left.lrx, right.lrx) &&
    almostEqual(left.lry, right.lry)
  )
}

const zoneContains = (
  block: { ulx: number; uly: number; lrx: number; lry: number },
  zone: { ulx: number; uly: number; lrx: number; lry: number },
  tolerance = 1,
) =>
  zone.ulx >= block.ulx - tolerance &&
  zone.uly >= block.uly - tolerance &&
  zone.lrx <= block.lrx + tolerance &&
  zone.lry <= block.lry + tolerance

const toZoneType = (zone: Element) =>
  [
    getElementAttr(zone, 'type'),
    getElementAttr(zone, 'subtype'),
    getElementAttr(zone, 'rendition'),
  ]
    .join(' ')
    .toLowerCase()

const isTextBlockType = (value: string) =>
  /text[\s_-]*block/.test(value.replace('#', ' '))

const getBlockToLineZoneLinks = (doc: Document) => {
  const links = new Map<string, Set<string>>()
  const blocks = doc.getElementsByTagNameNS('*', 'ab')
  for (let index = 0; index < blocks.length; index++) {
    const block = blocks[index]
    const blockZoneIds = parseCorrespRefs(block.getAttribute('facs'))
    if (!blockZoneIds.length) {
      continue
    }
    const lines = block.getElementsByTagNameNS('*', 'l')
    for (let lineIndex = 0; lineIndex < lines.length; lineIndex++) {
      const lineZoneIds = parseCorrespRefs(
        lines[lineIndex].getAttribute('facs'),
      )
      for (const blockZoneId of blockZoneIds) {
        const current = links.get(blockZoneId) || new Set<string>()
        for (const lineZoneId of lineZoneIds) {
          current.add(lineZoneId)
        }
        links.set(blockZoneId, current)
      }
    }
  }
  return links
}

const getZoneToTextMatchIds = (doc: Document) => {
  const links = new Map<string, Set<string>>()
  const add = (zoneId: string, ids: string[]) => {
    if (!zoneId || !ids.length) {
      return
    }
    const current = links.get(zoneId) || new Set<string>()
    for (const id of ids) {
      if (id) {
        current.add(id)
      }
    }
    links.set(zoneId, current)
  }

  const blocks = doc.getElementsByTagNameNS('*', 'ab')
  for (let index = 0; index < blocks.length; index++) {
    const block = blocks[index]
    const blockZoneIds = parseCorrespRefs(block.getAttribute('facs'))
    const blockTextIds = toUniqueSorted([
      getXmlId(block),
      ...parseCorrespRefs(block.getAttribute('corresp')),
    ])
    for (const zoneId of blockZoneIds) {
      add(zoneId, blockTextIds)
    }
  }

  const lines = doc.getElementsByTagNameNS('*', 'l')
  for (let index = 0; index < lines.length; index++) {
    const line = lines[index]
    const lineZoneIds = parseCorrespRefs(line.getAttribute('facs'))
    const lineTextIds = toUniqueSorted([
      getXmlId(line),
      ...parseCorrespRefs(line.getAttribute('corresp')),
    ])
    for (const zoneId of lineZoneIds) {
      add(zoneId, lineTextIds)
    }
  }

  return links
}

export const getTeiSurfaceZones = (tei: string): TeiSurfaceZone[] => {
  try {
    const doc = parseXml(tei.trim())
    const zones = doc.getElementsByTagNameNS('*', 'zone')
    const parsed: Array<
      Omit<
        TeiSurfaceZone,
        | 'hoverMatchIds'
        | 'zoneType'
        | 'refUlx'
        | 'refUly'
        | 'refLrx'
        | 'refLry'
        | 'hasSurfaceBounds'
      > & {
        parentBounds: {
          ulx: number
          uly: number
          lrx: number
          lry: number
        } | null
        type: string
        element: Element
      }
    > = []
    for (let index = 0; index < zones.length; index++) {
      const zone = zones[index]
      const id = getXmlId(zone) || `zone:${index}`
      const ulx = parseZoneCoordinate(zone, 'ulx')
      const uly = parseZoneCoordinate(zone, 'uly')
      const lrx = parseZoneCoordinate(zone, 'lrx')
      const lry = parseZoneCoordinate(zone, 'lry')
      if (
        !Number.isFinite(ulx) ||
        !Number.isFinite(uly) ||
        !Number.isFinite(lrx) ||
        !Number.isFinite(lry) ||
        lrx <= ulx ||
        lry <= uly
      ) {
        continue
      }
      const matchIds = toUniqueSorted([
        id,
        ...parseCorrespRefs(zone.getAttribute('corresp')),
        ...parseCorrespRefs(zone.getAttribute('facs')),
      ])
      const parentBounds = parseBoundsFromElement(zone.closest('surface'))
      parsed.push({
        id,
        matchIds: matchIds.length ? matchIds : [id],
        ulx,
        uly,
        lrx,
        lry,
        parentBounds,
        type: toZoneType(zone),
        element: zone,
      })
    }
    if (!parsed.length) {
      return []
    }

    const textBlockZones = parsed.filter((zone) => isTextBlockType(zone.type))
    const matchIdsByZoneId = new Map<string, Set<string>>(
      parsed.map((zone) => [zone.id, new Set(zone.matchIds)]),
    )
    const hoverMatchIdsByZoneId = new Map<string, Set<string>>(
      parsed.map((zone) => [zone.id, new Set(zone.matchIds)]),
    )
    const blockToLineLinks = getBlockToLineZoneLinks(doc)
    const zoneToTextMatchIds = getZoneToTextMatchIds(doc)

    for (const [zoneId, textMatchIds] of zoneToTextMatchIds.entries()) {
      const target = matchIdsByZoneId.get(zoneId)
      const hoverTarget = hoverMatchIdsByZoneId.get(zoneId)
      if (!target) {
        continue
      }
      for (const id of textMatchIds) {
        target.add(id)
        hoverTarget?.add(id)
      }
    }

    for (const block of textBlockZones) {
      const blockSet = matchIdsByZoneId.get(block.id) || new Set<string>()
      const linkedLineZoneIds = blockToLineLinks.get(block.id)
      if (linkedLineZoneIds && linkedLineZoneIds.size > 0) {
        for (const lineZoneId of linkedLineZoneIds) {
          blockSet.add(lineZoneId)
        }
      }
      for (const zone of parsed) {
        if (zone.id === block.id) {
          continue
        }
        if (isTextBlockType(zone.type)) {
          continue
        }
        if (
          sameBounds(block.parentBounds, zone.parentBounds) &&
          (block.element.contains(zone.element) || zoneContains(block, zone))
        ) {
          const zoneSet = matchIdsByZoneId.get(zone.id)
          if (!zoneSet) {
            continue
          }
          for (const id of zoneSet) {
            blockSet.add(id)
          }
        }
      }
      matchIdsByZoneId.set(block.id, blockSet)
    }

    const out = parsed.map((zone) => ({
      ...zone,
      matchIds: [
        ...(matchIdsByZoneId.get(zone.id) || new Set(zone.matchIds)),
      ].sort(),
    }))

    const fallbackUlx = Math.min(...out.map((zone) => zone.ulx))
    const fallbackUly = Math.min(...out.map((zone) => zone.uly))
    const fallbackLrx = Math.max(...out.map((zone) => zone.lrx))
    const fallbackLry = Math.max(...out.map((zone) => zone.lry))
    const hasValidFallback =
      fallbackLrx > fallbackUlx && fallbackLry > fallbackUly
    const fallbackBounds = hasValidFallback
      ? {
          ulx: fallbackUlx,
          uly: fallbackUly,
          lrx: fallbackLrx,
          lry: fallbackLry,
        }
      : { ulx: 0, uly: 0, lrx: 1, lry: 1 }

    return out.map((zone) => {
      const bounds = zone.parentBounds || fallbackBounds
      return {
        id: zone.id,
        matchIds: zone.matchIds,
        hoverMatchIds: [
          ...(hoverMatchIdsByZoneId.get(zone.id) || new Set(zone.matchIds)),
        ].sort(),
        zoneType: isTextBlockType(zone.type) ? 'block' : 'line',
        ulx: zone.ulx,
        uly: zone.uly,
        lrx: zone.lrx,
        lry: zone.lry,
        refUlx: bounds.ulx,
        refUly: bounds.uly,
        refLrx: bounds.lrx,
        refLry: bounds.lry,
        hasSurfaceBounds: !!zone.parentBounds,
      }
    })
  } catch {
    return []
  }
}

function applyHighlights(
  html: string,
  searchResultHighlight: string | null,
): string {
  const highlights = searchResultHighlight
    ? [...searchResultHighlight.matchAll(/<em>(.*?)<\/em>/g)].map(
        (match) => match[1],
      )
    : []
  let out = html
  for (const highlight of highlights) {
    out = out.replaceAll(highlight, `<em>${highlight}</em>`)
  }
  return out
}

export const teiToHtml = (
  tei: string,
  minCert: number,
  searchResultHighlight: string | null,
  maskChar: string = '@',
  viewMode: TeiViewMode = 'original',
  alignLines: boolean = false,
  highlightConfig?: TeiHighlightConfig,
) => {
  const doc = parseXml(tei.trim())
  const body = getBody(doc)
  const opts = { showPB: true, minCert, maskChar, alignLines }

  if (viewMode !== 'original') {
    const translationIndex = Number.parseInt(viewMode.split(':')[1] || '', 10)
    const joined = Number.isFinite(translationIndex)
      ? renderTranslationView(body, translationIndex, opts)
      : '<p></p>'
    return applyHighlights(joined, searchResultHighlight)
  }

  const joined = renderOriginalView(doc, body, opts, highlightConfig)
  return applyHighlights(joined, searchResultHighlight)
}
