import {
  getBody,
  getDirectChildrenByName,
  getElementAttr,
  getLineElements,
  getTeiTranslationSources,
  getXmlId,
  parseCorrespRefs,
  parseXml,
} from './teiDom.ts'
import { getCertaintyDegreeByTargetId } from './teiDom.ts'
import {
  getOriginalParagraphSpans,
  toReadingTextWithAnchors,
  trimTextWithAnchors,
} from './teiRendering.ts'
import type {
  TeiEditableHighlight,
  TeiHighlightConfig,
  TeiOriginalEditableLine,
  TeiTranslation,
} from './teiTypes.ts'

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
      {
        showPB: true,
        minCert: 0.8,
        maskChar: '@',
        alignLines: false,
        certaintyDegreeByTargetId: getCertaintyDegreeByTargetId(doc),
      },
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
