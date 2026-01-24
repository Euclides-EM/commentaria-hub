import { useState } from 'react'
import { useAppState } from '../context/AppStateContext.tsx'
import type {
  annotationrule_AddMargin,
  annotationrule_LinesDetect,
  annotationrule_PipelineStage,
  annotationrule_ReassignTextLinesByTolerance,
  annotationrule_RemoveCategories,
  annotationrule_RemoveOverlap,
  annotationrule_Segment,
  annotationrule_SlicePages,
  annotationrule_Stretch,
  annotationrule_TextBlockCorrections,
  annotationrule_Type,
  model_Annotation,
} from '../api'
import { AnnotationsApplyRulesService } from '../api'
import { RuleEditModal } from './RuleEditModal.tsx'
import { useDatasetSuggestedRules } from '../queries/datasets.ts'
import { type AnnotationRule, isRuleApplied } from '../utils/rules.ts'
import { RuleDisplay } from './RuleDisplay.tsx'
import { useAnnotationRules } from '../queries/metadata.ts'

type BaseRunRuleParams = {
  dataSetId: string
  id: string
  action: 'overwrite' | 'create_new'
}

type RuleRunner = (
  baseParams: BaseRunRuleParams,
  rule: AnnotationRule,
) => Promise<model_Annotation>

const ruleRunnerMap: Record<annotationrule_Type, RuleRunner> = {
  segment: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplySegment({
      ...baseParams,
      annotationSegmentRule: rule as annotationrule_Segment,
    }),
  slice_pages: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplySlicePages({
      ...baseParams,
      annotationSegmentRule: rule as annotationrule_SlicePages,
    }),
  stretch: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplyStretch({
      ...baseParams,
      annotationSegmentRule: rule as annotationrule_Stretch,
    }),
  add_margin: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplyAddMargin({
      ...baseParams,
      annotationSegmentRule: rule as annotationrule_AddMargin,
    }),
  lines_detect: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplyDetectLines({
      ...baseParams,
      annotationSegmentRule: rule as annotationrule_LinesDetect,
    }),
  remove_categories: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplyRemoveCategories({
      ...baseParams,
      annotationSegmentRule: rule as annotationrule_RemoveCategories,
    }),
  remove_overlap: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplyRemoveOverlap({
      ...baseParams,
      annotationSegmentRule: rule as annotationrule_RemoveOverlap,
    }),
  reassign_text_lines_by_tolerance: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplyReassignTextLinesByTolerance(
      {
        ...baseParams,
        annotationSegmentRule:
          rule as annotationrule_ReassignTextLinesByTolerance,
      },
    ),
  text_blocks_corrections: (baseParams, rule) =>
    AnnotationsApplyRulesService.putDatasetsAnnotationsApplyTextBlockCorrections(
      {
        ...baseParams,
        annotationTextBlockCorrections:
          rule as annotationrule_TextBlockCorrections,
      },
    ),
}

export function SuggestedRulesPane() {
  const {
    dataset,
    annotation,
    refetch: refetchAnnotation,
    setState,
  } = useAppState()
  const [editingRule, setEditingRule] = useState<AnnotationRule | null>(null)
  const [isManualModalOpen, setIsManualModalOpen] = useState(false)
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

    const baseParams: BaseRunRuleParams = {
      dataSetId: dataset.id,
      id: annotation.id,
      action,
    }

    const runner = ruleRunnerMap[rule.type!]

    if (!runner) {
      throw new Error(`Unsupported rule type: ${rule.type}`)
    }

    const annotationResult = await runner(baseParams, rule)

    refetchRules()
    refetchAnnotation()
    if (annotationResult.id !== annotation.id) {
      setState({ annotationId: annotationResult.id })
    }
  }

  const handleEditRule = (rule: AnnotationRule) => {
    setEditingRule(rule)
  }

  const handleManualRuleSubmit = async (
    payload: AnnotationRule,
    action: 'overwrite' | 'create_new',
  ) => {
    await handleRunRule(payload, action)
    setIsManualModalOpen(false)
  }

  const handleEditRuleSubmit = async (
    payload: AnnotationRule,
    action: 'overwrite' | 'create_new',
  ) => {
    await handleRunRule(payload, action)
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
          {annotation && (
            <button
              onClick={() => setIsManualModalOpen(true)}
              className="ml-auto px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Run Manually
            </button>
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
        onSubmit={handleEditRuleSubmit}
        initialPayload={editingRule as AnnotationRule}
        ruleMetadata={undefined}
      />

      <RuleEditModal
        isOpen={isManualModalOpen}
        onClose={() => setIsManualModalOpen(false)}
        onSubmit={handleManualRuleSubmit}
        ruleMetadata={allRules}
      />
    </>
  )
}
