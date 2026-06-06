import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useAppState } from '../../context/useAppState'
import { AnnotationsApplyRulesService } from '@hub-api'
import { RuleEditModal } from './RuleEditModal.tsx'
import { useDatasetSuggestedRules } from '../../queries/datasets.ts'
import {
  type AnnotationRule,
  isAnnotationRule,
  isRuleApplied,
  type RuleRequestPayload,
} from '../../utils/rules.ts'
import { RuleDisplay } from './RuleDisplay.tsx'
import { useAnnotationRules } from '../../queries/metadata.ts'
import { useAuthStore } from '../../store/authStore.ts'
import { ErrorMessage } from '../core/ErrorMessage'
import {
  isAnnotationRuleApplyJob,
  runningIntegrationJobsQueryKey,
  useRunningIntegrationJobsQuery,
} from '../../queries/integrations.ts'

export function SuggestedRulesPane() {
  const queryClient = useQueryClient()
  const {
    dataset,
    annotation,
    refetch: refetchAnnotation,
    setState,
  } = useAppState()
  const [editingRule, setEditingRule] = useState<AnnotationRule | null>(null)
  const [isManualModalOpen, setIsManualModalOpen] = useState(false)
  const { data: allRules } = useAnnotationRules()
  const { data: runningJobs } = useRunningIntegrationJobsQuery()
  const isAuthenticated = !!useAuthStore((store) => store.token)

  const {
    data: rules,
    refetch: refetchRules,
    isLoading,
    error,
  } = useDatasetSuggestedRules(dataset?.id || '')
  const suggestedRules = (rules || []).flat(2).filter(isAnnotationRule)
  const hasRunningApplyRulesJob = !!runningJobs?.some(
    (job) =>
      isAnnotationRuleApplyJob(job) &&
      job.target?.dataset_id === dataset?.id &&
      job.target?.annotation_id === annotation?.id,
  )

  const handleRunRule = async (
    rule: RuleRequestPayload,
    action: 'overwrite' | 'create_new',
    copyFeatureResults: boolean,
  ) => {
    if (!dataset?.id || !annotation?.id || hasRunningApplyRulesJob) {
      return
    }

    const { name, description, ...rulePayload } = rule
    const isCreate = action === 'create_new'
    const result =
      await AnnotationsApplyRulesService.putDatasetsAnnotationsApply({
        dataSetId: dataset.id,
        id: annotation.id,
        annotationApplyRules: {
          action,
          execution_mode: rule.execution_mode,
          ...(isCreate && {
            copy_feature_results: copyFeatureResults,
            ...(name && { name }),
            ...(description && { description }),
          }),
          rules: [rulePayload],
        },
      })

    void queryClient.invalidateQueries({
      queryKey: runningIntegrationJobsQueryKey(),
    })
    refetchRules()
    refetchAnnotation()
    if ('target' in result) {
      if (
        result.target?.annotation_id &&
        result.target.annotation_id !== annotation.id
      ) {
        setState({ annotationId: result.target.annotation_id })
      }
    } else if (result.id && result.id !== annotation.id) {
      setState({ annotationId: result.id })
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
              disabled={hasRunningApplyRulesJob}
              className="ml-auto px-2 py-1 text-xs font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Run Manually
            </button>
          )}
        </div>

        <div className="flex-1 min-h-0 overflow-auto p-2.5 box-border">
          {hasRunningApplyRulesJob && (
            <div className="mb-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
              A job is running for this annotation. Applying rules is
              unavailable until it finishes.{' '}
              <button
                type="button"
                className="font-medium text-teal-700 underline underline-offset-2 hover:text-teal-900"
                onClick={() => setState({ viewMode: 'jobs' })}
              >
                View jobs
              </button>
            </div>
          )}
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
                const ruleApplicableStages = rule.applicable_stages
                if (annotation?.pipeline_stage && ruleApplicableStages) {
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
                      !hasRunningApplyRulesJob &&
                      !isFutureRuleApplied &&
                      !applied &&
                      canRunRuleBasedOnStage
                        ? () => handleEditRule(rule)
                        : undefined
                    }
                    disabled={
                      isFutureRuleApplied ||
                      applied ||
                      !canRunRuleBasedOnStage ||
                      hasRunningApplyRulesJob
                    }
                    applicableStages={ruleApplicableStages}
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
        initialPayload={editingRule ?? undefined}
        ruleMetadata={allRules}
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
