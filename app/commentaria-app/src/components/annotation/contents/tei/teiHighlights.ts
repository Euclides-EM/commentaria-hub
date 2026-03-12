import {
  getElementAttr,
  getXmlId,
  parseAnaRefs,
  parseCorrespRefs,
  parseXml,
} from './teiDom.ts'
import type { TeiHighlightCategory, TeiHighlightSpan } from './teiTypes.ts'

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

export const isVerbCategory = (categoryId: string, categoryLabel: string) => {
  const normalized = `${categoryId} ${categoryLabel}`
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '')
  return normalized.includes('verb')
}

export const getTeiHighlightSpans = (doc: Document): TeiHighlightSpan[] => {
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
