import { Tei } from './Tei.tsx'
import { useAppState } from '../../../context/useAppState.ts'
import { useEffect, useState } from 'react'
import {
  useAnnotationTeiQuery,
  useEditionTeiQuery,
} from '../../../queries/annotations.ts'
import useLocalStorageState from 'use-local-storage-state'
import { RangeInput } from '../../core/RangeInput.tsx'

export function TeiPane() {
  const {
    annotation,
    dataset,
    state: { datasetId, annotationId, currentPageOrKey },
  } = useAppState()
  const [showTeiSource, setShowTeiSource] = useLocalStorageState(
    'showTeiSource',
    { defaultValue: false },
  )
  const [minCert, setMinCert] = useLocalStorageState('minCert', {
    defaultValue: 0.8,
  })
  const ocred = !!annotation?.ocred
  const editionId = dataset?.edition_id

  const editionTeiProbe = useEditionTeiQuery(
    editionId,
    currentPageOrKey,
    !!editionId,
  )
  const editionTeiAvailable =
    !!editionId &&
    editionTeiProbe.isSuccess &&
    !!editionTeiProbe.data

  const defaultUseEditionTei = !ocred && editionTeiAvailable
  const [storedTeiSource, setStoredTeiSource] = useLocalStorageState<
    'annotation' | 'edition' | null
  >('teiSource', { defaultValue: null })
  const effectiveUseEditionTei =
    storedTeiSource === 'edition' ||
    (storedTeiSource === null && defaultUseEditionTei)

  const annotationTeiQuery = useAnnotationTeiQuery(
    datasetId,
    annotationId,
    currentPageOrKey,
    !effectiveUseEditionTei,
  )

  const data = effectiveUseEditionTei
    ? editionTeiProbe.data
    : annotationTeiQuery.data

  const [teiContents, setTeiContents] = useState<string | null>(null)

  useEffect(() => {
    setTeiContents(data ?? null)
  }, [data])

  const hasTei = ocred || editionTeiAvailable
  const showPane = ocred || !!editionId

  if (!showPane) {
    return null
  }

  const canToggleSource = (ocred || editionTeiAvailable) && !!editionId
  const isLoading =
    (effectiveUseEditionTei && editionTeiProbe.isFetching) ||
    (!effectiveUseEditionTei && annotationTeiQuery.isFetching)
  const error =
    (effectiveUseEditionTei && editionTeiProbe.error) ||
    (!effectiveUseEditionTei && annotationTeiQuery.error)

  console.log('canToggleSource', canToggleSource, 'ocred', ocred, 'editionTeiAvailable', editionTeiAvailable, 'editionId', editionId)
  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Annotation Details</div>
      </div>

      <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
        <div className="flex gap-2 items-center flex-wrap mb-2.5">

            <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
              <input
                type="checkbox"
                checked={effectiveUseEditionTei}
                onChange={(e) =>
                  setStoredTeiSource(e.target.checked ? 'edition' : 'annotation')
                }
                className="rounded border-gray-300"
              />
              <span>Use edition TEI</span>
            </label>
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

          <RangeInput
            label="Min certainty"
            value={minCert}
            min={0.8}
            max={1}
            step={0.001}
            title="Hide tokens below certainty threshold"
            onChange={(value) => setMinCert(Math.round(value * 1000) / 1000)}
          />
        </div>

        {isLoading && !teiContents && (
          <p className="text-gray-500 text-sm py-2">Loading TEI…</p>
        )}
        {error && !teiContents && (
          <p className="text-red-600 text-sm py-2">
            Failed to load TEI. Try switching source.
          </p>
        )}
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
