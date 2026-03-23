import type { common_ALTOPart } from '@hub-api'

export const ANNOTATION_SEARCH_TERM_KEY = 'annotationSearchTerm'
export const ANNOTATION_SEARCH_CATEGORIES_KEY = 'annotationSearchCategories'
export const ANNOTATION_SEARCH_WITHIN_KEY = 'annotationSearchWithin'

export const buildSearchSnippet = (content: string, maxLength = 64) => {
  const startMatch = content.match(/<em[^>]*>/i)
  if (!startMatch) {
    if (content.length <= maxLength) return content
    return `${content.slice(0, maxLength)}...`
  }

  const startTag = startMatch[0]
  const startIndex = content.indexOf(startTag)
  const endIndex = content.toLowerCase().indexOf('</em>', startIndex)
  if (endIndex === -1) return content

  const beforeText = content.slice(0, startIndex)
  const matchHtml = content.slice(startIndex, endIndex + 5)
  const afterText = content.slice(endIndex + 5)
  const matchText = matchHtml.replace(/<[^>]*>/g, '')
  const remaining = maxLength - matchText.length

  if (remaining <= 0) {
    const truncated = matchText.slice(0, maxLength)
    const suffix = matchText.length > maxLength ? '...' : ''
    return `${startTag}${truncated}${suffix}</em>`
  }

  const beforeLen = Math.floor(remaining / 2)
  const afterLen = remaining - beforeLen
  const beforeTrim =
    beforeText.length > beforeLen
      ? beforeText.slice(beforeText.length - beforeLen)
      : beforeText
  const afterTrim =
    afterText.length > afterLen ? afterText.slice(0, afterLen) : afterText
  const prefix = beforeText.length > beforeLen ? '...' : ''
  const suffix = afterText.length > afterLen ? '...' : ''

  return `${prefix}${beforeTrim}${matchHtml}${afterTrim}${suffix}`
}

export const getSearchResultPageOrKey = (
  result: common_ALTOPart,
): string | null => {
  if (result.location?.page === '0' && result.location.text_block_id) {
    return result.location.text_block_id
  }
  if (result.location?.page) {
    return String(result.location.page)
  }
  return null
}
