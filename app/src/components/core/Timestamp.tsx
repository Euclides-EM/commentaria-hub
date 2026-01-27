import { timeAgo } from '../../utils/timeAgo'

type TimestampProps = {
  date: string | undefined
}

export const Timestamp = ({ date }: TimestampProps) => {
  if (!date) {
    return 'N/A'
  }

  const parsed = new Date(date)

  return (
    <div className="flex gap-2 items-center">
      <span>{timeAgo.format(parsed)}</span>
      <span className="text-xs text-gray-500">{parsed.toLocaleString()}</span>
    </div>
  )
}
