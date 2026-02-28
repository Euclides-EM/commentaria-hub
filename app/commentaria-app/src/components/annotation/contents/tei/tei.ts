export type TeiViewMode = 'original' | `translation:${number}`

export type TeiTranslation = {
  id: `translation:${number}`
  label: string
}

type TeiTranslationSource = {
  label: string
  element: Element
}

const escapeHtml = (text: string) => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

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
  // Look for a descendant <certainty degree="..."> inside the seg
  const certs = segEl.getElementsByTagNameNS('*', 'certainty')
  if (!certs || !certs.length) return null

  const degreeStr = certs[0].getAttribute('degree')
  const degree = degreeStr == null ? NaN : parseFloat(degreeStr)
  return Number.isFinite(degree) ? degree : null
}

function textContentExcludingCertainty(node: ChildNode) {
  let out = ''
  for (const child of node.childNodes) {
    if (child.nodeType === Node.TEXT_NODE) {
      out += child.nodeValue || ''
      continue
    }
    if (child.nodeType !== Node.ELEMENT_NODE) continue

    // Skip certainty markup from the visible text
    if (isElement(child, 'certainty')) continue

    // Recurse
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

function toReadingHtml(
  node: Element,
  opts: {
    showPB: boolean
    minCert: number
    maskChar: string
    alignLines: boolean
  },
) {
  let html = ''

  for (const child of node.childNodes) {
    if (child.nodeType === Node.TEXT_NODE) {
      html += escapeHtml(child.nodeValue || '')
      continue
    }
    if (child.nodeType !== Node.ELEMENT_NODE) continue

    if (isElement(child, 'lb')) {
      html += opts.alignLines ? ' ' : '<br>'
      continue
    }

    if (isElement(child, 'pb')) {
      if (opts.showPB) {
        const facs = child.getAttribute('facs') || ''
        const n = child.getAttribute('n') || ''
        const label =
          n || facs ? `Page break ${escapeHtml(n || facs)}` : 'Page break'
        html += `<div class="pb">${label}</div>`
      }
      continue
    }

    // Special handling: <seg> ... <certainty degree="..."/> ... </seg>
    if (isElement(child, 'seg')) {
      const degree = findFirstCertaintyDegree(child)
      const rawText = textContentExcludingCertainty(child)

      const minCert = Number.isFinite(opts.minCert) ? opts.minCert : 0
      const masked =
        degree != null && degree < minCert
          ? maskText(rawText, opts.maskChar)
          : rawText

      html += escapeHtml(masked)
      continue
    }

    const inner = toReadingHtml(child as Element, opts)
    html += inner
  }

  return html
}

function getBody(doc: Document): Element {
  const body =
    doc.getElementsByTagNameNS('*', 'body')[0] ||
    doc.getElementsByTagNameNS('*', 'text')[0] ||
    doc.documentElement
  return body as Element
}

const getElementAttr = (el: Element, name: string) =>
  (el.getAttribute(name) || '').trim()

const getXmlLang = (el: Element) =>
  getElementAttr(el, 'xml:lang') || getElementAttr(el, 'lang')

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

const renderLineBlock = (
  lineElements: Element[],
  opts: {
    showPB: boolean
    minCert: number
    maskChar: string
    alignLines: boolean
  },
) => {
  const renderedLines = lineElements.map((line) =>
    toReadingHtml(line, opts).trim(),
  )
  if (!opts.alignLines) {
    const withEmptyPlaceholders = renderedLines.map((line) => line || '&nbsp;')
    return [withEmptyPlaceholders.join('<br>')]
  }

  const paragraphs: string[] = []
  let current: string[] = []
  for (const line of renderedLines) {
    if (!line) {
      if (current.length > 0) {
        paragraphs.push(current.join(' ').trim())
        current = []
      }
      continue
    }
    current.push(line)
  }
  if (current.length > 0) {
    paragraphs.push(current.join(' ').trim())
  }
  return paragraphs.length ? paragraphs : ['&nbsp;']
}

const renderStructuredDiv = (
  div: Element,
  opts: {
    showPB: boolean
    minCert: number
    maskChar: string
    alignLines: boolean
  },
) => {
  const blocks = getDirectChildrenByName(div, 'ab')
  const containers = blocks.length ? blocks : [div]
  const parts: string[] = []
  for (const container of containers) {
    const lines = getLineElements(container)
    if (lines.length) {
      const blocks = renderLineBlock(lines, opts)
      for (const block of blocks) {
        parts.push(`<p>${block || '&nbsp;'}</p>`)
      }
      continue
    }
    const inner = toReadingHtml(container, opts).trim()
    parts.push(`<p>${inner || '&nbsp;'}</p>`)
  }
  return parts.join('')
}

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

function renderTranslationView(
  body: Element,
  translationIndex: number,
  opts: {
    showPB: boolean
    minCert: number
    maskChar: string
    alignLines: boolean
  },
): string {
  const sources = getTeiTranslationSources(body)
  const source = sources[translationIndex]
  if (!source) return '<p></p>'
  return renderStructuredDiv(source.element, opts) || '<p></p>'
}

function renderOriginalView(
  body: Element,
  opts: {
    showPB: boolean
    minCert: number
    maskChar: string
    alignLines: boolean
  },
) {
  const directDivs = getDirectChildrenByName(body, 'div')
  const transcriptionDivs = directDivs.filter(
    (div) => getElementAttr(div, 'type') === 'transcription',
  )

  if (transcriptionDivs.length) {
    const parts: string[] = []
    for (let i = 0; i < body.children.length; i++) {
      const child = body.children[i]
      if (
        child.localName === 'div' &&
        getElementAttr(child, 'type') === 'transcription'
      ) {
        const rendered = renderStructuredDiv(child, opts)
        if (rendered) {
          parts.push(rendered)
        }
      }
    }
    return parts.join('') || '<p></p>'
  }
  return '<p></p>'
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

  const joined = renderOriginalView(body, opts)
  return applyHighlights(joined, searchResultHighlight)
}
