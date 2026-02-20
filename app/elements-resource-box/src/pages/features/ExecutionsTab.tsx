import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { TITLE_PAGES_DATASET_ID } from "../../constants";
import { mapEditionsToItems } from "../../utils/dataUtils";
import { Item } from "../../types";
import { STUDY_CORPORA_FILTER } from "./types";
import { formatDate } from "./helpers";
import {
  Button,
  ButtonRow,
  Card,
  CheckboxList,
  CheckboxRow,
  CollapseButton,
  EmptyState,
  ErrorText,
  ExecutionEditionsToggle,
  ExecutionEditionsToggleContent,
  FeatureActions,
  FeatureDescription,
  FeatureHeader,
  FeatureList,
  FeatureTitle,
  FeatureToken,
  FeatureTokenColor,
  FeatureTokenList,
  Field,
  Form,
  InlineValue,
  ItemDetails,
  ItemRow,
  Label,
  NoRevisionText,
  SearchInput,
  SearchRow,
  Section,
  Select,
  SmallText,
  StatusTag,
} from "./styles";
import pluralize from "pluralize";
import {
  ExecutionsService,
  feature_Execution,
  feature_ExecutionApplyItem,
  feature_ExecutionSkipIf,
  feature_ExecutionStatus,
  feature_Feature,
  FeaturesService,
} from '@hub-api';
import { MultiSelect } from "../../components/tps/filters/MultiSelect.tsx";
import { listAllEditions } from "../../api/editionApi";

const EXECUTION_STATUS_LABELS: Record<feature_ExecutionStatus, string> = {
  success: "Completed",
  failed: "Failed",
  in_progress: "In progress",
  canceling: "Canceling",
  canceled: "Canceled",
};

const EXECUTION_SKIP_IF_OPTIONS: feature_ExecutionSkipIf[] = [
  "feature_exist",
  "revision_exist",
  "human_reviewed",
];

const EXECUTION_SKIP_IF_LABELS: Record<feature_ExecutionSkipIf, string> = {
  feature_exist: "Feature exist",
  revision_exist: "Revision exist",
  human_reviewed: "Human reviewed",
};

const FEATURES_QUERY_KEY = [
  "features",
  TITLE_PAGES_DATASET_ID,
  "revisions",
] as const;
const EXECUTIONS_QUERY_KEY = ["executions", TITLE_PAGES_DATASET_ID] as const;
const EDITIONS_QUERY_KEY = ["editions", "all", "items"] as const;

