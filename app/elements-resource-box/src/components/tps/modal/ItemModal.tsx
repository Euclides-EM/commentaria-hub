import { Item } from "../../../types";
import { StyledImage } from "../../common.ts";
import {
  Modal,
  ModalClose,
  ModalContent,
  ModalTextColumn,
  ModalTextContainer,
  ModalTitle,
  TextColumnsContainer,
} from "./ModalComponents.tsx";
import { lazy, Suspense, useMemo, useState } from "react";
import { openImage } from "../../../utils/dataUtils.ts";
import styled from "@emotion/styled";
import { HelpTip } from "../../map/Filter.tsx";
import {
  TOOLTIP_EN_TRANSLATION,
  TOOLTIP_TRANSCRIPTION,
} from "../../map/MapTooltips.tsx";
import { NotesEditor } from "./NotesEditor.tsx";
import { ItemInfo } from "./ItemInfo.tsx";
import { inEditMode } from "../../../utils/mode.ts";
import { TITLE_PAGES_DATASET_ID } from "../../../constants";
import pluralize from "pluralize";
import { toItemImageUrl } from "../../../utils/util.ts";
import {
  HighlightAction,
  HighlightSelection,
  HighlightSpan,
} from "../features/highlightedTextTypes.ts";
import { SingleSelect } from "../filters/SingleSelect.tsx";
import { feature_Feature, FeatureResultsService } from '@hub-api';

const HighlightedText = lazy(() =>
  import("../features/HighlightedText.tsx").then((module) => ({
    default: module.HighlightedText,
  })),
);

type ItemModalProps = {
  item: Item;
  featuresById: Record<string, feature_Feature> | null;
  onClose: () => void;
};

const StyledHelpTip = styled(HelpTip)`
  margin: 0 0 0 -0.5rem;
  z-index: 100;
  svg {
    margin-bottom: 4px;
  }
`;

const NoTitlePage = styled.div`
  flex: 1;
  text-align: center;
  color: darkgray;
`;

const FeatureEditToolbar = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #f7f7f7;
  border: 1px solid #e2e2e2;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  margin-bottom: 0.75rem;
  font-size: 0.9rem;
  color: #333;
  gap: 1rem;
`;

const FeatureEditHint = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  width: 100%;
  text-align: center;
`;

const ModalActions = styled.div`
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
`;

const ActionButton = styled.button<{ variant?: "primary" | "ghost" }>`
  border-radius: 0.4rem;
  border: 1px solid
    ${({ variant }) => (variant === "primary" ? "#1f1f1f" : "#cfcfcf")};
  background: ${({ variant }) => (variant === "primary" ? "#1f1f1f" : "white")};
  color: ${({ variant }) => (variant === "primary" ? "white" : "#333")};
  padding: 0.5rem 0.9rem;
  font-size: 0.9rem;
  cursor: pointer;
  min-width: 5rem;

  &:disabled {
    opacity: 0.6;
    cursor: default;
  }
`;

const SaveStatus = styled.div`
  margin-top: 0.5rem;
  color: #b00020;
  font-size: 0.85rem;
`;

const AnnotationBackdrop = styled.div`
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 12000;
`;

const AnnotationModal = styled.div`
  background: white;
  border-radius: 0.75rem;
  padding: 1.5rem;
  width: min(32rem, 90vw);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
`;

const AnnotationText = styled.div`
  background: #f5f5f5;
  border-radius: 0.4rem;
  padding: 0.6rem 0.75rem;
  font-size: 0.9rem;
  max-height: 8rem;
  overflow: auto;
  white-space: pre-wrap;
  text-align: center;
`;

const FormLabel = styled.label`
  font-size: 0.85rem;
  font-weight: 600;
  color: #444;
`;

const TextArea = styled.textarea`
  background: white;
  color: black;
  min-height: 5rem;
  border-radius: 0.4rem;
  border: 1px solid #cfcfcf;
  padding: 0.5rem;
  font-size: 0.9rem;
  resize: vertical;
`;

type PendingHighlightEdit = {
  id: string;
  action: "add" | "remove";
  featureKey: string;
  start: number;
  end: number;
  text: string;
  note?: string;
};

