import {
  getBody,
  getCertaintyDegreeByTargetId,
  getDirectChildrenByName,
  getElementAttr,
  getElementCertaintyDegree,
  getLineElements,
  getTeiTranslationSources,
  getXmlId,
  isElement,
  parseCorrespRefs,
  toUniqueSorted,
} from './teiDom.ts'
import { getTeiHighlightSpans, isVerbCategory } from './teiHighlights.ts'
import type {
  LineMatchMode,
  LineTextWithAnchors,
  ParagraphAnchorLocation,
  ParagraphHighlightSpan,
  ParagraphLineRange,
  ParagraphTextWithAnchors,
  ReadingOptions,
  TeiHighlightConfig,
  TextWithAnchors,
} from './teiTypes.ts'

export const escapeHtml = (text: string) => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

export const escapeHtmlAttr = (text: string) =>
  escapeHtml(text).replaceAll('"', '&quot;').replaceAll("'", '&#39;')

export function textContentExcludingCertainty(node: ChildNode) {
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

export function maskText(s: string, maskChar: string) {
  const m = maskChar && String(maskChar).length ? String(maskChar)[0] : '@'
  let out = ''
  for (const ch of s) {
    out += /\s/.test(ch) ? ch : m
  }
  return out
}

export const getDisplayedCertaintyDegree = (
  degree: number | null,
): number | null => {
  if (degree == null || !Number.isFinite(degree)) {
    return null
  }
  return Math.min(1, Math.max(0.8, degree))
}

export const shouldMaskByCertainty = (
  degree: number | null,
  opts: ReadingOptions,
): boolean => {
  const minCert = Number.isFinite(opts.minCert) ? opts.minCert : 0
  const displayedDegree = getDisplayedCertaintyDegree(degree)
  return displayedDegree != null && displayedDegree < minCert
}

const startsWithClosingPunctuation = (text: string) =>
  /^[\s]*[.,;:!?)[\]{}]/.test(text)

const trimTrailingSpaces = (text: string) => text.replace(/ +$/, '')

const clampTrailingAnchorOffsets = (
  anchors: Record<string, number>,
  nextLength: number,
) => {
  for (const id of Object.keys(anchors)) {
    if (anchors[id] > nextLength) {
      anchors[id] = nextLength
    }
  }
}

