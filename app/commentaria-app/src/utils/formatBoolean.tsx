export const formatBoolean = (value: boolean | null | undefined) =>
  value ? <span className="text-teal-700 font-semibold">{'\u2713'}</span> : '-'