export function ExecutionsTab() {
  const queryClient = useQueryClient();
  const [actionError, setActionError] = useState<string | null>(null);
  const [cancelingExecutionId, setCancelingExecutionId] = useState<
    string | null
  >(null);

  const [execFormOpen, setExecFormOpen] = useState(false);
  const [editionSearch, setEditionSearch] = useState("");
  const [execFeatureSearch, setExecFeatureSearch] = useState("");
  const [selectedFeatureIds, setSelectedFeatureIds] = useState<Set<string>>(
    new Set(),
  );
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set());
  const [execSkipIf, setExecSkipIf] = useState<feature_ExecutionSkipIf[]>([]);
  const [executionStatusFilter, setExecutionStatusFilter] = useState<
    "all" | feature_ExecutionStatus
  >("all");
  const [expandedExecutionEditions, setExpandedExecutionEditions] = useState<
    Record<string, boolean>
  >({});

  const featuresQuery = useQuery({
    queryKey: FEATURES_QUERY_KEY,
    queryFn: () =>
      FeaturesService.getDatasetsFeatures({
        dataSetId: TITLE_PAGES_DATASET_ID,
        expand: ["revisions"],
      }),
    refetchOnWindowFocus: false,
  });

  const executionsQuery = useQuery({
    queryKey: EXECUTIONS_QUERY_KEY,
    queryFn: () =>
      ExecutionsService.getFeaturesExecutions({
        dataset: TITLE_PAGES_DATASET_ID,
      }),
    refetchOnWindowFocus: false,
  });

  const editionsQuery = useQuery({
    queryKey: EDITIONS_QUERY_KEY,
    queryFn: async () => mapEditionsToItems(await listAllEditions()),
    refetchOnWindowFocus: false,
  });

  const cancelExecutionMutation = useMutation({
    mutationFn: (executionId: string) =>
      ExecutionsService.putFeaturesExecutionsCancel({ executionId }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: EXECUTIONS_QUERY_KEY });
    },
  });

  const createExecutionMutation = useMutation({
    mutationFn: (execution: feature_Execution) =>
      ExecutionsService.postFeaturesExecutions({ execution }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: EXECUTIONS_QUERY_KEY });
    },
  });

  const sortedFeatures = useMemo(
    () =>
      [...(featuresQuery.data ?? [])].sort((a, b) =>
        (a.name || "").localeCompare(b.name || "", undefined, {
          sensitivity: "base",
        }),
      ),
    [featuresQuery.data],
  );
  const executions = executionsQuery.data;
  const loading = featuresQuery.isLoading;
  const executionsLoading =
    executionsQuery.isLoading || executionsQuery.isFetching;
  const editionItemsLoading =
    editionsQuery.isLoading || editionsQuery.isFetching;
  const submittingExecution = createExecutionMutation.isPending;
  const executionsError = useMemo(() => {
    if (actionError) {
      return actionError;
    }
    if (featuresQuery.error instanceof Error) {
      return featuresQuery.error.message;
    }
    if (executionsQuery.error instanceof Error) {
      return executionsQuery.error.message;
    }
    if (editionsQuery.error instanceof Error) {
      return editionsQuery.error.message;
    }
    if (featuresQuery.error || executionsQuery.error || editionsQuery.error) {
      return "Failed to load features data.";
    }
    return null;
  }, [
    actionError,
    editionsQuery.error,
    executionsQuery.error,
    featuresQuery.error,
  ]);

  const uniqueEditionItems = useMemo(
    () =>
      Array.from(
        new Map(
          (editionsQuery.data ?? []).map((item) => [item.key, item]),
        ).values(),
      ),
    [editionsQuery.data],
  );

  const corpusEditionItems = useMemo(
    () =>
      uniqueEditionItems.filter((item) =>
        item.study_corpora.includes(STUDY_CORPORA_FILTER),
      ),
    [uniqueEditionItems],
  );

  const execFilteredFeatures = useMemo(() => {
    const q = execFeatureSearch.trim().toLowerCase();
    if (!q) return sortedFeatures;
    return sortedFeatures.filter(
      (f) =>
        (f.name?.toLowerCase() ?? "").includes(q) ||
        (f.description?.toLowerCase() ?? "").includes(q),
    );
  }, [sortedFeatures, execFeatureSearch]);

  const featureInfoById = useMemo(() => {
    const map: Record<string, { name: string; color?: string }> = {};
    for (const feature of sortedFeatures) {
      if (!feature.id) {
        continue;
      }
      map[feature.id] = {
        name: feature.name || feature.id,
        color: feature.color || undefined,
      };
    }
    return map;
  }, [sortedFeatures]);

  const filteredExecutions = useMemo(() => {
    const executionList = executions ?? [];
    if (executionStatusFilter === "all") {
      return executionList;
    }
    return executionList.filter(
      (execution) => execution.status === executionStatusFilter,
    );
  }, [executions, executionStatusFilter]);

  const filteredEditionItems = useMemo(() => {
    const q = editionSearch.trim().toLowerCase();
    if (!q) {
      return corpusEditionItems;
    }
    return corpusEditionItems.filter((item) => {
      const hay = [
        item.year,
        item.authors.join(", "),
        item.cities.join(", "),
        item.shortTitle,
      ]
        .filter(Boolean)
        .join(" ")
        .toLowerCase();
      return hay.includes(q);
    });
  }, [corpusEditionItems, editionSearch]);

  const editionItemByKey = useMemo(() => {
    const map = new Map<string, Item>();
    for (const item of uniqueEditionItems) {
      map.set(item.key, item);
    }
    return map;
  }, [uniqueEditionItems]);

  const handleCancelExecution = async (executionId: string) => {
    setCancelingExecutionId(executionId);
    setActionError(null);
    try {
      await cancelExecutionMutation.mutateAsync(executionId);
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : "Failed to cancel execution.",
      );
    } finally {
      setCancelingExecutionId(null);
    }
  };

  const toggleFeatureSelection = (featureId: string) => {
    setSelectedFeatureIds((prev) => {
      const next = new Set(prev);
      if (next.has(featureId)) {
        next.delete(featureId);
      } else {
        next.add(featureId);
      }
      return next;
    });
  };

  const toggleKeySelection = (key: string) => {
    setSelectedKeys((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  };

  const toggleExecutionEditions = (executionKey: string) => {
    setExpandedExecutionEditions((prev) => ({
      ...prev,
      [executionKey]: !prev[executionKey],
    }));
  };

  const handleSubmitExecution = async (
    event: React.FormEvent<HTMLFormElement>,
  ) => {
    event.preventDefault();
    if (selectedFeatureIds.size === 0) {
      setActionError("Select at least one feature.");
      return;
    }
    if (selectedKeys.size === 0) {
      setActionError("Select at least one edition.");
      return;
    }
    setActionError(null);
    const featureById = new Map(
      sortedFeatures
        .filter((feature): feature is feature_Feature & { id: string } =>
          Boolean(feature.id),
        )
        .map((feature) => [feature.id, feature]),
    );
    const apply: feature_ExecutionApplyItem[] = [];
    for (const fId of selectedFeatureIds) {
      const feature = featureById.get(fId);
      if (!feature) continue;
      const revisions = [...(feature.revisions ?? [])].sort((a, b) => {
        const tA = a.created_at ? new Date(a.created_at).getTime() : 0;
        const tB = b.created_at ? new Date(b.created_at).getTime() : 0;
        return tB - tA;
      });
      apply.push({
        dataset_id: TITLE_PAGES_DATASET_ID,
        feature: fId,
        revision: revisions[0]?.id,
      });
    }
    const executionPayload: feature_Execution = {
      dataset_id: TITLE_PAGES_DATASET_ID,
      apply,
      keys: Array.from(selectedKeys),
      policy: execSkipIf.length ? { skip_if: execSkipIf } : undefined,
    };
    try {
      await createExecutionMutation.mutateAsync(executionPayload);
      setSelectedFeatureIds(new Set());
      setSelectedKeys(new Set());
      setExecSkipIf([]);
      setExecFormOpen(false);
    } catch (err) {
      setActionError(
        err instanceof Error ? err.message : "Failed to create execution.",
      );
    }
  };

  return (
    <>
      <Section>
        <Card>
          <FeatureHeader>
            <FeatureTitle
              clickable
              onClick={() => setExecFormOpen((prev) => !prev)}
              role="button"
              tabIndex={0}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  setExecFormOpen((prev) => !prev);
                }
              }}
            >
              <CollapseButton
                type="button"
                onClick={() => setExecFormOpen((prev) => !prev)}
                aria-label={
                  execFormOpen
                    ? "Collapse new execution"
                    : "Expand new execution"
                }
              >
                {execFormOpen ? "▾" : "▸"}
              </CollapseButton>{" "}
              Execute
            </FeatureTitle>
          </FeatureHeader>
          {execFormOpen && (
            <Form onSubmit={handleSubmitExecution}>
              <Field>
                <Label>Features ({selectedFeatureIds.size} selected)</Label>
                {loading ? (
                  <EmptyState>Loading features...</EmptyState>
                ) : sortedFeatures.length === 0 ? (
                  <EmptyState>No features available.</EmptyState>
                ) : (
                  <>
                    <SearchRow>
                      <SearchInput
                        placeholder="Search features"
                        value={execFeatureSearch}
                        onChange={(event) =>
                          setExecFeatureSearch(event.target.value)
                        }
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() =>
                          setSelectedFeatureIds(
                            new Set(
                              execFilteredFeatures
                                .map((f) => f.id!)
                                .filter(Boolean),
                            ),
                          )
                        }
                      >
                        Select all
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => setSelectedFeatureIds(new Set())}
                      >
                        Deselect all
                      </Button>
                    </SearchRow>
                    <CheckboxList>
                      {execFilteredFeatures.map((feature) => {
                        const fId = feature.id ?? "";
                        if (!fId) return null;
                        const revisions = [...(feature.revisions ?? [])].sort(
                          (a, b) => {
                            const tA = a.created_at
                              ? new Date(a.created_at).getTime()
                              : 0;
                            const tB = b.created_at
                              ? new Date(b.created_at).getTime()
                              : 0;
                            return tB - tA;
                          },
                        );
                        const latestRev = revisions[0];
                        return (
                          <CheckboxRow key={fId}>
                            <input
                              type="checkbox"
                              checked={selectedFeatureIds.has(fId)}
                              onChange={() => toggleFeatureSelection(fId)}
                            />
                            <span>{feature.name || "Untitled"}</span>
                            {!latestRev && (
                              <NoRevisionText>no revisions</NoRevisionText>
                            )}
                          </CheckboxRow>
                        );
                      })}
                    </CheckboxList>
                  </>
                )}
              </Field>
              <Field>
                <Label>Editions ({selectedKeys.size} selected)</Label>
                {editionItemsLoading ? (
                  <EmptyState>Loading editions...</EmptyState>
                ) : corpusEditionItems.length === 0 ? (
                  <EmptyState>No matching edition items found.</EmptyState>
                ) : (
                  <>
                    <SearchRow>
                      <SearchInput
                        placeholder="Search editions"
                        value={editionSearch}
                        onChange={(event) =>
                          setEditionSearch(event.target.value)
                        }
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() =>
                          setSelectedKeys(
                            new Set(filteredEditionItems.map((i) => i.key)),
                          )
                        }
                      >
                        Select all
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => setSelectedKeys(new Set())}
                      >
                        Deselect all
                      </Button>
                    </SearchRow>
                    <CheckboxList>
                      {filteredEditionItems.map((item) => (
                        <ItemRow key={item.key} interactive>
                          <input
                            type="checkbox"
                            checked={selectedKeys.has(item.key)}
                            onChange={() => toggleKeySelection(item.key)}
                          />
                          <ItemDetails>
                            <InlineValue>
                              {[
                                item.year,
                                item.authors.join(", "),
                                item.cities.join(", "),
                              ]
                                .filter(Boolean)
                                .join(", ") || "—"}
                            </InlineValue>
                            {(item.shortTitle || item.title) && (
                              <span> - {item.shortTitle || item.title}</span>
                            )}
                          </ItemDetails>
                        </ItemRow>
                      ))}
                    </CheckboxList>
                  </>
                )}
              </Field>
              <Field>
                <Label>Skip if (optional)</Label>
                <MultiSelect
                  name="Skip if"
                  options={EXECUTION_SKIP_IF_OPTIONS}
                  value={execSkipIf}
                  onChange={(values) =>
                    setExecSkipIf(values as feature_ExecutionSkipIf[])
                  }
                  labelFn={(value) =>
                    EXECUTION_SKIP_IF_LABELS[
                      value as feature_ExecutionSkipIf
                    ] || value
                  }
                  placeholder="Select skip-if rules"
                />
              </Field>
              {executionsError && <ErrorText>{executionsError}</ErrorText>}
              <ButtonRow>
                <Button type="submit" disabled={submittingExecution}>
                  {submittingExecution ? "Submitting..." : "Submit Execution"}
                </Button>
              </ButtonRow>
            </Form>
          )}
        </Card>
      </Section>

      <Section>
        <FeatureHeader>
          <FeatureTitle>Executions</FeatureTitle>
          <FeatureActions>
            <Button
              type="button"
              variant="primary"
              onClick={() => void executionsQuery.refetch()}
              disabled={executionsLoading}
            >
              {executionsLoading ? "Refreshing..." : "Refresh"}
            </Button>
            <Select
              id="execution-status-filter"
              value={executionStatusFilter}
              aria-label="Filter executions by status"
              onChange={(event) =>
                setExecutionStatusFilter(
                  event.target.value as "all" | feature_ExecutionStatus,
                )
              }
            >
              <option value="all">All statuses</option>
              {(
                Object.keys(
                  EXECUTION_STATUS_LABELS,
                ) as Array<feature_ExecutionStatus>
              ).map((status) => (
                <option key={status} value={status}>
                  {EXECUTION_STATUS_LABELS[status]}
                </option>
              ))}
            </Select>
          </FeatureActions>
        </FeatureHeader>
        {executionsError && !execFormOpen && (
          <ErrorText>{executionsError}</ErrorText>
        )}
        {executionsLoading ? (
          <EmptyState>Loading executions...</EmptyState>
        ) : filteredExecutions.length === 0 ? (
          <EmptyState>
            {(executions?.length ?? 0) === 0
              ? "No executions found yet."
              : "No executions match the selected status."}
          </EmptyState>
        ) : (
          <FeatureList>
            {filteredExecutions.map((execution, index) => {
              const execId = execution.id ?? "";
              const executionCardKey =
                execId || execution.created_at || String(index);
              const isCanceling = cancelingExecutionId === execId;
              const canCancel =
                execution.status === "in_progress" ||
                execution.status === "canceling";
              const executionKeys = Array.from(new Set(execution.keys ?? []));
              const showExecutionEditions =
                expandedExecutionEditions[executionCardKey] ?? false;
              return (
                <Card key={executionCardKey}>
                  <FeatureHeader>
                    <FeatureTitle>
                      {execution.name || execId || "Unnamed"}
                    </FeatureTitle>
                    <FeatureActions>
                      <StatusTag status={execution.status}>
                        {execution.status
                          ? EXECUTION_STATUS_LABELS[execution.status] ||
                            execution.status
                          : "Unknown"}
                      </StatusTag>
                      {canCancel && (
                        <Button
                          variant="danger"
                          type="button"
                          onClick={() =>
                            execId && handleCancelExecution(execId)
                          }
                          disabled={isCanceling}
                        >
                          {isCanceling ? "Canceling..." : "Cancel"}
                        </Button>
                      )}
                    </FeatureActions>
                  </FeatureHeader>
                  {execution.description && (
                    <FeatureDescription>
                      {execution.description}
                    </FeatureDescription>
                  )}
                  <SmallText>
                    Created: {formatDate(execution.created_at)}
                  </SmallText>
                  {executionKeys.length > 0 && (
                    <>
                      <ExecutionEditionsToggle
                        role="button"
                        tabIndex={0}
                        onClick={() =>
                          toggleExecutionEditions(executionCardKey)
                        }
                        onKeyDown={(event) => {
                          if (event.key === "Enter" || event.key === " ") {
                            event.preventDefault();
                            toggleExecutionEditions(executionCardKey);
                          }
                        }}
                      >
                        <ExecutionEditionsToggleContent>
                          <CollapseButton
                            type="button"
                            onClick={(event) => {
                              event.stopPropagation();
                              toggleExecutionEditions(executionCardKey);
                            }}
                            aria-label={
                              showExecutionEditions
                                ? "Collapse included editions"
                                : "Expand included editions"
                            }
                          >
                            {showExecutionEditions ? "▾" : "▸"}
                          </CollapseButton>
                          <span>
                            Including {executionKeys.length}{" "}
                            {pluralize("edition", executionKeys.length)}.
                          </span>
                        </ExecutionEditionsToggleContent>
                      </ExecutionEditionsToggle>
                      {showExecutionEditions && (
                        <CheckboxList>
                          {executionKeys.map((editionKey) => {
                            const item = editionItemByKey.get(editionKey);
                            if (!item) {
                              return (
                                <ItemRow key={editionKey} as="div">
                                  <ItemDetails>
                                    <InlineValue>{editionKey}</InlineValue>
                                    <span> - details unavailable</span>
                                  </ItemDetails>
                                </ItemRow>
                              );
                            }
                            return (
                              <ItemRow key={editionKey} as="div">
                                <ItemDetails>
                                  <InlineValue>
                                    {[
                                      item.year,
                                      item.authors.join(", "),
                                      item.cities.join(", "),
                                    ]
                                      .filter(Boolean)
                                      .join(", ") || "—"}
                                  </InlineValue>
                                  {(item.shortTitle || item.title) && (
                                    <span>
                                      {" "}
                                      - {item.shortTitle || item.title}
                                    </span>
                                  )}
                                </ItemDetails>
                              </ItemRow>
                            );
                          })}
                        </CheckboxList>
                      )}
                    </>
                  )}
                  {execution.apply && execution.apply.length > 0 && (
                    <SmallText>
                      Features:
                      <FeatureTokenList>
                        {execution.apply.map((applyItem, index) => {
                          const featureId = applyItem.feature ?? "";
                          const featureInfo = featureInfoById[featureId];
                          const label =
                            featureInfo?.name || featureId || "Unknown feature";
                          return (
                            <FeatureToken
                              key={`${featureId || "unknown"}-${index}`}
                              title={featureId}
                            >
                              <FeatureTokenColor color={featureInfo?.color} />
                              {label}
                            </FeatureToken>
                          );
                        })}
                      </FeatureTokenList>
                    </SmallText>
                  )}
                </Card>
              );
            })}
          </FeatureList>
        )}
      </Section>
    </>
  );
}
