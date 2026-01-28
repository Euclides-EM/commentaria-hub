import { timeAgo } from '../../utils/timeAgo'

type TimestampProps = {
  date: string | undefined
  hideFullDate?: boolean
}

export const Timestamp = ({ date, hideFullDate = false }: TimestampProps) => {
  if (!date) {
    return 'N/A'
  }

  const parsed = new Date(date)

  const timeAgoFormatted = timeAgo.format(parsed)
  const fullDateFormatted = parsed.toLocaleString()
  return (
    <div className="flex gap-2 items-center">
      <span title={hideFullDate ? parsed.toLocaleString() : timeAgoFormatted}>
        {timeAgoFormatted}
      </span>
      {!hideFullDate && (
        <span className="text-xs text-gray-500">{fullDateFormatted}</span>
      )}
    </div>
  )
}
