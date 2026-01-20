import { useEffect, useState } from 'react'
import { useAppState } from '../context/AppStateContext.tsx'
import type {
  annotationrule_Segment,
  annotationrule_SlicePages,
  annotationrule_Stretch,
  annotationrule_AddMargin,
  annotationrule_LinesDetect,
  annotationrule_RemoveCategories,
  annotationrule_RemoveOverlap,
  annotationrule_ReassignTextLinesByTolerance,
  annotationrule_TextBlockCorrections,
} from '../api'
import { DatasetsService, AnnotationsApplyRulesService } from '../api'
import { RuleEditModal } from './RuleEditModal.tsx'

type AnnotationRule =
  | annotationrule_Segment
  | annotationrule_SlicePages
  | annotationrule_Stretch
  | annotationrule_AddMargin
  | annotationrule_LinesDetect
  | annotationrule_RemoveCategories
  | annotationrule_RemoveOverlap
  | annotationrule_ReassignTextLinesByTolerance
  | annotationrule_TextBlockCorrections

interface AnnotationRuleWithIndex {
  _index?: number
  type?: string
  pages?: string
  random_pages?: number
  model?: string
  [key: string]: unknown
}

export function SuggestedRulesPane() {
  const { dataset, annotation } = useAppState()
  const [suggestedRules, setSuggestedRules] = useState<AnnotationRule[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [editingRule, setEditingRule] =
    useState<AnnotationRuleWithIndex | null>(null)
  const [runningRuleIndex, setRunningRuleIndex] = useState<number | null>(null)

  useEffect(() => {
    if (!dataset?.id) {
      setSuggestedRules([])
      return
    }

    const fetchSuggestedRules = async () => {
      setLoading(true)
      setError(null)
      try {
        const rules = await DatasetsService.getDatasetsSuggestedRules({
          dataSetId: dataset.id!,
        })
        const flattenedRules = rules.flat(2) as AnnotationRule[]
        setSuggestedRules(flattenedRules)
      } catch (err) {
        setError('Failed to load suggested rules')
        console.error('Error fetching suggested rules:', err)
      } finally {
        setLoading(false)
      }
    }

    fetchSuggestedRules()
  }, [dataset?.id])

  const isRuleApplied = (suggestedRule: AnnotationRule): boolean => {
    if (!annotation?.applied_rules || annotation.applied_rules.length === 0) {
      return false
    }

    return annotation.applied_rules.some((appliedRule: unknown) => {
      const typedAppliedRule = appliedRule as AnnotationRule
      if (suggestedRule.type === typedAppliedRule.type) {
        switch (suggestedRule.type) {
          case 'slice_pages':
            const sliceSuggested = suggestedRule as annotationrule_SlicePages
            const sliceApplied = typedAppliedRule as annotationrule_SlicePages
            return (
              sliceSuggested.pages === sliceApplied.pages ||
              sliceSuggested.random_pages === sliceApplied.random_pages
            )
          case 'segment':
            const segmentSuggested = suggestedRule as annotationrule_Segment
            const segmentApplied = typedAppliedRule as annotationrule_Segment
            return segmentSuggested.model === segmentApplied.model
          case 'stretch':
          case 'add_margin':
          case 'lines_detect':
          case 'remove_categories':
          case 'remove_overlap':
          case 'reassign_text_lines_by_tolerance':
          case 'text_blocks_corrections':
            const suggestedKeys = Object.keys(suggestedRule).filter(
              (k) => k !== 'type',
            )
            const appliedKeys = Object.keys(typedAppliedRule).filter(
              (k) => k !== 'type',
            )

            if (suggestedKeys.length === 0 && appliedKeys.length === 0) {
              return true
            }

            return suggestedKeys.every(
              (key) =>
                (suggestedRule as Record<string, unknown>)[key] ===
                (typedAppliedRule as Record<string, unknown>)[key],
            )
          default:
            return (
              JSON.stringify(suggestedRule) === JSON.stringify(typedAppliedRule)
            )
        }
      }
      return false
    })
  }

  const getRuleDisplayName = (rule: AnnotationRule): string => {
    if (!rule.type) return 'Unknown Rule'

    const typeDisplay = rule.type
      .split('_')
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ')

    const details: string[] = []

    if (rule.type === 'slice_pages') {
      const sliceRule = rule as annotationrule_SlicePages
      if (sliceRule.pages) {
        details.push(`Pages: ${sliceRule.pages}`)
      } else if (sliceRule.random_pages) {
        details.push(`Random: ${sliceRule.random_pages} pages`)
      }
    } else if (rule.type === 'segment') {
      const segmentRule = rule as annotationrule_Segment
      if (segmentRule.model) {
        details.push(`Model: ${segmentRule.model}`)
      }
    }

    return details.length > 0
      ? `${typeDisplay} (${details.join(', ')})`
      : typeDisplay
  }

  const handleRunRule = async (
    rule: AnnotationRule,
    action: 'overwrite' | 'create_new',
    ruleIndex: number,
  ) => {
    if (!dataset?.id || !annotation?.id) return

    setRunningRuleIndex(ruleIndex)
    setError(null)

    try {
      const params = {
        dataSetId: dataset.id,
        id: annotation.id,
        action,
      }

      switch (rule.type) {
        case 'segment':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplySegment(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'slice_pages':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplySlicePages(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'stretch':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyStretch(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'add_margin':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyAddMargin(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'lines_detect':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyDetectLines(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'remove_categories':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyRemoveCategories(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'remove_overlap':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyRemoveOverlap(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'reassign_text_lines_by_tolerance':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyReassignTextLinesByTolerance(
            {
              ...params,
              annotationSegmentRule: rule,
            },
          )
          break
        case 'text_blocks_corrections':
          await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyTextBlockCorrections(
            {
              ...params,
              annotationTextBlockCorrections: rule,
            },
          )
          break
        default:
          throw new Error(`Unsupported rule type: ${rule.type}`)
      }

      window.location.reload()
    } catch (err) {
      setError(`Failed to apply rule: ${err}`)
      console.error('Error applying rule:', err)
    } finally {
      setRunningRuleIndex(null)
    }
  }

  const handleEditRule = (rule: AnnotationRule, ruleIndex: number) => {
    // Store the rule with its index for reference
    const ruleWithIndex = {
      ...rule,
      _index: ruleIndex,
    } as AnnotationRuleWithIndex
    setEditingRule(ruleWithIndex)
  }

  const handleModalSubmit = (
    payload: AnnotationRule,
    action: 'overwrite' | 'create_new',
  ) => {
    if (editingRule && editingRule._index !== undefined) {
      handleRunRule(payload, action, editingRule._index)
    }
    setEditingRule(null)
  }

  return (
    <>
      <section className="border border-gray-300 rounded-xl overflow-hidden flex flex-col min-h-0 bg-white m-3 mb-0">
        <div className="px-2.5 py-2 border-b border-gray-200 text-sm font-semibold bg-gray-50 flex items-center justify-between gap-2.5">
          <div>Suggested Rules</div>
          {annotation && suggestedRules.length > 0 && (
            <div className="text-xs font-normal text-gray-600">
              {suggestedRules.filter((r) => isRuleApplied(r)).length} /{' '}
              {suggestedRules.length} applied
            </div>
          )}
        </div>

        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          {loading && (
            <div className="text-gray-500 text-sm p-2">Loading rules...</div>
          )}

          {error && <div className="text-red-500 text-sm p-2">{error}</div>}

          {!loading && !error && suggestedRules.length === 0 && (
            <div className="text-gray-500 text-sm p-2">
              No suggested rules available
            </div>
          )}

          {!loading && !error && suggestedRules.length > 0 && (
            <div className="space-y-2">
              {suggestedRules.map((rule, index) => {
                const applied = annotation ? isRuleApplied(rule) : false
                const isRunning = runningRuleIndex === index
                return (
                  <div
                    key={index}
                    className={`p-3 border rounded-lg transition-colors ${
                      isRunning
                        ? 'bg-blue-50 border-blue-300'
                        : applied
                          ? 'bg-green-50 border-green-300'
                          : 'bg-gray-50 border-gray-200 hover:bg-gray-100'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex-1">
                        <div className="font-medium text-sm mb-1">
                          {getRuleDisplayName(rule)}
                        </div>
                        <details className="text-xs">
                          <summary className="cursor-pointer text-gray-600 hover:text-gray-800">
                            View raw rule data
                          </summary>
                          <pre className="mt-2 p-2 bg-white rounded border border-gray-200 font-mono overflow-auto whitespace-pre-wrap break-all">
                            {JSON.stringify(rule, null, 2)}
                          </pre>
                        </details>
                      </div>
                      <div className="flex items-center gap-2">
                        {isRunning ? (
                          <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-blue-100 text-blue-700 rounded-full">
                            Running...
                          </span>
                        ) : (
                          <>
                            {applied && (
                              <span className="inline-flex items-center px-2 py-1 text-xs font-medium bg-green-100 text-green-700 rounded-full">
                                Applied
                              </span>
                            )}
                            {annotation && (
                              <button
                                onClick={() => handleEditRule(rule, index)}
                                disabled={isRunning}
                                className="inline-flex items-center px-3 py-1 text-xs font-medium text-white bg-blue-600 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                              >
                                Run
                              </button>
                            )}
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </section>

      {editingRule && (
        <RuleEditModal
          isOpen={!!editingRule}
          onClose={() => setEditingRule(null)}
          onSubmit={handleModalSubmit}
          initialPayload={editingRule as AnnotationRule}
          ruleType={editingRule.type as AnnotationRule['type']}
        />
      )}
    </>
  )
}
