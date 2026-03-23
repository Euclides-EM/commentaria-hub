import type { common_ALTOPart } from '@hub-api'

export const ANNOTATION_SEARCH_TERM_KEY = 'annotationSearchTerm'
export const ANNOTATION_SEARCH_CATEGORIES_KEY = 'annotationSearchCategories'
export const ANNOTATION_SEARCH_WITHIN_KEY = 'annotationSearchWithin'

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
