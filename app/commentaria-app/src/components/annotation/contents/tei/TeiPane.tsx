import {
  getTeiTranslations,
  teiToHtml,
  type TeiTranslation,
  type TeiViewMode,
} from './tei.ts'
import { useAppState } from '../../../../context/useAppState.ts'
import { useEffect, useMemo, useState } from 'react'
import {
  useAnnotationTeiQuery,
  useEditionTeiQuery,
} from '../../../../queries/annotations.ts'
import useLocalStorageState from 'use-local-storage-state'
import { RangeInput } from '../../../core/RangeInput.tsx'
import Select from 'react-select'
import { selectStyles } from '../../../../styles/selectStyles.ts'
import { MultiSelectDropdown } from '../../../core/MultiSelectDropdown.tsx'

const VIEW_LABEL_MAP: Record<string, string> = {
  modern_en: 'English',
}

const normalizeTeiViewModes = (
  viewModes: TeiViewMode[],
  allowedViewModes: TeiViewMode[],
): TeiViewMode[] => {
  if (!allowedViewModes.length) {
    return ['original']
  }
  const allowed = new Set(allowedViewModes)
  const next = viewModes.filter((mode) => allowed.has(mode))
  if (!next.length) {
    return ['original']
  }
  return next
}

type Props = {
  data: string
  minCert: number
  viewMode: TeiViewMode
  viewLabel: string
  alignLines: boolean
}