export const ItemModal = ({ item, featuresById, onClose }: ItemModalProps) => {
  const highlightFeatures = useMemo(() => featuresById || {}, [featuresById]);
  const hasTitleText = !!item.title && item.title !== "?";
  const imageUrl = toItemImageUrl(item.tpImageName);
  const canEditHighlights =
    inEditMode() && Object.keys(highlightFeatures).length > 0;
  const [pendingEdits, setPendingEdits] = useState<PendingHighlightEdit[]>([]);
  const [addedHighlights, setAddedHighlights] = useState<HighlightSpan[]>([]);
  const [removedHighlightIds, setRemovedHighlightIds] = useState<Set<string>>(
    new Set(),
  );
  const [annotationDraft, setAnnotationDraft] =
    useState<HighlightSelection | null>(null);
  const [annotationFeature, setAnnotationFeature] = useState("");
  const [annotationNote, setAnnotationNote] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const sortedFeatureIds = useMemo(() => {
    const ids = Object.keys(highlightFeatures);
    return ids
      .slice()
      .sort((a, b) =>
        (highlightFeatures[a].name || a).localeCompare(
          highlightFeatures[b].name || b,
          undefined,
          { sensitivity: "base" },
        ),
      );
  }, [highlightFeatures]);

  const featureOptions = useMemo(
    () =>
      sortedFeatureIds.map((featureId) => ({
        value: featureId,
        label: highlightFeatures[featureId].name || featureId,
        color: highlightFeatures[featureId].color,
      })),
    [highlightFeatures, sortedFeatureIds],
  );

  const openAnnotationModal = (selection: HighlightSelection) => {
    setAnnotationDraft(selection);
    setAnnotationNote("");
    if (sortedFeatureIds.length > 0) {
      setAnnotationFeature((prev) =>
        sortedFeatureIds.includes(prev) ? prev : sortedFeatureIds[0],
      );
    }
  };

  const closeAnnotationModal = () => {
    setAnnotationDraft(null);
    setAnnotationNote("");
  };

  const handleAddAnnotation = () => {
    if (!annotationDraft || !annotationFeature) {
      return;
    }
    const id = `local-${Date.now()}-${Math.floor(Math.random() * 100000)}`;
    const normalizedNote = annotationNote.trim();
    const highlight: HighlightSpan = {
      id,
      start: annotationDraft.start,
      end: annotationDraft.end,
      featureKey: annotationFeature,
      normalized: "",
      source: "local",
    };
    setAddedHighlights((prev) => [...prev, highlight]);
    setPendingEdits((prev) => [
      ...prev,
      {
        id,
        action: "add",
        featureKey: annotationFeature,
        start: annotationDraft.start,
        end: annotationDraft.end,
        text: annotationDraft.text,
        note: normalizedNote || undefined,
      },
    ]);
    closeAnnotationModal();
  };

  const handleRemoveHighlight = (highlight: HighlightAction) => {
    setSaveError(null);
    if (highlight.id.startsWith("local-")) {
      setAddedHighlights((prev) =>
        prev.filter((entry) => entry.id !== highlight.id),
      );
      setPendingEdits((prev) =>
        prev.filter((entry) => entry.id !== highlight.id),
      );
      return;
    }

    setRemovedHighlightIds((prev) => {
      if (prev.has(highlight.id)) {
        return prev;
      }
      const next = new Set(prev);
      next.add(highlight.id);
      return next;
    });

    setPendingEdits((prev) => {
      if (prev.some((entry) => entry.id === highlight.id)) {
        return prev;
      }
      return [
        ...prev,
        {
          id: highlight.id,
          action: "remove",
          featureKey: highlight.featureKey,
          start: highlight.start,
          end: highlight.end,
          text: highlight.text,
          note: "Removed highlight",
        },
      ];
    });
  };

  const handleSaveEdits = async () => {
    if (pendingEdits.length === 0) {
      return;
    }
    setIsSaving(true);
    setSaveError(null);
    try {
      await Promise.all(
        pendingEdits.map((edit) =>
          FeatureResultsService.postDatasetsAnnotationsResults({
            dataSetId: TITLE_PAGES_DATASET_ID,
            id: item.key,
            result: {
              dataset_id: TITLE_PAGES_DATASET_ID,
              feature: edit.featureKey,
              key: item.key,
              note:
                edit.action === "remove"
                  ? edit.note || "Removed annotation"
                  : edit.note,
              values: edit.text
                ? [
                    {
                      root: edit.text,
                      source: {
                        resp: "human",
                      },
                    },
                  ]
                : undefined,
            },
          }),
        ),
      );
      setPendingEdits([]);
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : "Failed to save edits.",
      );
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Modal onClick={onClose}>
      <ModalContent
        onClick={(e) => e.stopPropagation()}
        hasImage={!!item.tpImageName}
      >
        <ModalClose title="Close" onClick={onClose}>
          ✕
        </ModalClose>
        <ItemInfo isRow={Boolean(imageUrl || hasTitleText)} item={item} />
        <ModalTextContainer>
          {imageUrl && (
            <ModalTextColumn isImage>
              <StyledImage
                large
                clickable
                src={imageUrl}
                onClick={() => openImage(item)}
              />
            </ModalTextColumn>
          )}
          {hasTitleText ? (
            <TextColumnsContainer>
              <ModalTextColumn isTextContent alignCenter={!!imageUrl}>
                <ModalTitle justifyStart gap={1}>
                  Original Text
                  <StyledHelpTip tooltipId={TOOLTIP_TRANSCRIPTION} />
                </ModalTitle>
                {canEditHighlights && (
                  <FeatureEditToolbar>
                    <FeatureEditHint>
                      Select text to add an annotation. Hover a highlight to
                      remove it.
                    </FeatureEditHint>
                  </FeatureEditToolbar>
                )}
                <Suspense fallback={<div>{item.title}</div>}>
                  <HighlightedText
                    text={item.title!}
                    featuresById={highlightFeatures}
                    itemKey={item.key}
                    editable={canEditHighlights}
                    addedHighlights={addedHighlights}
                    removedHighlightIds={removedHighlightIds}
                    onRequestAddAnnotation={openAnnotationModal}
                    onRemoveHighlight={handleRemoveHighlight}
                  />
                </Suspense>
                {item.imprint && (
                  <>
                    <hr style={{ opacity: 0.3 }} />
                    {item.imprint}
                  </>
                )}
              </ModalTextColumn>
              {(item.titleEn || item.imprintEn) && (
                <ModalTextColumn isTextContent alignCenter={!!imageUrl}>
                  <ModalTitle justifyStart gap={1}>
                    English Translation{" "}
                    <StyledHelpTip tooltipId={TOOLTIP_EN_TRANSLATION} />
                  </ModalTitle>
                  <div>{item.titleEn}</div>
                  {item.imprintEn && (
                    <>
                      {item.imprint && <hr style={{ opacity: 0.3 }} />}
                      <div>{item.imprintEn}</div>
                    </>
                  )}
                </ModalTextColumn>
              )}
            </TextColumnsContainer>
          ) : (
            <NoTitlePage>
              This edition has no title page or it is not available.
            </NoTitlePage>
          )}
        </ModalTextContainer>
        {pendingEdits.length > 0 && (
          <>
            <ModalActions>
              <ActionButton
                variant="primary"
                onClick={handleSaveEdits}
                disabled={isSaving}
              >
                {isSaving
                  ? "Saving..."
                  : `Save ${pendingEdits.length} annotation ${pluralize("edit", pendingEdits.length)}`}
              </ActionButton>
            </ModalActions>
            {saveError && <SaveStatus>{saveError}</SaveStatus>}
          </>
        )}
        {inEditMode() && <NotesEditor item={item} />}
        {annotationDraft && (
          <AnnotationBackdrop onClick={closeAnnotationModal}>
            <AnnotationModal onClick={(e) => e.stopPropagation()}>
              <strong>Tag selection as annotation</strong>
              <AnnotationText>{annotationDraft.text}</AnnotationText>
              <FormLabel>Feature</FormLabel>
              <SingleSelect
                name="feature"
                options={featureOptions}
                value={annotationFeature || null}
                onChange={(value) =>
                  setAnnotationFeature(typeof value === "string" ? value : "")
                }
                placeholder="Select feature"
                formatOptionLabel={(option) => (
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "0.5rem",
                    }}
                  >
                    <span
                      style={{
                        width: "0.7rem",
                        height: "0.7rem",
                        borderRadius: "999px",
                        background: option.color || "#d0d0d0",
                        border: "1px solid rgba(0,0,0,0.1)",
                      }}
                    />
                    <span>{option.label}</span>
                  </div>
                )}
              />
              <FormLabel>Note (optional)</FormLabel>
              <TextArea
                value={annotationNote}
                onChange={(e) => setAnnotationNote(e.target.value)}
              />
              <ModalActions>
                <ActionButton variant="ghost" onClick={closeAnnotationModal}>
                  Cancel
                </ActionButton>
                <ActionButton
                  variant="primary"
                  onClick={handleAddAnnotation}
                  disabled={!annotationFeature}
                >
                  Add
                </ActionButton>
              </ModalActions>
            </AnnotationModal>
          </AnnotationBackdrop>
        )}
      </ModalContent>
    </Modal>
  );
};
