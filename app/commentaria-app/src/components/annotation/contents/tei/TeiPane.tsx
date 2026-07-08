import {
  getTeiAllZoneCategories,
  getTeiHighlightCategories,
  getTeiOriginalEditableLines,
  getTeiSurfaceZones,
  getTeiTranslations,
  getTeiZoneToServerTextBlockId,
  hasTeiCertaintyDegrees,
  type TeiHighlightConfig,
  type TeiManualHighlight,
  type TeiSurfaceZone,
  type TeiTranslation,
  type TeiViewMode,
} from './tei.ts'
import { useAppState } from '../../../../context/useAppState.ts'
import { useEffect, useMemo, useState } from 'react'
import {
  annotationTeiQueryKey,
  editionTeiQueryKey,
  useAnnotationTeiQuery,
  useEditionTeiQuery,
} from '../../../../queries/annotations.ts'
import { useDatasetFeaturesQuery } from '../../../../queries/datasets.ts'
import useLocalStorageState from 'use-local-storage-state'
import Select from 'react-select'
import { selectStyles } from '../../../../styles/selectStyles.ts'
import { MultiSelectDropdown } from '../../../core/MultiSelectDropdown.tsx'
import {
  type annotationrule_TextBlockCorrections,
  AnnotationsApplyRulesService,
  type feature_Result,
  FeatureResultsService,
} from '@hub-api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useAuthStore } from '../../../../store/authStore.ts'
import { TITLE_PAGES_DATASET_ID } from '../../../../utils/editions.ts'
import { FeatureHighlightModal } from './FeatureHighlightModal.tsx'
import { OriginalViewEditor } from './OriginalViewEditor.tsx'
import { TeiContentView } from './TeiContentView.tsx'
import { TeiDisplayControls } from './TeiDisplayControls.tsx'
import type {
  DraftHighlight,
  EditableOriginalLine,
  FeatureModalState,
  FeatureOption,
  ResolvedTeiFeature,
  SelectionDraft,
  SourceOption,
  TeiTooltipItem,
} from './TeiPane.types.ts'
import {
  VIEW_LABEL_MAP,
  getComparableValues,
  groupByFeature,
  hasTeiPositionProperties,
  isVerbLike,
  matchTeiCategoryToFeature,
  normalizeTeiViewModes,
  removeHighlightFromDrafts,
  sameStringArray,
  toDraftHighlightsFromResults,
  toFeatureOptions,
  toResultValues,
} from './teiPaneUtils.tsx'

type TeiPaneProps = {
  activeLineMatchIds: string[]
  onHoverLineMatchIds: (ids: string[]) => void
  onSurfaceZonesChange: (zones: TeiSurfaceZone[]) => void
  onAllZoneCategoriesChange: (categories: string[]) => void
}

