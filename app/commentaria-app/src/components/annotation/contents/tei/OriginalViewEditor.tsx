import type { EditableOriginalLine } from './TeiPane.types.ts'

type OriginalViewEditorProps = {
  lines: EditableOriginalLine[]
  showViewLabel: boolean
  onChangeLine: (lineId: string, text: string) => void
}

export const OriginalViewEditor = ({
  lines,
  showViewLabel,
  onChangeLine,
}: OriginalViewEditorProps) => (
  <div className="relative">
    {showViewLabel && (
      <div className="absolute top-2 right-2 z-10 rounded bg-white/90 border border-gray-300 px-1.5 py-0.5 text-[10px] font-medium text-gray-700">
        Original
      </div>
    )}
    <div
      className={`text-xs leading-relaxed border border-gray-300 rounded-xl bg-gray-50 p-2 ${showViewLabel ? 'pt-7' : ''} flex flex-col gap-1.5`}
    >
      {lines.map((line) => (
        <input
          key={line.id}
          type="text"
          value={line.text}
          onChange={(event) => onChangeLine(line.id, event.target.value)}
          className="w-full border border-gray-300 rounded-md px-2 py-1 bg-white text-xs text-gray-800 focus:outline-none focus:ring-2 focus:ring-teal-200 focus:border-teal-400"
          spellCheck={false}
        />
      ))}
    </div>
  </div>
)
