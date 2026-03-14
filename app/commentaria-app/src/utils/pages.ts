export const expandRange = (range: string): string[] => {
  const parts = range.trim().split('-')

  if (parts.length !== 2) {
    return [range]
  }

  const min = parseInt(parts[0].trim(), 10)
  const max = parseInt(parts[1].trim(), 10)

  if (Number.isNaN(min) || Number.isNaN(max)) {
    return [range]
  }

  return Array.from({ length: Math.max(0, max - min + 1) }, (_, i) =>
    String(min + i),
  )
}

/**
 * Count pages from a comma-separated string that may contain ranges (e.g. "4-9,7,78").
 * Ranges are inclusive (4-9 = 6 pages).
 */
export function countPages(pagesStr: string): number {
  const parts = pagesStr
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
  let total = 0
  for (const part of parts) {
    const dash = part.indexOf('-')
    if (dash >= 0) {
      const start = parseInt(part.slice(0, dash), 10)
      const end = parseInt(part.slice(dash + 1), 10)
      if (!Number.isNaN(start) && !Number.isNaN(end) && end >= start) {
        total += end - start + 1
      }
    } else {
      const n = parseInt(part, 10)
      if (!Number.isNaN(n)) total += 1
    }
  }
  return total
}