export function TeiPane({
  activeLineMatchIds,
  onHoverLineMatchIds,
  onSurfaceZonesChange,
  onAllZoneCategoriesChange,
}: TeiPaneProps) {
  const {
    annotation,
    dataset,
    state: { datasetId, annotationId, currentPageOrKey },
  } = useAppState()
  const queryClient = useQueryClient()
  const isAuthenticated = !!useAuthStore((store) => store.token)

  const [minCert, setMinCert] = useLocalStorageState('minCert', {
    defaultValue: 0.8,
    storageSync: false,
  })
  const [showTeiLineHighlights, setShowTeiLineHighlights] =
    useLocalStorageState('showTeiLineHighlights', { defaultValue: true })
  const [alignLines, setAlignLines] = useLocalStorageState('alignTeiLines', {
    defaultValue: false,
    storageSync: false,
  })
  const [showCertaintyVisualization, setShowCertaintyVisualization] =
    useLocalStorageState('showTeiCertaintyVisualization', {
      defaultValue: false,
    })
  const [isFeatureSelectExpanded, setIsFeatureSelectExpanded] =
    useLocalStorageState('teiFeatureSelectExpanded', {
      defaultValue: false,
      storageSync: false,
    })
  const [featureModalState, setFeatureModalState] =
    useState<FeatureModalState | null>(null)
  const [modalFeatureId, setModalFeatureId] = useState<string>('')
  const [saveError, setSaveError] = useState<string | null>(null)
  const [isTextEditMode, setIsTextEditMode] = useState(false)
  const [textEditError, setTextEditError] = useState<string | null>(null)
  const [textEditSavePending, setTextEditSavePending] = useState(false)
  const [editableOriginalLines, setEditableOriginalLines] = useState<
    EditableOriginalLine[]
  >([])
  const [baselineHighlights, setBaselineHighlights] = useState<
    DraftHighlight[]
  >([])
  const [draftHighlights, setDraftHighlights] = useState<DraftHighlight[]>([])
  const [removedTeiHighlightIds, setRemovedTeiHighlightIds] = useState<
    string[]
  >([])
  const [forcedChangedFeatureIds, setForcedChangedFeatureIds] = useState<
    string[]
  >([])
  const segmented =
    !!annotation?.segmented || datasetId === TITLE_PAGES_DATASET_ID
  const editionId = dataset?.edition_id

  const candidateSources = useMemo(() => {
    const sources: Array<'annotation' | 'edition'> = []
    if (segmented) {
      sources.push('annotation')
    }
    if (editionId) {
      sources.push('edition')
    }
    return sources
  }, [editionId, segmented])

  const [storedTeiSource, setStoredTeiSource] = useLocalStorageState<
    'annotation' | 'edition'
  >('teiSource', { defaultValue: 'annotation', storageSync: false })
  const preferredTeiSource = candidateSources.includes(storedTeiSource)
    ? storedTeiSource
    : candidateSources[0]

  const editionTeiQuery = useEditionTeiQuery(
    editionId,
    currentPageOrKey,
    !!editionId,
  )
  const annotationTeiQuery = useAnnotationTeiQuery(
    datasetId,
    annotationId,
    currentPageOrKey,
    true,
    'preview',
  )
  const featuresQuery = useDatasetFeaturesQuery(datasetId, !!datasetId)
  const featureResultsQuery = useQuery({
    queryKey: ['featureResults', datasetId, annotationId, currentPageOrKey],
    queryFn: () =>
      FeatureResultsService.getDatasetsAnnotationsResults({
        dataSetId: datasetId,
        id: annotationId,
        keys: [String(currentPageOrKey)],
        fallbackToOrigin: true,
      }),
    enabled: !!datasetId && !!annotationId,
  })

  const availableSources = useMemo(() => {
    const sources: Array<'annotation' | 'edition'> = []
    if (segmented && annotationTeiQuery.isSuccess && annotationTeiQuery.data) {
      sources.push('annotation')
    }
    if (editionId && editionTeiQuery.isSuccess && editionTeiQuery.data) {
      sources.push('edition')
    }
    return sources
  }, [
    editionId,
    segmented,
    annotationTeiQuery.data,
    annotationTeiQuery.isSuccess,
    editionTeiQuery.data,
    editionTeiQuery.isSuccess,
  ])

  const effectiveTeiSource = availableSources.includes(storedTeiSource)
    ? storedTeiSource
    : availableSources[0] || preferredTeiSource

  useEffect(() => {
    if (
      availableSources.length > 0 &&
      effectiveTeiSource &&
      effectiveTeiSource !== storedTeiSource
    ) {
      setStoredTeiSource(effectiveTeiSource)
    }
  }, [
    availableSources,
    effectiveTeiSource,
    setStoredTeiSource,
    storedTeiSource,
  ])

  const data =
    effectiveTeiSource === 'edition'
      ? editionTeiQuery.data
      : effectiveTeiSource === 'annotation'
        ? annotationTeiQuery.data
        : undefined

  const teiContents = data ?? null
  const baseEditableOriginalLines = useMemo(
    () =>
      teiContents
        ? getTeiOriginalEditableLines(teiContents).map((line, index) => ({
            ...line,
            id: `${line.id}:${String(index)}`,
            originalText: line.text,
          }))
        : [],
    [teiContents],
  )
  const zoneToServerTextBlockId = useMemo(
    () => (teiContents ? getTeiZoneToServerTextBlockId(teiContents) : {}),
    [teiContents],
  )
  const surfaceZoneTeiContents =
    teiContents ?? annotationTeiQuery.data ?? editionTeiQuery.data ?? null
  const [teiViewModes, setTeiViewModes] = useLocalStorageState<TeiViewMode[]>(
    'teiViewModes',
    { defaultValue: [] },
  )

  const teiTranslations = useMemo<TeiTranslation[]>(
    () => (teiContents ? getTeiTranslations(teiContents) : []),
    [teiContents],
  )
  const showMinCertControl = useMemo(
    () => (teiContents ? hasTeiCertaintyDegrees(teiContents) : false),
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

  useEffect(() => {
    onSurfaceZonesChange(
      surfaceZoneTeiContents ? getTeiSurfaceZones(surfaceZoneTeiContents) : [],
    )
  }, [currentPageOrKey, onSurfaceZonesChange, surfaceZoneTeiContents])

  useEffect(
    () => () => {
      onSurfaceZonesChange([])
    },
    [onSurfaceZonesChange],
  )

  useEffect(() => {
    onAllZoneCategoriesChange(
      surfaceZoneTeiContents
        ? getTeiAllZoneCategories(surfaceZoneTeiContents)
        : [],
    )
  }, [currentPageOrKey, onAllZoneCategoriesChange, surfaceZoneTeiContents])

  useEffect(
    () => () => {
      onAllZoneCategoriesChange([])
    },
    [onAllZoneCategoriesChange],
  )

  useEffect(() => {
    setEditableOriginalLines(baseEditableOriginalLines)
  }, [baseEditableOriginalLines])

  useEffect(() => {
    setIsTextEditMode(false)
    setTextEditError(null)
    setTextEditSavePending(false)
  }, [annotationId, currentPageOrKey, datasetId, effectiveTeiSource])

  const teiCategories = useMemo(
    () => (teiContents ? getTeiHighlightCategories(teiContents) : []),
    [teiContents],
  )

  const datasetFeatures = useMemo(
    () => featuresQuery.data ?? [],
    [featuresQuery.data],
  )

  const categoryToFeatureId = useMemo(() => {
    const next: Record<string, string> = {}
    for (const category of teiCategories) {
      const matched = matchTeiCategoryToFeature(
        category.id,
        category.label,
        datasetFeatures,
      )
      next[category.id] = matched?.id || category.id
    }
    return next
  }, [datasetFeatures, teiCategories])

  const resolvedTeiFeatures = useMemo<ResolvedTeiFeature[]>(() => {
    const byId = new Map<string, ResolvedTeiFeature>()
    for (const category of teiCategories) {
      const matched = matchTeiCategoryToFeature(
        category.id,
        category.label,
        datasetFeatures,
      )
      const featureId = matched?.id || category.id
      if (byId.has(featureId)) {
        continue
      }
      byId.set(featureId, {
        id: featureId,
        label: matched?.name?.trim() || category.label,
        description: matched?.description?.trim() || '',
        color: matched?.color || '#f2f2f2',
        isDefault: !!matched?.is_default,
        renderMode: isVerbLike(
          matched?.id,
          matched?.name,
          category.id,
          category.label,
        )
          ? 'outline'
          : 'fill',
      })
    }
    return [...byId.values()].sort((left, right) =>
      left.label.localeCompare(right.label, undefined, { sensitivity: 'base' }),
    )
  }, [datasetFeatures, teiCategories])

  const currentFeatureResults = useMemo(
    () => featureResultsQuery.data || [],
    [featureResultsQuery.data],
  )
  const baselineResultsByFeature = useMemo(() => {
    const map: Record<string, feature_Result> = {}
    for (const result of currentFeatureResults) {
      if (!result.feature_id) {
        continue
      }
      map[result.feature_id] = result
    }
    return map
  }, [currentFeatureResults])

  useEffect(() => {
    const next = toDraftHighlightsFromResults(currentFeatureResults)
    let cancelled = false
    queueMicrotask(() => {
      if (cancelled) {
        return
      }
      setBaselineHighlights(next)
      setDraftHighlights(next)
      setRemovedTeiHighlightIds([])
      setForcedChangedFeatureIds([])
      setFeatureModalState(null)
      setModalFeatureId('')
      setSaveError(null)
    })
    return () => {
      cancelled = true
    }
  }, [
    annotationId,
    currentPageOrKey,
    currentFeatureResults,
    datasetId,
    effectiveTeiSource,
  ])

  const allResolvedFeatures = useMemo<ResolvedTeiFeature[]>(() => {
    const byId = new Map<string, ResolvedTeiFeature>()
    for (const feature of datasetFeatures) {
      if (!feature.id) continue
      byId.set(feature.id, {
        id: feature.id,
        label: feature.name?.trim() || feature.id,
        description: feature.description?.trim() || '',
        color: feature.color || '#f2f2f2',
        isDefault: !!feature.is_default,
        renderMode: isVerbLike(feature.id, feature.name) ? 'outline' : 'fill',
      })
    }
    for (const feature of resolvedTeiFeatures) {
      if (!byId.has(feature.id)) {
        byId.set(feature.id, feature)
      }
    }
    return [...byId.values()].sort((left, right) =>
      left.label.localeCompare(right.label, undefined, { sensitivity: 'base' }),
    )
  }, [datasetFeatures, resolvedTeiFeatures])

  const highlightStorageKey = datasetId
    ? `teiVisibleHighlightFeatures:${datasetId}`
    : 'teiVisibleHighlightFeatures'
  const [storedVisibleFeatureIds, setStoredVisibleFeatureIds] =
    useLocalStorageState<string[] | null>(highlightStorageKey, {
      defaultValue: null,
    })

  const visibleFeatureIds = useMemo(() => {
    const availableIds = allResolvedFeatures.map((feature) => feature.id)
    if (!availableIds.length) {
      return []
    }

    const availableSet = new Set(availableIds)
    const defaultIds = allResolvedFeatures
      .filter((feature) => feature.isDefault)
      .map((feature) => feature.id)

    const order = (ids: string[]) =>
      availableIds.filter((id) => ids.includes(id))

    if (storedVisibleFeatureIds === null) {
      return order(defaultIds)
    }

    const filtered = order(
      storedVisibleFeatureIds.filter((id) => availableSet.has(id)),
    )

    if (storedVisibleFeatureIds.length === 0) {
      return []
    }

    return filtered
  }, [allResolvedFeatures, storedVisibleFeatureIds])

  useEffect(() => {
    if (!allResolvedFeatures.length) {
      if (storedVisibleFeatureIds !== null) {
        setStoredVisibleFeatureIds(null)
      }
      return
    }

    if (!sameStringArray(storedVisibleFeatureIds, visibleFeatureIds)) {
      setStoredVisibleFeatureIds(visibleFeatureIds)
    }
  }, [
    allResolvedFeatures,
    setStoredVisibleFeatureIds,
    storedVisibleFeatureIds,
    visibleFeatureIds,
  ])

  const highlightConfig = useMemo<TeiHighlightConfig | undefined>(() => {
    if (!allResolvedFeatures.length) {
      return undefined
    }

    const categoryConfigById = allResolvedFeatures.reduce<
      Record<
        string,
        {
          label: string
          color: string
          description: string
          renderMode: 'fill' | 'outline'
        }
      >
    >((acc, feature) => {
      acc[feature.id] = {
        label: feature.label,
        color: feature.color,
        description: feature.description,
        renderMode: feature.renderMode,
      }
      return acc
    }, {})

    const manualHighlights: TeiManualHighlight[] = draftHighlights
      .filter((highlight) => highlight.sourceId.startsWith('manual'))
      .map((highlight) => ({
        id: highlight.localId,
        paragraphIndex: highlight.paragraphIndex,
        start: highlight.start,
        end: highlight.end,
        featureId: highlight.featureId,
        surface: highlight.surface,
        normalized: highlight.normalized,
        institution: highlight.institution,
        ancientPersona: highlight.ancientPersona,
      }))

    const draftFeatureIds = [
      ...new Set(draftHighlights.map((h) => h.featureId)),
    ]
    const selectedCategoryIds = [
      ...new Set([...visibleFeatureIds, ...draftFeatureIds]),
    ]
    const draftLocalIds = new Set(
      draftHighlights.map((highlight) => highlight.localId),
    )
    const hiddenFromDraft = baselineHighlights
      .filter(
        (highlight) =>
          !highlight.sourceId.startsWith('manual') &&
          !draftLocalIds.has(highlight.localId),
      )
      .map((highlight) => highlight.sourceId)
    const hiddenTeiHighlightIds = [
      ...new Set([...hiddenFromDraft, ...removedTeiHighlightIds]),
    ]

    return {
      selectedCategoryIds,
      categoryConfigById,
      categoryToFeatureId,
      manualHighlights,
      hiddenTeiHighlightIds,
    }
  }, [
    allResolvedFeatures,
    baselineHighlights,
    categoryToFeatureId,
    draftHighlights,
    removedTeiHighlightIds,
    visibleFeatureIds,
  ])

  const removeHighlightFromTooltip = (item: TeiTooltipItem) => {
    if (isTextEditMode) {
      return
    }
    setRemovedTeiHighlightIds((previous) =>
      previous.includes(item.id) ? previous : [...previous, item.id],
    )
    setDraftHighlights((previous) => removeHighlightFromDrafts(previous, item))
    setForcedChangedFeatureIds((previous) =>
      previous.includes(item.featureId)
        ? previous
        : [...previous, item.featureId],
    )
  }

  const allFeatureOptions = useMemo<FeatureOption[]>(
    () => toFeatureOptions(allResolvedFeatures),
    [allResolvedFeatures],
  )

  const selectedFeatureOptions = useMemo(
    () =>
      allFeatureOptions.filter((option) =>
        visibleFeatureIds.includes(option.value),
      ),
    [allFeatureOptions, visibleFeatureIds],
  )
  const visibleFeatureIdsKey = visibleFeatureIds.join('|')

  const changedFeatureIds = useMemo(() => {
    const baselineByFeature = groupByFeature(baselineHighlights)
    const draftByFeature = groupByFeature(draftHighlights)
    const featureIds = new Set([
      ...Object.keys(baselineByFeature),
      ...Object.keys(draftByFeature),
    ])

    const changedByDiff = [...featureIds].filter((featureId) => {
      const baselineValues = getComparableValues(
        toResultValues(baselineByFeature[featureId] || []),
      )
      const draftValues = getComparableValues(
        toResultValues(draftByFeature[featureId] || []),
      )
      return JSON.stringify(baselineValues) !== JSON.stringify(draftValues)
    })
    return [...new Set([...changedByDiff, ...forcedChangedFeatureIds])]
  }, [baselineHighlights, draftHighlights, forcedChangedFeatureIds])

  const unsavedFeatureCount = changedFeatureIds.length
  const hasUnsavedChanges = unsavedFeatureCount > 0
  const unsavedTextLineCount = useMemo(
    () =>
      editableOriginalLines.filter((line) => line.text !== line.originalText)
        .length,
    [editableOriginalLines],
  )
  const hasUnsavedTextChanges = unsavedTextLineCount > 0
  const canStartTextEdit =
    isAuthenticated &&
    !!teiContents &&
    baseEditableOriginalLines.length > 0 &&
    !hasUnsavedChanges &&
    !isTextEditMode

  const showPane =
    annotation?.ocred || datasetId === TITLE_PAGES_DATASET_ID || !!editionId
  const annotationSourceFailed =
    segmented && annotationTeiQuery.isError && !annotationTeiQuery.data
  const editionSourceFailed =
    !!editionId && editionTeiQuery.isError && !editionTeiQuery.data
  const allCandidateSourcesFailed =
    (segmented ? annotationSourceFailed : true) &&
    (editionId ? editionSourceFailed : true)
  const isFetchingCandidateSource =
    (segmented && annotationTeiQuery.isFetching) ||
    (!!editionId && editionTeiQuery.isFetching)

  const saveMutation = useMutation({
    mutationFn: (results: feature_Result[]) =>
      FeatureResultsService.postDatasetsAnnotationsResults({
        dataSetId: datasetId,
        id: annotationId,
        result: results,
        pushToOrigin: true,
      }),
  })
  const textEditMutation = useMutation({
    mutationFn: (payload: annotationrule_TextBlockCorrections) =>
      AnnotationsApplyRulesService.putDatasetsAnnotationsApplyTextBlockCorrections(
        {
          dataSetId: datasetId,
          id: annotationId,
          annotationTextBlockCorrections: payload,
        },
      ),
  })

  const handleStartTextEdit = () => {
    if (hasUnsavedChanges || !isAuthenticated || !teiContents) {
      return
    }
    if (!orderedSelectedViewModes.includes('original')) {
      setTeiViewModes(
        normalizeTeiViewModes(
          ['original', ...teiViewModes],
          availableViewModes,
        ),
      )
    }
    setEditableOriginalLines(baseEditableOriginalLines)
    setFeatureModalState(null)
    setModalFeatureId('')
    setTextEditError(null)
    setTextEditSavePending(false)
    setIsTextEditMode(true)
  }

  const handleCancelTextEdit = () => {
    setEditableOriginalLines(baseEditableOriginalLines)
    setTextEditError(null)
    setTextEditSavePending(false)
    setIsTextEditMode(false)
  }

  const handleSaveTextEdit = async () => {
    if (!datasetId || !annotationId || !isAuthenticated || !teiContents) {
      return
    }

    if (currentFeatureResults.length > 0 && !textEditSavePending) {
      setTextEditSavePending(true)
      return
    }

    const grouped = new Map<string, { old: string[]; correction: string[] }>()
    for (const line of editableOriginalLines) {
      if (line.text === line.originalText) {
        continue
      }
      const current = grouped.get(line.blockId) || {
        old: [],
        correction: [],
      }
      current.old.push(line.originalText)
      current.correction.push(line.text)
      grouped.set(line.blockId, current)
    }

    if (!grouped.size) {
      setIsTextEditMode(false)
      setTextEditError(null)
      setTextEditSavePending(false)
      return
    }

    const parsedPage = Number.parseInt(String(currentPageOrKey), 10)
    const payload: annotationrule_TextBlockCorrections = {
      type: 'text_blocks_corrections',
      corrections: [...grouped.entries()].map(([blockId, value]) => ({
        text_block_id: zoneToServerTextBlockId[blockId] || blockId,
        old: value.old,
        correction: value.correction,
        page: Number.isFinite(parsedPage) ? parsedPage : undefined,
      })),
    }

    try {
      setTextEditError(null)
      await textEditMutation.mutateAsync(payload)
      setTextEditSavePending(false)
      await Promise.all([
        annotationTeiQuery.refetch(),
        queryClient.invalidateQueries({
          queryKey: [
            'featureResults',
            datasetId,
            annotationId,
            currentPageOrKey,
          ],
        }),
        editionId ? editionTeiQuery.refetch() : Promise.resolve(),
      ])
      setIsTextEditMode(false)
    } catch (error) {
      setTextEditError(
        error instanceof Error ? error.message : 'Failed to save text edits.',
      )
    }
  }

  const handleSave = async () => {
    if (!datasetId || !annotationId || !isAuthenticated || isTextEditMode) {
      return
    }

    const draftByFeature = groupByFeature(draftHighlights)
    const payloads: feature_Result[] = changedFeatureIds.map((featureId) => {
      const existing = baselineResultsByFeature[featureId]
      const preservedValues = (existing?.values || []).filter(
        (value) => !hasTeiPositionProperties(value),
      )
      return {
        ...(existing || {}),
        scope: {
          type: 'dataset',
          dataset_id: datasetId,
          annotation_id: annotationId,
        },
        feature_id: featureId,
        key: existing?.key || String(currentPageOrKey),
        values: [
          ...preservedValues,
          ...toResultValues(draftByFeature[featureId] || []),
        ],
      }
    })

    if (!payloads.length) {
      return
    }

    try {
      setSaveError(null)
      await saveMutation.mutateAsync(payloads)
      setBaselineHighlights(draftHighlights)
      await featureResultsQuery.refetch()
      await queryClient.invalidateQueries({
        queryKey: annotationTeiQueryKey(
          datasetId,
          annotationId,
          currentPageOrKey,
        ),
      })
      if (editionId) {
        await queryClient.invalidateQueries({
          queryKey: editionTeiQueryKey(editionId, currentPageOrKey),
        })
      }
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : 'Failed to save highlights.',
      )
    }
  }

  const addHighlight = (selection: SelectionDraft, featureId: string) => {
    const localId = [
      'manual',
      featureId,
      selection.paragraphIndex,
      selection.start,
      selection.end,
      Date.now(),
      Math.random().toString(36).slice(2, 8),
    ].join(':')

    const next: DraftHighlight = {
      localId,
      sourceId: localId,
      paragraphIndex: selection.paragraphIndex,
      start: selection.start,
      end: selection.end,
      featureId,
      categoryId: featureId,
      surface: selection.surface,
      normalized: '',
      institution: '',
      ancientPersona: '',
      fromAnchorId: '',
      toAnchorId: '',
    }

    setDraftHighlights((previous) => [...previous, next])
  }

  const openModalForAdd = (selection: SelectionDraft) => {
    if (isTextEditMode || allFeatureOptions.length === 0) {
      return
    }
    setFeatureModalState({ selection })
    setModalFeatureId(allFeatureOptions[0]?.value || '')
  }

  const submitFeatureModal = () => {
    if (!featureModalState || !modalFeatureId || isTextEditMode) {
      return
    }

    addHighlight(featureModalState.selection, modalFeatureId)

    setFeatureModalState(null)
    setModalFeatureId('')
  }

  const closeFeatureModal = () => {
    setFeatureModalState(null)
    setModalFeatureId('')
  }

  if (!showPane || allCandidateSourcesFailed) {
    return null
  }

  if (!teiContents && isFetchingCandidateSource) {
    return null
  }

  const sourceOptions: SourceOption[] = availableSources.map((source) => ({
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

  const paneTitle = hasUnsavedChanges
    ? `Contents (${unsavedFeatureCount} unsaved feature${unsavedFeatureCount === 1 ? '' : 's'})`
    : isTextEditMode && hasUnsavedTextChanges
      ? `Contents (${unsavedTextLineCount} unsaved text line${unsavedTextLineCount === 1 ? '' : 's'})`
      : 'Contents'
  const centerTeiRows = datasetId === TITLE_PAGES_DATASET_ID

  return (
    <>
      <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 h-full bg-white">
        <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
          <div>{paneTitle}</div>
          <div className="flex items-center gap-2">
            {textEditError && (
              <span className="text-xs text-red-600">{textEditError}</span>
            )}
            {saveError && (
              <span className="text-xs text-red-600">{saveError}</span>
            )}
            {isAuthenticated && isTextEditMode && !textEditSavePending && (
              <>
                <button
                  type="button"
                  className="px-2 py-1 rounded border border-gray-300 text-gray-700 bg-white hover:bg-gray-50 text-xs"
                  onClick={handleCancelTextEdit}
                  disabled={textEditMutation.isPending}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  className="px-2 py-1 rounded border border-teal-300 text-teal-700 bg-white hover:bg-teal-50 text-xs"
                  onClick={() => {
                    void handleSaveTextEdit()
                  }}
                  disabled={
                    textEditMutation.isPending || !hasUnsavedTextChanges
                  }
                >
                  {textEditMutation.isPending ? 'Saving…' : 'Save'}
                </button>
              </>
            )}
            {isAuthenticated && isTextEditMode && textEditSavePending && (
              <>
                <span className="text-xs text-amber-600">
                  Feature results may be affected by these text edits.
                </span>
                <button
                  type="button"
                  className="px-2 py-1 rounded border border-gray-300 text-gray-700 bg-white hover:bg-gray-50 text-xs"
                  onClick={() => setTextEditSavePending(false)}
                  disabled={textEditMutation.isPending}
                >
                  Back
                </button>
                <button
                  type="button"
                  className="px-2 py-1 rounded border border-teal-300 text-teal-700 bg-white hover:bg-teal-50 text-xs"
                  onClick={() => {
                    void handleSaveTextEdit()
                  }}
                  disabled={textEditMutation.isPending}
                >
                  {textEditMutation.isPending ? 'Saving…' : 'Confirm Save'}
                </button>
              </>
            )}
            {isAuthenticated && hasUnsavedChanges && !isTextEditMode && (
              <button
                type="button"
                className="px-2 py-1 rounded border border-teal-300 text-teal-700 bg-white hover:bg-teal-50 text-xs"
                onClick={() => {
                  void handleSave()
                }}
                disabled={saveMutation.isPending}
              >
                {saveMutation.isPending ? 'Saving…' : 'Save'}
              </button>
            )}
            {!isTextEditMode && !hasUnsavedChanges && (
              <button
                type="button"
                className="px-2 py-1 rounded border border-teal-300 text-teal-700 bg-white hover:bg-teal-50 text-xs disabled:opacity-50 disabled:cursor-not-allowed"
                onClick={handleStartTextEdit}
                disabled={!canStartTextEdit}
              >
                Edit
              </button>
            )}
          </div>
        </div>

        <div className="flex-1 min-h-0 overflow-hidden p-2.5 box-border flex flex-col">
          <div className="flex gap-2 items-center flex-wrap mb-2.5">
            {isLoading && (
              <span className="text-xs text-gray-400 flex items-center gap-1.5">
                <span className="inline-block w-3 h-3 border border-gray-300 border-t-gray-500 rounded-full animate-spin" />
                Loading…
              </span>
            )}
            {sourceOptions.length > 1 && (
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium text-gray-600">
                  Source:
                </span>
                <div className="w-36">
                  <Select
                    value={selectedSourceOption}
                    onChange={(option: SourceOption | null) => {
                      if (option) {
                        setStoredTeiSource(option.value)
                      }
                    }}
                    options={sourceOptions}
                    isClearable={false}
                    isDisabled={isTextEditMode}
                    styles={selectStyles<SourceOption>()}
                    menuPortalTarget={document.body}
                    menuPosition="fixed"
                  />
                </div>
              </div>
            )}
            {availableViewModes.length > 1 && (
              <div className="flex items-center gap-1.5">
                <span className="text-xs font-medium text-gray-600">
                  Views:
                </span>
                <MultiSelectDropdown<TeiViewMode>
                  allItems={availableViewModes}
                  selectedItems={orderedSelectedViewModes}
                  setSelectedItems={(items) => {
                    if (isTextEditMode) {
                      return
                    }
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
            <TeiDisplayControls
              showMinCertControl={showMinCertControl}
              minCert={minCert}
              setMinCert={setMinCert}
              showTeiLineHighlights={showTeiLineHighlights}
              setShowTeiLineHighlights={setShowTeiLineHighlights}
              alignLines={alignLines}
              setAlignLines={setAlignLines}
              showCertaintyVisualization={showCertaintyVisualization}
              setShowCertaintyVisualization={setShowCertaintyVisualization}
              allFeatureOptions={allFeatureOptions}
              selectedFeatureOptions={selectedFeatureOptions}
              isFeatureSelectExpanded={isFeatureSelectExpanded}
              setIsFeatureSelectExpanded={setIsFeatureSelectExpanded}
              setVisibleFeatureIds={setStoredVisibleFeatureIds}
              onResetVisibleFeatureIds={() => setStoredVisibleFeatureIds(null)}
              isFeaturesLoading={featuresQuery.isLoading}
            />
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
          {!isAuthenticated && teiContents && (
            <p className="text-gray-600 text-xs py-1">
              Log in to add, edit, remove, and save highlights.
            </p>
          )}
          {teiContents && (
            <div className="mt-4 flex-1 min-h-0 overflow-y-auto overscroll-none">
              <div className="flex flex-wrap gap-3">
                {orderedSelectedViewModes.map((viewMode) => (
                  <div
                    key={`${viewMode}:${effectiveTeiSource}:${String(currentPageOrKey)}`}
                    className="min-w-105 basis-105 flex-1"
                  >
                    {isTextEditMode && viewMode === 'original' ? (
                      <OriginalViewEditor
                        lines={editableOriginalLines}
                        showViewLabel={availableViewModes.length > 1}
                        onChangeLine={(lineId, text) => {
                          setEditableOriginalLines((previous) =>
                            previous.map((line) =>
                              line.id === lineId ? { ...line, text } : line,
                            ),
                          )
                        }}
                      />
                    ) : (
                      <TeiContentView
                        key={`${viewMode}:${effectiveTeiSource}:${String(currentPageOrKey)}:${visibleFeatureIdsKey}`}
                        data={teiContents}
                        minCert={minCert}
                        showCertaintyVisualization={showCertaintyVisualization}
                        viewMode={viewMode}
                        viewLabel={getViewModeLabel(viewMode)}
                        showViewLabel={availableViewModes.length > 1}
                        alignLines={isTextEditMode ? false : alignLines}
                        centerRows={centerTeiRows}
                        highlightConfig={
                          isTextEditMode ? undefined : highlightConfig
                        }
                        editable={isAuthenticated && !isTextEditMode}
                        canAddHighlight={allFeatureOptions.length > 0}
                        activeLineMatchIds={
                          showTeiLineHighlights ? activeLineMatchIds : []
                        }
                        onHoverLineMatchIds={onHoverLineMatchIds}
                        onRequestAddHighlight={openModalForAdd}
                        onRequestRemoveHighlight={removeHighlightFromTooltip}
                      />
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </section>

      <FeatureHighlightModal
        state={featureModalState}
        isOpen={!isTextEditMode}
        modalFeatureId={modalFeatureId}
        allFeatureOptions={allFeatureOptions}
        onChangeFeatureId={setModalFeatureId}
        onClose={closeFeatureModal}
        onSubmit={submitFeatureModal}
      />
    </>
  )
}
