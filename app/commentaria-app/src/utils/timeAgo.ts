import TimeAgo from 'javascript-time-ago'
import en from 'javascript-time-ago/locale/en'

const TIME_AGO_LOCALE_KEY = '__commentaria_timeago_default_locale__'
const globalWithTimeAgo = globalThis as typeof globalThis & {
  [TIME_AGO_LOCALE_KEY]?: boolean
}

if (!globalWithTimeAgo[TIME_AGO_LOCALE_KEY]) {
  TimeAgo.addDefaultLocale(en)
  globalWithTimeAgo[TIME_AGO_LOCALE_KEY] = true
}

export const timeAgo = new TimeAgo('en-US')
