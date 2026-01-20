type Props = {
  data: string
}
const escapeHtml = (text: string) => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

export const Tei = ({ data }: Props) =>
  data ? (
    <pre className="font-mono text-sm leading-relaxed">${escapeHtml(data)}</pre>
  ) : (
    <div className="text-gray-500 text-sm italic text-center p-5">
      Click Load.
    </div>
  )
