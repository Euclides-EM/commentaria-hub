import { useState } from 'react'
import { useAppState } from '../../context/useAppState'
import type { annotationrule_PipelineStage } from '@hub-api'
import { AnnotationsApplyRulesService } from '@hub-api'
import { RuleEditModal } from './RuleEditModal.tsx'
import { useDatasetSuggestedRules } from '../../queries/datasets.ts'
import {
  type AnnotationRule,
  isRuleApplied,
  type RuleRequestPayload,
} from '../../utils/rules.ts'
import { RuleDisplay } from './RuleDisplay.tsx'
import { useAnnotationRules } from '../../queries/metadata.ts'
import { useAuthStore } from '../../store/authStore.ts'
import { ErrorMessage } from '../core/ErrorMessage'

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
  const isAuthenticated = !!useAuthStore((store) => store.token)

  const {
    data: rules,
    refetch: refetchRules,
    isLoading,
    error,
  } = useDatasetSuggestedRules(dataset?.id || '')
  const suggestedRules = (rules || []).flat(2) as AnnotationRule[]

  const handleRunRule = async (
    rule: RuleRequestPayload,
    action: 'overwrite' | 'create_new',
    copyFeatureResults: boolean,
  ) => {
    if (!dataset?.id || !annotation?.id) {
      return
    }

    const { name, description, ...rulePayload } = rule
    const isCreate = action === 'create_new'
    const annotationResult =
      await AnnotationsApplyRulesService.putDatasetsAnnotationsApply({
        dataSetId: dataset.id,
        id: annotation.id,
        annotationApplyRules: {
          action,
          ...(isCreate && {
            copy_feature_results: copyFeatureResults,
            ...(name && { name }),
            ...(description && { description }),
          }),
          rules: [rulePayload as AnnotationRule],
        },
      })

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
    payload: RuleRequestPayload,
    action: 'overwrite' | 'create_new',
    copyFeatureResults: boolean,
  ) => {
    await handleRunRule(payload, action, copyFeatureResults)
    setIsManualModalOpen(false)
  }

  const handleEditRuleSubmit = async (
    payload: RuleRequestPayload,
    action: 'overwrite' | 'create_new',
    copyFeatureResults: boolean,
  ) => {
    await handleRunRule(payload, action, copyFeatureResults)
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
          {annotation && isAuthenticated && (
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

          {error && <ErrorMessage error={error} variant="compact" />}

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
                      isAuthenticated &&
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