const Tei = ({ minCert, data, viewMode, viewLabel, alignLines }: Props) => {
  const { searchResultHighlight } = useAppState()
  const html = useMemo(
    () =>
      teiToHtml(
        data,
        minCert,
        searchResultHighlight,
        '@',
        viewMode,
        alignLines,
      ),
    [alignLines, data, minCert, searchResultHighlight, viewMode],
  )
  return (
    <div className="relative">
      <div className="absolute top-2 right-2 z-10 rounded bg-white/90 border border-gray-300 px-1.5 py-0.5 text-[10px] font-medium text-gray-700">
        {viewLabel}
      </div>
      <div
        className="text-xs leading-relaxed border border-gray-300 rounded-xl bg-gray-50 p-2 pt-7 [&_p]:mb-2 [&_p:last-child]:mb-0 [&_[data-tei-selected='true']]:bg-yellow-200/70 [&_[data-tei-selected='true']]:text-gray-900 [&_[data-tei-selected='true']]:rounded-sm [&_[data-tei-selected='true']]:px-0.5"
        style={{ whiteSpace: 'normal' }}
        dangerouslySetInnerHTML={{ __html: html }}
      />
    </div>
  )
}

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
  const [alignLines, setAlignLines] = useLocalStorageState('alignTeiLines', {
    defaultValue: false,
  })
  const ocred = !!annotation?.ocred
  const editionId = dataset?.edition_id

  const availableSources = useMemo(() => {
    const sources: Array<'annotation' | 'edition'> = []
    if (ocred) {
      sources.push('annotation')
    }
    if (editionId) {
      sources.push('edition')
    }
    return sources
  }, [editionId, ocred])

  const [storedTeiSource, setStoredTeiSource] = useLocalStorageState<
    'annotation' | 'edition'
  >('teiSource', { defaultValue: 'annotation' })
  const effectiveTeiSource = availableSources.includes(storedTeiSource)
    ? storedTeiSource
    : availableSources[0]

  useEffect(() => {
    if (effectiveTeiSource && effectiveTeiSource !== storedTeiSource) {
      setStoredTeiSource(effectiveTeiSource)
    }
  }, [effectiveTeiSource, setStoredTeiSource, storedTeiSource])

  const editionTeiQuery = useEditionTeiQuery(
    editionId,
    currentPageOrKey,
    !!editionId && effectiveTeiSource === 'edition',
  )
  const annotationTeiQuery = useAnnotationTeiQuery(
    datasetId,
    annotationId,
    currentPageOrKey,
    effectiveTeiSource === 'annotation',
  )

  const data =
    effectiveTeiSource === 'edition'
      ? editionTeiQuery.data
      : annotationTeiQuery.data

  const [teiContents, setTeiContents] = useState<string | null>(null)
  const [teiViewModes, setTeiViewModes] = useLocalStorageState<TeiViewMode[]>(
    'teiViewModes',
    { defaultValue: ['original'] },
  )

  useEffect(() => {
    setTeiContents(data ?? null)
  }, [data])

  const teiTranslations = useMemo<TeiTranslation[]>(
    () => (teiContents ? getTeiTranslations(teiContents) : []),
    [teiContents],
  )

  const availableViewModes = useMemo<TeiViewMode[]>(
    () => ['original', ...teiTranslations.map((translation) => translation.id)],
    [teiTranslations],
  )

  useEffect(() => {
    if (teiContents == null) {
      return
    }
    const next = normalizeTeiViewModes(teiViewModes, availableViewModes)
    if (
      next.length === teiViewModes.length &&
      next.every((mode, index) => mode === teiViewModes[index])
    ) {
      return
    }
    setTeiViewModes(next)
  }, [availableViewModes, setTeiViewModes, teiContents, teiViewModes])

  const orderedSelectedViewModes = useMemo(() => {
    const selected = new Set(
      normalizeTeiViewModes(teiViewModes, availableViewModes),
    )
    return availableViewModes.filter((mode) => selected.has(mode))
  }, [availableViewModes, teiViewModes])

  const showPane = ocred || !!editionId

  if (!showPane) {
    return null
  }

  const sourceOptions = availableSources.map((source) => ({
    value: source,
    label: source === 'edition' ? 'Edition' : 'Annotation',
  }))
  const selectedSourceOption =
    sourceOptions.find((option) => option.value === effectiveTeiSource) || null
  const getViewModeLabel = (mode: TeiViewMode) => {
    if (mode === 'original') {
      return 'Original'
    }
    const rawLabel =
      teiTranslations.find((translation) => translation.id === mode)?.label ||
      mode
    return VIEW_LABEL_MAP[rawLabel] || rawLabel
  }

  const isLoading =
    (effectiveTeiSource === 'edition' && editionTeiQuery.isFetching) ||
    (effectiveTeiSource === 'annotation' && annotationTeiQuery.isFetching)
  const error =
    (effectiveTeiSource === 'edition' && editionTeiQuery.error) ||
    (effectiveTeiSource === 'annotation' && annotationTeiQuery.error)

  return (
    <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 h-full bg-white">
      <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
        <div>Contents</div>
      </div>

      <div className="flex-1 min-h-0 overflow-hidden p-2.5 box-border flex flex-col">
        <div className="flex gap-2 items-center flex-wrap mb-2.5">
          {sourceOptions.length > 1 && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs font-medium text-gray-600">Source:</span>
              <div className="w-36">
                <Select
                  value={selectedSourceOption}
                  onChange={(
                    option: {
                      value: 'annotation' | 'edition'
                      label: string
                    } | null,
                  ) => {
                    if (option) {
                      setStoredTeiSource(option.value)
                    }
                  }}
                  options={sourceOptions}
                  isClearable={false}
                  styles={selectStyles<{
                    value: 'annotation' | 'edition'
                    label: string
                  }>()}
                  menuPortalTarget={document.body}
                  menuPosition="fixed"
                />
              </div>
            </div>
          )}
          {availableViewModes.length > 1 && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs font-medium text-gray-600">Views:</span>
              <MultiSelectDropdown<TeiViewMode>
                allItems={availableViewModes}
                selectedItems={orderedSelectedViewModes}
                setSelectedItems={(items) => {
                  if (!items || items.length === 0) {
                    setTeiViewModes(['original'])
                    return
                  }
                  setTeiViewModes(items)
                }}
                itemsLabel="views"
                getItemLabel={getViewModeLabel}
                minWidth="180px"
              />
            </div>
          )}
          <label className="flex items-center gap-1.5 cursor-pointer text-xs font-medium">
            <input
              type="checkbox"
              checked={alignLines}
              onChange={(e) => setAlignLines(e.target.checked)}
              className="rounded border-gray-300"
            />
            <span>Align lines</span>
          </label>
          <RangeInput
            label="Min certainty"
            value={minCert}
            min={0.8}
            max={1}
            step={0.001}
            title="Hide tokens below certainty threshold"
            onChange={(value) => setMinCert(Math.round(value * 1000) / 1000)}
          />
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
        </div>

        {isLoading && !teiContents && (
          <p className="text-gray-500 text-sm py-2">Loading TEI…</p>
        )}
        {effectiveTeiSource === 'annotation' &&
          !teiContents &&
          !annotationTeiQuery.isFetching &&
          !annotationTeiQuery.error &&
          (!datasetId || !annotationId) && (
            <p className="text-amber-700 text-sm py-2">
              Select a dataset and an annotation to view annotation TEI.
            </p>
          )}
        {error && !teiContents && (
          <p className="text-red-600 text-sm py-2">
            {effectiveTeiSource === 'edition' &&
            (error as Error)?.message?.includes('404')
              ? 'Edition TEI is not available for this page. Use annotation TEI or another source.'
              : 'Failed to load TEI. Try switching source.'}
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
        {teiContents && (
          <div className="mt-4 flex-1 min-h-0 overflow-y-auto">
            <div className="flex flex-wrap gap-3">
              {orderedSelectedViewModes.map((viewMode) => (
                <div key={viewMode} className="min-w-105 basis-105 flex-1">
                  <Tei
                    data={teiContents}
                    minCert={minCert}
                    viewMode={viewMode}
                    viewLabel={getViewModeLabel(viewMode)}
                    alignLines={alignLines}
                  />
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </section>
  )
}
