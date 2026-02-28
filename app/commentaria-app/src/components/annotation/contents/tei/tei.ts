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
    builder.text += node.nodeValue || ''
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

const joinLineTexts = (
  lines: TextWithAnchors[],
  alignLines: boolean,
): ParagraphTextWithAnchors[] => {
  if (!alignLines) {
    let text = ''
    const anchors: Record<string, number> = {}

    for (let i = 0; i < lines.length; i++) {
      if (i > 0) {
        text += '\n'
      }
      const offset = text.length
      text += lines[i].text
      for (const [id, pos] of Object.entries(lines[i].anchors)) {
        anchors[id] = offset + pos
      }
    }

    return [{ text, anchors }]
  }

  const paragraphs: ParagraphTextWithAnchors[] = []
  let currentText = ''
  let currentAnchors: Record<string, number> = {}

  const pushCurrent = () => {
    if (!currentText) {
      return
    }
    paragraphs.push({ text: currentText, anchors: currentAnchors })
    currentText = ''
    currentAnchors = {}
  }

  for (const line of lines) {
    if (!line.text) {
      pushCurrent()
      continue
    }

    if (currentText) {
      currentText += ' '
    }
    const offset = currentText.length
    currentText += line.text
    for (const [id, pos] of Object.entries(line.anchors)) {
      currentAnchors[id] = offset + pos
    }
  }

  pushCurrent()
  return paragraphs.length ? paragraphs : [{ text: '', anchors: {} }]
}

const renderStructuredDivToParagraphs = (
  div: Element,
  opts: ReadingOptions,
): ParagraphTextWithAnchors[] => {
  const blocks = getDirectChildrenByName(div, 'ab')
  const containers = blocks.length ? blocks : [div]
  const paragraphs: ParagraphTextWithAnchors[] = []

  for (const container of containers) {
    const lines = getLineElements(container)
    if (lines.length) {
      const renderedLines = lines.map((line) =>
        trimTextWithAnchors(toReadingTextWithAnchors(line, opts)),
      )
      const blocks = joinLineTexts(renderedLines, opts.alignLines)
      for (const block of blocks) {
        paragraphs.push(block)
      }
      continue
    }

    paragraphs.push(
      trimTextWithAnchors(toReadingTextWithAnchors(container, opts)),
    )
  }

  return paragraphs
}

const renderStructuredDiv = (div: Element, opts: ReadingOptions) => {
  const paragraphs = renderStructuredDivToParagraphs(div, opts)
  return paragraphs
    .map((paragraph) => {
      const html = escapeHtml(paragraph.text).replaceAll('\n', '<br>')
      return `<p>${html || '&nbsp;'}</p>`
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
  return renderStructuredDiv(source.element, opts) || '<p></p>'
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
      start: span.start,
      end: span.end,
      fromAnchorId: span.fromAnchorId || '',
      toAnchorId: span.toAnchorId || '',
      color: span.color,
    }))
    const tooltipItemsAttr = escapeHtmlAttr(
      encodeURIComponent(JSON.stringify(tooltipItems)),
    )
    html += `<span data-tei-highlight="true" data-tei-highlight-tooltip="${tooltipItemsAttr}" style="${style}">${escapedSegment}</span>`
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
      const rendered = renderStructuredDivToParagraphs(child, opts)
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
    return `<p data-tei-paragraph-index="${index}">${renderParagraphWithHighlights(paragraph.text, spans, index)}</p>`
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

function getBody(doc: Document): Element {
  const body =
    doc.getElementsByTagNameNS('*', 'body')[0] ||
    doc.getElementsByTagNameNS('*', 'text')[0] ||
    doc.documentElement
  return body as Element
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
