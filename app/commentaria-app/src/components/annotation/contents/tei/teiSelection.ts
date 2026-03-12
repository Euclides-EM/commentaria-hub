import type { TeiParagraphSelection } from './teiTypes.ts'

const getOffsetInParagraph = (
  paragraph: Element,
  node: Node,
  offset: number,
): number | null => {
  try {
    const range = document.createRange()
    range.setStart(paragraph, 0)
    range.setEnd(node, offset)
    return range.toString().length
  } catch {
    return null
  }
}

const getCanonicalParagraphText = (paragraph: Element) => {
  const encoded = paragraph.getAttribute('data-tei-paragraph-text') || ''
  if (!encoded) {
    return paragraph.textContent || ''
  }
  try {
    return decodeURIComponent(encoded)
  } catch {
    return paragraph.textContent || ''
  }
}

const findClosestSubstringIndex = (
  text: string,
  query: string,
  approxStart: number,
) => {
  if (!query) {
    return null
  }
  let bestIndex = -1
  let bestDistance = Number.POSITIVE_INFINITY
  let searchFrom = 0
  while (searchFrom <= text.length) {
    const index = text.indexOf(query, searchFrom)
    if (index < 0) {
      break
    }
    const distance = Math.abs(index - approxStart)
    if (distance < bestDistance) {
      bestIndex = index
      bestDistance = distance
    }
    searchFrom = index + 1
  }
  return bestIndex >= 0 ? bestIndex : null
}

const normalizeWhitespaceWithMap = (value: string) => {
  let normalized = ''
  const map: number[] = []
  let previousWasWhitespace = false

  for (let index = 0; index < value.length; index++) {
    const char = value[index]
    if (/\s/.test(char)) {
      if (!previousWasWhitespace) {
        normalized += ' '
        map.push(index)
        previousWasWhitespace = true
      }
      continue
    }
    normalized += char
    map.push(index)
    previousWasWhitespace = false
  }

  return { normalized, map }
}

const resolveSelectionOffsets = (
  canonicalText: string,
  selectedText: string,
  startOffset: number,
  endOffset: number,
) => {
  const exactStart = findClosestSubstringIndex(
    canonicalText,
    selectedText,
    startOffset,
  )
  if (exactStart != null) {
    return {
      start: exactStart,
      end: exactStart + selectedText.length,
      surface: canonicalText.slice(
        exactStart,
        exactStart + selectedText.length,
      ),
    }
  }

  const normalizedSelected = normalizeWhitespaceWithMap(selectedText)
  const normalizedCanonical = normalizeWhitespaceWithMap(canonicalText)
  const normalizedApproxStart = normalizeWhitespaceWithMap(
    canonicalText.slice(0, startOffset),
  ).normalized.length
  const normalizedStart = findClosestSubstringIndex(
    normalizedCanonical.normalized,
    normalizedSelected.normalized,
    normalizedApproxStart,
  )
  if (
    normalizedStart != null &&
    normalizedSelected.map.length > 0 &&
    normalizedCanonical.map.length > 0
  ) {
    const canonicalStart = normalizedCanonical.map[normalizedStart]
    const normalizedEnd = normalizedStart + normalizedSelected.normalized.length
    const endMapIndex = Math.max(0, normalizedEnd - 1)
    const canonicalEnd =
      (normalizedCanonical.map[endMapIndex] ?? canonicalStart) + 1
    return {
      start: canonicalStart,
      end: canonicalEnd,
      surface: canonicalText.slice(canonicalStart, canonicalEnd),
    }
  }

  const start = Math.max(0, Math.min(canonicalText.length, startOffset))
  const end = Math.max(start, Math.min(canonicalText.length, endOffset))
  return {
    start,
    end,
    surface: canonicalText.slice(start, end),
  }
}

export const getTeiParagraphSelection = (
  paragraph: Element,
  range: Range,
): TeiParagraphSelection | null => {
  const startOffset = getOffsetInParagraph(
    paragraph,
    range.startContainer,
    range.startOffset,
  )
  const endOffset = getOffsetInParagraph(
    paragraph,
    range.endContainer,
    range.endOffset,
  )
  if (startOffset == null || endOffset == null) {
    return null
  }

  const domStart = Math.min(startOffset, endOffset)
  const domEnd = Math.max(startOffset, endOffset)
  const canonicalParagraphText = getCanonicalParagraphText(paragraph)
  const selectedText = range.toString()
  const { start, end, surface } = resolveSelectionOffsets(
    canonicalParagraphText,
    selectedText,
    domStart,
    domEnd,
  )
  if (end <= start || !surface.trim()) {
    return null
  }

  return { start, end, surface }
}
