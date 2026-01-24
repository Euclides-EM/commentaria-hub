import { useState } from 'react'
import { useAppState } from '../context/AppStateContext.tsx'
import { AnnotationsApplyRulesService } from '../api'
import { RuleEditModal } from './RuleEditModal.tsx'
import { useDatasetSuggestedRules } from '../queries/datasets.ts'
import { type AnnotationRule, isRuleApplied } from '../utils/rules.ts'
import { RuleDisplay } from './RuleDisplay.tsx'
import type { annotationrule_PipelineStage } from '../api/models/annotationrule_PipelineStage.ts'
import { useAnnotationRules } from '../queries/metadata.ts'

export function SuggestedRulesPane() {
  const { dataset, annotation, refetch: refetchAnnotation } = useAppState()
  const [editingRule, setEditingRule] = useState<AnnotationRule | null>(null)
  const { data: allRules } = useAnnotationRules()

  const {
    data: rules,
    refetch: refetchRules,
    isLoading,
    error,
  } = useDatasetSuggestedRules(dataset?.id || '')
  const suggestedRules = (rules || []).flat(2) as AnnotationRule[]

  const handleRunRule = async (
    rule: AnnotationRule,
    action: 'overwrite' | 'create_new',
  ) => {
    if (!dataset?.id || !annotation?.id) {
      return
    }

    const params = {
      dataSetId: dataset.id,
      id: annotation.id,
      action,
    }

    switch (rule.type) {
      case 'segment':
        await AnnotationsApplyRulesService.putDatasetsAnnotationsApplySegment({
          ...params,
          annotationSegmentRule: rule,
        })
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
        await AnnotationsApplyRulesService.putDatasetsAnnotationsApplyStretch({
          ...params,
          annotationSegmentRule: rule,
        })
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

    refetchRules()
    refetchAnnotation()
  }

  const handleEditRule = (rule: AnnotationRule) => {
    setEditingRule(rule)
  }

  const handleModalSubmit = async (
    payload: AnnotationRule,
    action: 'overwrite' | 'create_new',
  ) => {
    if (editingRule) {
      await handleRunRule(payload, action)
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
              {
                suggestedRules.filter((r) => isRuleApplied(r, annotation))
                  .length
              }{' '}
              / {suggestedRules.length} applied
            </div>
          )}
        </div>

        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          {isLoading && (
            <div className="text-gray-500 text-sm p-2">Loading rules...</div>
          )}

          {error && (
            <div className="text-red-500 text-sm p-2">{error.message}</div>
          )}

          {!isLoading && !error && suggestedRules.length === 0 && (
            <div className="text-gray-500 text-sm p-2">
              No suggested rules available
            </div>
          )}

          {!isLoading && !error && suggestedRules.length > 0 && (
            <div className="space-y-2">
              {suggestedRules.map((rule, index) => {
                const applied = annotation
                  ? isRuleApplied(rule, annotation)
                  : false
                const isFutureRuleApplied = annotation
                  ? suggestedRules
                      .slice(index + 1)
                      .some((futureRule) =>
                        isRuleApplied(futureRule, annotation),
                      )
                  : false

                let canRunRuleBasedOnStage = true
                if (
                  annotation?.pipeline_stage &&
                  (rule as AnnotationRule).applicable_stages
                ) {
                  const ruleApplicableStages = (rule as AnnotationRule)
                    .applicable_stages as annotationrule_PipelineStage[]
                  const currentAnnotationStage = annotation.pipeline_stage

                  canRunRuleBasedOnStage = ruleApplicableStages.includes(
                    currentAnnotationStage,
                  )
                }

                return (
                  <RuleDisplay
                    key={index}
                    rule={rule}
                    isApplied={applied}
                    onRun={
                      annotation &&
                      !isFutureRuleApplied &&
                      !applied &&
                      canRunRuleBasedOnStage
                        ? () => handleEditRule(rule)
                        : undefined
                    }
                    disabled={
                      isFutureRuleApplied || applied || !canRunRuleBasedOnStage
                    }
                    applicableStages={
                      (rule as AnnotationRule).applicable_stages
                    }
                  />
                )
              })}
            </div>
          )}
        </div>
      </section>

      <RuleEditModal
        isOpen={!!editingRule}
        onClose={() => setEditingRule(null)}
        onSubmit={handleModalSubmit}
        initialPayload={editingRule as AnnotationRule}
      />
    </>
  )
}
