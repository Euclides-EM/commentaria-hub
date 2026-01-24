import { Tei } from './Tei.tsx'
import { useAppState } from '../context/useAppState'
import { useEffect, useState } from 'react'
import { useAnnotationTeiQuery } from '../queries/annotations.ts'
import useLocalStorageState from 'use-local-storage-state'

export function TeiPane() {
  const {
    annotation,
    state: { datasetId, annotationId, currentPage },
  } = useAppState()
  const [showTeiSource, setShowTeiSource] = useLocalStorageState(
    'showTeiSource',
    { defaultValue: false },
  )
  const [minCert, setMinCert] = useLocalStorageState('minCert', {
    defaultValue: 0.8,
  })
  const hasTei = !!annotation?.ocred
  const { data } = useAnnotationTeiQuery(
    datasetId,
    annotationId,
    currentPage,
    hasTei,
  )
  const [teiContents, setTeiContents] = useState<string | null>(null)

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setTeiContents(data || '')
  }, [data])

  if (!hasTei) {
    return
  }

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Annotation Details</div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
        <div className="flex gap-2 items-center flex-wrap mb-2.5">
          <button
            className={`px-2.5 py-1.5 border rounded-lg font-semibold text-xs ${
              showTeiSource
                ? 'bg-black text-white border-black'
                : 'border-gray-300 bg-white hover:bg-gray-50'
            }`}
            onClick={() => setShowTeiSource(!showTeiSource)}
          >
            TEI source code
          </button>

          <div
            className="flex items-center gap-2 px-1.5 py-1 border border-gray-200 rounded-lg bg-white"
            title="Hide tokens below certainty threshold"
          >
            <label htmlFor="minCert" className="text-xs opacity-75">
              Min certainty
            </label>
            <input
              type="range"
              min="0.800"
              max="1.000"
              step="0.001"
              className="w-40"
              value={minCert}
              onChange={(e) =>
                setMinCert(Math.round(parseFloat(e.target.value) * 1000) / 1000)
              }
            />
            <input
              type="number"
              min="0.800"
              max="1.000"
              step="0.001"
              className="text-xs opacity-75 font-mono p-1 rounded-lg border border-gray-300"
              value={minCert}
              onChange={(e) => setMinCert(parseFloat(e.target.value))}
            />
          </div>
        </div>

        {teiContents && showTeiSource && (
          <>
            <textarea
              className={`w-full mt-4 h-36 box-border resize-y border border-gray-300 rounded-lg p-2.5 outline-none font-mono text-xs leading-snug ${!showTeiSource ? 'hidden' : ''}`}
              spellCheck={false}
              placeholder="TEI XML..."
              value={teiContents || ''}
              onChange={(e) => setTeiContents(e.target.value)}
            />
          </>
        )}
        {teiContents && <Tei data={teiContents} minCert={minCert} />}
      </div>
    </section>
  )
}
