import type { model_Annotation } from '../api'

interface TeiPaneProps {
  teiStatus: string
  showAnnotationDetails: boolean
  onToggleAnnotationDetails: () => void
  showTeiSource: boolean
  onToggleTeiSource: () => void
  minCert: number
  onMinCertChange: (value: number) => void
  teiInput: string
  onTeiInputChange: (value: string) => void
  onRenderTei: () => void
  annotations: model_Annotation[]
  selectedAnnotationId: string
  renderedTei: string
}

export function TeiPane({
  teiStatus,
  showAnnotationDetails,
  onToggleAnnotationDetails,
  showTeiSource,
  onToggleTeiSource,
  minCert,
  onMinCertChange,
  teiInput,
  onTeiInputChange,
  onRenderTei,
  annotations,
  selectedAnnotationId,
  renderedTei,
}: TeiPaneProps) {
  const escapeHtml = (text: string) => {
    const div = document.createElement('div')
    div.textContent = text
    return div.innerHTML
  }

  const getAnnotationDetails = () => {
    if (!selectedAnnotationId || !annotations.length) {
      return '<div class="text-gray-500 text-sm italic text-center p-5">Select a dataset and an annotation.</div>'
    }

    const annotation = annotations.find(
      (ann) => ann.id === selectedAnnotationId,
    )
    if (!annotation) {
      return '<div class="text-gray-500 text-sm italic text-center p-5">Annotation details not available.</div>'
    }

    const appliedRulesPretty = annotation.applied_rules
      ? JSON.stringify(annotation.applied_rules, null, 2)
      : null

    return `
      <div class="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 items-start">
        <div class="font-semibold text-xs opacity-80 pt-0.5">Name</div><div class="text-sm leading-tight break-all">${escapeHtml(String(annotation.name || ''))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">ID</div><div class="text-sm leading-tight break-all font-mono">${escapeHtml(String(annotation.id || ''))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Dataset</div><div class="text-sm leading-tight break-all font-mono">${escapeHtml(String(annotation.dataset_id || ''))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Pages</div><div class="text-sm leading-tight break-all">${escapeHtml(String(annotation.pages || ''))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Segmented</div><div class="text-sm leading-tight break-all">${escapeHtml(String(!!annotation.segmented))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Ground truth</div><div class="text-sm leading-tight break-all">${escapeHtml(String(!!annotation.ground_truth))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">OCRed</div><div class="text-sm leading-tight break-all">${escapeHtml(String(!!annotation.ocred))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Created</div><div class="text-sm leading-tight break-all font-mono">${escapeHtml(String(annotation.created_at || ''))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Updated</div><div class="text-sm leading-tight break-all font-mono">${escapeHtml(String(annotation.updated_at || ''))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Description</div><div class="text-sm leading-tight break-all">${escapeHtml(String(annotation.description || ''))}</div>
        <div class="font-semibold text-xs opacity-80 pt-0.5">Applied rules</div>
        <div class="text-sm leading-tight">
          ${appliedRulesPretty ? `<pre class="m-0 p-2.5 rounded bg-gray-100 text-xs leading-normal overflow-auto font-mono">${escapeHtml(appliedRulesPretty)}</pre>` : `<span class="text-gray-500">None</span>`}
        </div>
      </div>
    `
  }

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>TEI</div>
        <div className="text-xs opacity-75">{teiStatus}</div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
        <div className="flex gap-2 items-center flex-wrap mb-2.5">
          <button
            className={`px-2.5 py-1.5 border rounded-lg font-semibold text-xs ${
              showAnnotationDetails
                ? 'bg-black text-white border-black'
                : 'border-gray-300 bg-white hover:bg-gray-50'
            }`}
            onClick={onToggleAnnotationDetails}
          >
            Annotation details
          </button>

          <button
            className={`px-2.5 py-1.5 border rounded-lg font-semibold text-xs ${
              showTeiSource
                ? 'bg-black text-white border-black'
                : 'border-gray-300 bg-white hover:bg-gray-50'
            }`}
            onClick={onToggleTeiSource}
          >
            TEI source code
          </button>

          <div className={`${!showTeiSource ? 'hidden' : ''}`}>
            <button
              className="px-2.5 py-1.5 border border-gray-300 rounded-lg bg-white hover:bg-gray-50 font-semibold text-xs"
              onClick={onRenderTei}
            >
              Render
            </button>
          </div>

          <div
            className="flex items-center gap-2 px-1.5 py-1 border border-gray-200 rounded-lg bg-white"
            title="Hide tokens below certainty threshold"
          >
            <label htmlFor="minCert" className="text-xs opacity-75">
              Min certainty
            </label>
            <input
              id="minCert"
              type="range"
              min="0.800"
              max="1.000"
              step="0.001"
              className="w-40"
              value={minCert}
              onChange={(e) => onMinCertChange(parseFloat(e.target.value))}
            />
            <span className="text-xs opacity-75 font-mono">
              {minCert.toFixed(3)}
            </span>
          </div>
        </div>

        <textarea
          className={`w-full h-36 box-border resize-y border border-gray-300 rounded-lg p-2.5 outline-none font-mono text-xs leading-snug ${!showTeiSource ? 'hidden' : ''}`}
          spellCheck={false}
          placeholder="TEI XML…"
          value={teiInput}
          onChange={(e) => onTeiInputChange(e.target.value)}
        />

        <div
          className={`mt-2.5 border border-gray-200 rounded-lg bg-gray-50 p-3.5 overflow-auto leading-normal text-base box-border ${!showAnnotationDetails ? 'hidden' : ''}`}
          dangerouslySetInnerHTML={{ __html: getAnnotationDetails() }}
        />

        <div
          className={`mt-2.5 border border-gray-200 rounded-lg bg-gray-50 p-3.5 overflow-auto leading-normal text-base box-border ${showAnnotationDetails ? 'hidden' : ''}`}
          dangerouslySetInnerHTML={{ __html: renderedTei }}
        />
      </div>
    </section>
  )
}
