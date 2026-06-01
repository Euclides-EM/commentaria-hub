import {
  getElementAttr,
  getXmlId,
  parseAnaRefs,
  parseCorrespRefs,
  parseXml,
  toUniqueSorted,
} from './teiDom.ts'
import type { TeiSurfaceZone } from './teiTypes.ts'

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

export const getTeiAllZoneCategories = (tei: string): string[] => {
  try {
    const doc = parseXml(tei.trim())
    const interpGroups = doc.getElementsByTagNameNS('*', 'interpGrp')
    for (let i = 0; i < interpGroups.length; i++) {
      const group = interpGroups[i]
      if (getElementAttr(group, 'type') !== 'zone_categories') continue
      const interps = group.getElementsByTagNameNS('*', 'interp')
      const cats: string[] = []
      for (let j = 0; j < interps.length; j++) {
        const id = getXmlId(interps[j])
        if (id) cats.push(id)
      }
      return cats
    }
    return []
  } catch {
    return []
  }
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
      const anaRefs = parseAnaRefs(zone.getAttribute('ana'))
      const zoneCategory = anaRefs[0] || ''
      parsed.push({
        id,
        matchIds: matchIds.length ? matchIds : [id],
        zoneCategory,
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
        zoneCategory: zone.zoneCategory,
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
