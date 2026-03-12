import type { ReadingOptions, TeiTranslationSource } from './teiTypes.ts'

export const parseXml = (xmlString: string) => {
  const parser = new DOMParser()
  const doc = parser.parseFromString(xmlString, 'application/xml')
  const pe = doc.getElementsByTagName('parsererror')[0]
  if (pe) {
    throw new Error('XML parse error (often unescaped & or mismatched tags).')
  }
  return doc
}

export const isElement = (
  node: ChildNode,
  localName: string,
): node is Element =>
  node &&
  node.nodeType === Node.ELEMENT_NODE &&
  (node as Element).localName === localName

export function findFirstCertaintyDegree(segEl: Element) {
  const certs = segEl.getElementsByTagNameNS('*', 'certainty')
  if (!certs || !certs.length) return null

  const degreeStr = certs[0].getAttribute('degree')
  const degree = degreeStr == null ? NaN : parseFloat(degreeStr)
  return Number.isFinite(degree) ? degree : null
}

export const getCertaintyDegreeByTargetId = (doc: Document) => {
  const out = new Map<string, number>()
  const certs = doc.getElementsByTagNameNS('*', 'certainty')
  for (let i = 0; i < certs.length; i++) {
    const cert = certs[i]
    const degree = Number.parseFloat(cert.getAttribute('degree') || '')
    if (!Number.isFinite(degree)) {
      continue
    }
    const targets = parseCorrespRefs(cert.getAttribute('target'))
    for (const target of targets) {
      if (!target) {
        continue
      }
      out.set(target, degree)
    }
  }
  return out
}

export const getElementCertaintyDegree = (
  element: Element,
  opts: ReadingOptions,
): number | null => {
  const targetId = getXmlId(element)
  if (targetId) {
    const byTargetIdDegree = opts.certaintyDegreeByTargetId?.get(targetId)
    if (byTargetIdDegree != null && Number.isFinite(byTargetIdDegree)) {
      return byTargetIdDegree
    }
  }
  return findFirstCertaintyDegree(element)
}

export const getElementAttr = (el: Element, name: string) =>
  (el.getAttribute(name) || '').trim()

export const getXmlId = (el: Element) =>
  getElementAttr(el, 'xml:id') || getElementAttr(el, 'id')

export const parseAnaRefs = (value: string | null) =>
  (value || '')
    .split(/\s+/)
    .map((entry) => entry.replace(/^#/, '').trim())
    .filter(Boolean)

export const parseCorrespRefs = (value: string | null) =>
  (value || '')
    .split(/\s+/)
    .map((entry) => entry.replace(/^#/, '').trim())
    .filter(Boolean)

export const toUniqueSorted = (values: Array<string | null | undefined>) =>
  [
    ...new Set(values.map((value) => (value || '').trim()).filter(Boolean)),
  ].sort()

export const getDirectChildrenByName = (parent: Element, localName: string) => {
  const children: Element[] = []
  for (let i = 0; i < parent.children.length; i++) {
    const child = parent.children[i]
    if (child.localName === localName) {
      children.push(child)
    }
  }
  return children
}

export const getLineElements = (container: Element): Element[] => {
  const directLines = getDirectChildrenByName(container, 'l')
  if (directLines.length) return directLines
  const lines = container.getElementsByTagNameNS('*', 'l')
  const out: Element[] = []
  for (let i = 0; i < lines.length; i++) {
    out.push(lines[i])
  }
  return out
}

export function getBody(doc: Document): Element {
  const body =
    doc.getElementsByTagNameNS('*', 'body')[0] ||
    doc.getElementsByTagNameNS('*', 'text')[0] ||
    doc.documentElement
  return body as Element
}

export const getXmlLang = (el: Element) =>
  getElementAttr(el, 'xml:lang') || getElementAttr(el, 'lang')

export const getTeiTranslationSources = (
  body: Element,
): TeiTranslationSource[] => {
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