export const appendTextWithAnchors = (
  node: ChildNode,
  opts: ReadingOptions,
  builder: TextWithAnchors,
) => {
  if (node.nodeType === Node.TEXT_NODE) {
    const raw = node.nodeValue || ''
    let text = /[\n\r\t]/.test(raw) ? raw.replace(/\s+/g, ' ') : raw
    if (startsWithClosingPunctuation(text)) {
      builder.text = trimTrailingSpaces(builder.text)
      clampTrailingAnchorOffsets(builder.anchors, builder.text.length)
      text = text.replace(/^\s+/, '')
    }
    builder.text += text
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
    const degree = getElementCertaintyDegree(element, opts)
    const rawText = textContentExcludingCertainty(element)
    builder.text += shouldMaskByCertainty(degree, opts)
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

export const toReadingTextWithAnchors = (
  node: Element,
  opts: ReadingOptions,
) => {
  const builder: TextWithAnchors = { text: '', anchors: {} }
  for (let i = 0; i < node.childNodes.length; i++) {
    appendTextWithAnchors(node.childNodes[i], opts, builder)
  }
  return builder
}

export const trimTextWithAnchors = (
  value: TextWithAnchors,
): TextWithAnchors => {
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
  blockType?: string,
): ParagraphTextWithAnchors[] => {
  if (!alignLines) {
    const paragraphs: ParagraphTextWithAnchors[] = []
    let currentText = ''
    let currentAnchors: Record<string, number> = {}
    let currentLineRanges: ParagraphLineRange[] = []

    const pushCurrent = () => {
      if (!currentText) {
        return
      }
      paragraphs.push({
        text: currentText,
        anchors: currentAnchors,
        lineRanges: currentLineRanges,
        blockType,
      })
      currentText = ''
      currentAnchors = {}
      currentLineRanges = []
    }

    for (const line of lines) {
      if (!line.text) {
        pushCurrent()
        continue
      }

      if (currentText) {
        currentText += '\n'
      }
      const offset = currentText.length
      currentText += line.text
      const end = currentText.length
      if (line.matchIds.length > 0 && end > offset) {
        currentLineRanges.push({
          start: offset,
          end,
          matchIds: line.matchIds,
          certaintyDegree: line.certaintyDegree,
        })
      }
      for (const [id, pos] of Object.entries(line.anchors)) {
        currentAnchors[id] = offset + pos
      }
    }

    pushCurrent()
    return paragraphs.length
      ? paragraphs
      : [{ text: '', anchors: {}, lineRanges: [], blockType }]
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
      blockType,
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

    if (
      currentText &&
      !previousLineEndedWithMergeDash &&
      !startsWithClosingPunctuation(line.text)
    ) {
      currentText += ' '
    } else if (startsWithClosingPunctuation(line.text)) {
      currentText = trimTrailingSpaces(currentText)
      clampTrailingAnchorOffsets(currentAnchors, currentText.length)
    }
    const offset = currentText.length
    currentText += line.text
    const end = currentText.length
    if (line.matchIds.length > 0 && end > offset) {
      currentLineRanges.push({
        start: offset,
        end,
        matchIds: line.matchIds,
        certaintyDegree: line.certaintyDegree,
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
    : [{ text: '', anchors: {}, lineRanges: [], blockType }]
}

const getTeiBlockType = (element: Element) =>
  getElementAttr(element, 'type') || undefined

export const renderStructuredDivToParagraphs = (
  div: Element,
  opts: ReadingOptions,
  lineMatchMode: LineMatchMode = 'none',
): ParagraphTextWithAnchors[] => {
  const blocks = getDirectChildrenByName(div, 'ab')
  const containers = blocks.length ? blocks : [div]
  const paragraphs: ParagraphTextWithAnchors[] = []

  for (const container of containers) {
    const blockType = getTeiBlockType(container)
    const lines = getLineElements(container)
    if (lines.length) {
      const renderedLines = lines.map((line) => {
        const lineText = trimTextWithAnchors(
          toReadingTextWithAnchors(line, opts),
        )
        const degree = getElementCertaintyDegree(line, opts)
        return {
          ...lineText,
          text: shouldMaskByCertainty(degree, opts)
            ? maskText(lineText.text, opts.maskChar)
            : lineText.text,
          matchIds: getMatchIdsForElement(line, lineMatchMode),
          certaintyDegree: degree,
        } satisfies LineTextWithAnchors
      })
      const blocks = joinLineTexts(renderedLines, opts.alignLines, blockType)
      for (const block of blocks) {
        paragraphs.push(block)
      }
      continue
    }

    const block = trimTextWithAnchors(toReadingTextWithAnchors(container, opts))
    const degree = getElementCertaintyDegree(container, opts)
    const blockMatchIds = getMatchIdsForElement(container, lineMatchMode)
    paragraphs.push({
      ...block,
      text: shouldMaskByCertainty(degree, opts)
        ? maskText(block.text, opts.maskChar)
        : block.text,
      lineRanges:
        blockMatchIds.length > 0 && block.text.length > 0
          ? [
              {
                start: 0,
                end: block.text.length,
                matchIds: blockMatchIds,
                certaintyDegree: degree,
              },
            ]
          : [],
      blockType,
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

export const renderParagraphWithHighlights = (
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

export const renderParagraphWithLineRanges = (
  text: string,
  spans: ParagraphHighlightSpan[],
  paragraphIndex: number,
  lineRanges: ParagraphLineRange[],
  showCertaintyVisualization: boolean,
) => {
  const validRanges = lineRanges
    .map((range) => ({
      start: Math.max(0, Math.min(text.length, range.start)),
      end: Math.max(0, Math.min(text.length, range.end)),
      matchIds: [...new Set(range.matchIds.filter(Boolean))],
      certaintyDegree:
        range.certaintyDegree != null && Number.isFinite(range.certaintyDegree)
          ? range.certaintyDegree
          : null,
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
    const attrs = [
      `data-tei-line-match-ids="${escapeHtmlAttr(range.matchIds.join(' '))}"`,
    ]
    if (showCertaintyVisualization && range.certaintyDegree != null) {
      const displayDegree = getDisplayedCertaintyDegree(range.certaintyDegree)
      if (displayDegree == null) {
        html += `<span ${attrs.join(' ')}>${lineHtml}</span>`
        cursor = rangeEnd
        continue
      }
      attrs.push(`data-tei-certainty-degree="${displayDegree.toFixed(3)}"`)
      if (range.certaintyDegree >= 0.8) {
        const strength = (1 - displayDegree) / 0.2
        const alpha = strength * 0.42
        attrs.push(
          `style="background-color: rgba(249, 115, 22, ${alpha.toFixed(3)}); border-radius: 3px;"`,
        )
      }
    }
    html += `<span ${attrs.join(' ')}>${lineHtml}</span>`
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

export const renderStructuredDiv = (
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
        !!opts.showCertaintyVisualization,
      )
      const blockTypeAttr = paragraph.blockType
        ? ` data-tei-block-type="${escapeHtmlAttr(paragraph.blockType)}"`
        : ''
      return `<p${blockTypeAttr}>${html}</p>`
    })
    .join('')
}

export const getOriginalParagraphSpans = (
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

export const renderOriginalView = (
  doc: Document,
  body: Element,
  opts: ReadingOptions,
  highlightConfig?: TeiHighlightConfig,
) => {
  const { paragraphs, paragraphSpans } = getOriginalParagraphSpans(
    doc,
    body,
    opts,
    highlightConfig,
  )
  const parts = paragraphs.map((paragraph, index) => {
    const spans = paragraphSpans.get(index) || []
    const paragraphTextAttr = escapeHtmlAttr(encodeURIComponent(paragraph.text))
    const blockTypeAttr = paragraph.blockType
      ? ` data-tei-block-type="${escapeHtmlAttr(paragraph.blockType)}"`
      : ''
    return `<p data-tei-paragraph-index="${index}" data-tei-paragraph-text="${paragraphTextAttr}"${blockTypeAttr}>${renderParagraphWithLineRanges(paragraph.text, spans, index, paragraph.lineRanges, !!opts.showCertaintyVisualization)}</p>`
  })

  return parts.join('') || '<p></p>'
}

export const renderTranslationView = (
  body: Element,
  translationIndex: number,
  opts: ReadingOptions,
): string => {
  const sources = getTeiTranslationSources(body)
  const source = sources[translationIndex]
  if (!source) return '<p></p>'
  return renderStructuredDiv(source.element, opts, 'corresp') || '<p></p>'
}

export function applyHighlights(
  html: string,
  searchResultHighlight: string | null,
): string {
  const highlights = searchResultHighlight
    ? (() => {
        const container = document.createElement('div')
        container.innerHTML = searchResultHighlight
        return [
          ...new Set(
            [...container.querySelectorAll('em')]
              .map((element) => element.textContent?.trim() || '')
              .filter(Boolean),
          ),
        ].sort((left, right) => right.length - left.length)
      })()
    : []

  if (highlights.length === 0) {
    return html
  }

  const root = document.createElement('div')
  root.innerHTML = html
  const escapedHighlights = highlights.map((highlight) =>
    highlight.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'),
  )
  const highlightPattern = new RegExp(`(${escapedHighlights.join('|')})`, 'g')
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  const textNodes: Text[] = []

  let currentNode = walker.nextNode()
  while (currentNode) {
    if (
      currentNode instanceof Text &&
      currentNode.parentElement?.closest('em') == null &&
      highlights.some((highlight) =>
        currentNode?.textContent?.includes(highlight),
      )
    ) {
      textNodes.push(currentNode)
    }
    currentNode = walker.nextNode()
  }

  for (const textNode of textNodes) {
    const text = textNode.textContent || ''
    const parts = text.split(highlightPattern)
    if (parts.length <= 1) {
      continue
    }

    const fragment = document.createDocumentFragment()
    for (const part of parts) {
      if (!part) continue
      if (highlights.includes(part)) {
        const highlight = document.createElement('em')
        highlight.textContent = part
        fragment.appendChild(highlight)
      } else {
        fragment.appendChild(document.createTextNode(part))
      }
    }
    textNode.replaceWith(fragment)
  }

  return root.innerHTML
}

export const buildReadingOptions = (
  doc: Document,
  minCert: number,
  maskChar: string,
  alignLines: boolean,
  showCertaintyVisualization: boolean,
): ReadingOptions => ({
  showPB: true,
  minCert,
  maskChar,
  alignLines,
  showCertaintyVisualization,
  certaintyDegreeByTargetId: getCertaintyDegreeByTargetId(doc),
})

export const renderTeiHtml = (
  doc: Document,
  minCert: number,
  searchResultHighlight: string | null,
  maskChar: string,
  viewMode: string,
  alignLines: boolean,
  showCertaintyVisualization: boolean,
  highlightConfig?: TeiHighlightConfig,
) => {
  const body = getBody(doc)
  const opts = buildReadingOptions(
    doc,
    minCert,
    maskChar,
    alignLines,
    showCertaintyVisualization,
  )

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
