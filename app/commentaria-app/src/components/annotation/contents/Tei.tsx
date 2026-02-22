import { useMemo } from 'react'
import { useAppState } from '../../../context/useAppState.ts'

export type TeiViewMode = 'original' | 'translation'

type Props = {
  data: string
  minCert: number
  viewMode?: TeiViewMode
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
      // Render line breaks as space so text flows; only <p> creates visual breaks
      html += ' '
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

/** Whether this TEI has a translations div (edition-style original + translation). */
export function teiHasTranslations(tei: string): boolean {
  try {
    const doc = parseXml(tei.trim())
    const body = getBody(doc)
    const divs = body.getElementsByTagNameNS('*', 'div')
    for (let i = 0; i < divs.length; i++) {
      const div = divs[i]
      if (div.getAttribute('type') === 'translations') {
        const abs = div.getElementsByTagNameNS('*', 'ab')
        for (let j = 0; j < abs.length; j++) {
          if (abs[j].getAttribute('type') === 'translation') return true
        }
      }
    }
    return false
  } catch {
    return false
  }
}

function renderTranslationView(body: Element): string {
  const divs = body.getElementsByTagNameNS('*', 'div')
  const parts: string[] = []
  for (let i = 0; i < divs.length; i++) {
    const div = divs[i]
    if (div.getAttribute('type') !== 'translations') continue
    const abs = div.getElementsByTagNameNS('*', 'ab')
    for (let j = 0; j < abs.length; j++) {
      const ab = abs[j]
      if (ab.getAttribute('type') !== 'translation') continue
      const segs = ab.getElementsByTagNameNS('*', 'seg')
      const texts: string[] = []
      for (let k = 0; k < segs.length; k++) {
        texts.push((segs[k].textContent || '').trim())
      }
      const paragraph = texts.filter(Boolean).join(' ')
      if (paragraph) parts.push(`<p>${escapeHtml(paragraph)}</p>`)
    }
  }
  return parts.length ? parts.join('') : '<p></p>'
}

const teiToHtml = (
  tei: string,
  minCert: number,
  searchResultHighlight: string | null,
  maskChar: string = '@',
  viewMode: TeiViewMode = 'original',
) => {
  const doc = parseXml(tei.trim())
  const body = getBody(doc)
  const opts = { showPB: true, minCert, maskChar }

  if (viewMode === 'translation') {
    const joined = renderTranslationView(body)
    const highlights = searchResultHighlight
      ? [...searchResultHighlight.matchAll(/<em>(.*?)<\/em>/g)].map(
          (match) => match[1],
        )
      : []
    let out = joined
    for (const highlight of highlights) {
      out = out.replaceAll(highlight, `<em>${highlight}</em>`)
    }
    return out
  }

  // Original: only direct <p> children of body (exclude div type="translations")
  const directPs: Element[] = []
  for (let i = 0; i < body.children.length; i++) {
    const el = body.children[i]
    if (el.nodeType === Node.ELEMENT_NODE && (el as Element).localName === 'p') {
      directPs.push(el as Element)
    }
  }

  const parts: string[] = []
  if (directPs.length) {
    for (const p of directPs) {
      const inner = toReadingHtml(p, opts).trim()
      if (!inner) continue
      parts.push(`<p>${inner || '&nbsp;'}</p>`)
    }
  } else {
    const inner = toReadingHtml(body, opts).trim()
    parts.push(`<p>${inner || '&nbsp;'}</p>`)
  }
  let joined = parts.join('')

  const highlights = searchResultHighlight
    ? [...searchResultHighlight.matchAll(/<em>(.*?)<\/em>/g)].map(
        (match) => match[1],
      )
    : []
  for (const highlight of highlights) {
    joined = joined.replaceAll(highlight, `<em>${highlight}</em>`)
  }
  return joined
}

export const Tei = ({ minCert, data, viewMode = 'original' }: Props) => {
  const { searchResultHighlight } = useAppState()
  const html = useMemo(
    () => teiToHtml(data, minCert, searchResultHighlight, '@', viewMode),
    [data, minCert, searchResultHighlight, viewMode],
  )
  return (
    <div
      className="mt-4 text-xs leading-relaxed border border-gray-300 rounded-xl bg-gray-50 p-2 [&_p]:mb-2 [&_p:last-child]:mb-0 [&_[data-tei-selected='true']]:bg-yellow-200/70 [&_[data-tei-selected='true']]:text-gray-900 [&_[data-tei-selected='true']]:rounded-sm [&_[data-tei-selected='true']]:px-0.5"
      style={{ whiteSpace: 'normal' }}
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}
