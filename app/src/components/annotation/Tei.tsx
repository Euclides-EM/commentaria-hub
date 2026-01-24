import { useMemo } from 'react'
import { useAppState } from '../../context/useAppState'

type Props = {
  data: string
  minCert: number
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
      html += '<br>'
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

const teiToHtml = (
  tei: string,
  minCert: number,
  searchResultHighlight: string | null,
  maskChar: string = '@',
) => {
  const doc = parseXml(tei.trim())

  const body =
    doc.getElementsByTagNameNS('*', 'body')[0] ||
    doc.getElementsByTagNameNS('*', 'text')[0] ||
    doc.documentElement

  const opts = { showPB: true, minCert, maskChar }
  const ps = Array.from(body.getElementsByTagNameNS('*', 'p'))

  const parts = []

  if (ps.length) {
    for (const p of ps) {
      const inner = toReadingHtml(p, opts).trim()
      if (!inner) {
        continue
      }
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

export const Tei = ({ minCert, data }: Props) => {
  const { searchResultHighlight } = useAppState()
  const html = useMemo(
    () => teiToHtml(data, minCert, searchResultHighlight),
    [data, minCert, searchResultHighlight],
  )
  return (
    <pre
      className="font-mono w-fit mt-4 text-xs leading-relaxed border border-gray-300 rounded-xl bg-gray-50 p-2 [&_[data-tei-selected='true']]:bg-yellow-200/70 [&_[data-tei-selected='true']]:text-gray-900 [&_[data-tei-selected='true']]:rounded-sm [&_[data-tei-selected='true']]:px-0.5"
      dangerouslySetInnerHTML={{ __html: html }}
    />
  )
}
