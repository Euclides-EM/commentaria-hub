import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react";
import FeatureHighlightTooltip from "./FeatureHighlightTooltip";
import {
  handleHighlightTooltipMouseLeave,
  handleHighlightTooltipMouseMove,
  HighlightTooltipState,
} from "./highlightTooltipUtils";
import {
  buildHighlightHitLayers,
  buildHighlightLayers,
  buildTextHtml,
  normalizeDisplayText,
} from "./highlightedTextRenderUtils";
import { getSelectionState } from "./highlightedTextSelectionUtils";
import {
  getCachedOrFetchTei,
  parseTeiToSpans,
} from "./highlightedTextTeiUtils";
import type {
  HighlightAction,
  HighlightSelection,
  HighlightSpan,
} from "./highlightedTextTypes";
import { feature_Feature } from '@hub-api';

export type { HighlightAction, HighlightSelection, HighlightSpan };

type HighlightedTextProps = {
  text: string;
  featuresById: Record<string, feature_Feature>;
  itemKey: string;
  editable?: boolean;
  addedHighlights?: HighlightSpan[];
  removedHighlightIds?: Set<string>;
  onRequestAddAnnotation?: (selection: HighlightSelection) => void;
  onRemoveHighlight?: (highlight: HighlightAction) => void;
};

export const HighlightedText = memo(
  ({
    text,
    featuresById,
    itemKey,
    editable = false,
    addedHighlights = [],
    removedHighlightIds,
    onRequestAddAnnotation,
    onRemoveHighlight,
  }: HighlightedTextProps) => {
    const [renderedHtml, setRenderedHtml] = useState<string>("");
    const [isReady, setIsReady] = useState(false);
    const [tooltipState, setTooltipState] =
      useState<HighlightTooltipState | null>(null);
    const [tooltipPinned, setTooltipPinned] = useState(false);
    const [teiSpans, setTeiSpans] = useState<HighlightSpan[]>([]);
    const [displayText, setDisplayText] = useState("");
    const [selectionState, setSelectionState] =
      useState<HighlightSelection | null>(null);
    const containerRef = useRef<HTMLDivElement>(null);

    const plainHtml = useMemo(() => buildTextHtml(text), [text]);
    const normalizedPlainText = useMemo(
      () => normalizeDisplayText(text),
      [text],
    );

    const combinedSpans = useMemo(() => {
      const selectedSet = new Set(Object.keys(featuresById));
      const removedIds = removedHighlightIds ?? new Set<string>();
      const normalizedText = displayText || normalizedPlainText;

      return [
        ...teiSpans.filter((span) => !removedIds.has(span.id)),
        ...addedHighlights,
      ]
        .map((span) => ({
          ...span,
          start: Math.max(0, Math.min(span.start, normalizedText.length)),
          end: Math.max(0, Math.min(span.end, normalizedText.length)),
        }))
        .filter((span) => span.end > span.start)
        .filter(
          (span) => selectedSet.size === 0 || selectedSet.has(span.featureKey),
        );
    }, [
      addedHighlights,
      displayText,
      featuresById,
      normalizedPlainText,
      removedHighlightIds,
      teiSpans,
    ]);

    const normalizedText = displayText || normalizedPlainText;

    const renderedLayers = useMemo(
      () => buildHighlightLayers(normalizedText, combinedSpans, featuresById),
      [combinedSpans, featuresById, normalizedText],
    );

    const renderedHitLayers = useMemo(
      () =>
        buildHighlightHitLayers(normalizedText, combinedSpans, featuresById),
      [combinedSpans, featuresById, normalizedText],
    );

    useEffect(() => {
      let isMounted = true;

      const finalize = (
        html: string,
        spans: HighlightSpan[],
        textValue: string,
      ) => {
        if (!isMounted) {
          return;
        }
        setRenderedHtml(html);
        setTeiSpans(spans);
        setDisplayText(textValue);
        setIsReady(true);
      };

      const loadHighlights = async () => {
        setIsReady(false);

        if (!itemKey) {
          finalize(plainHtml, [], normalizedPlainText);
          return;
        }

        try {
          const tei = await getCachedOrFetchTei(itemKey);
          if (!tei) {
            finalize(plainHtml, [], normalizedPlainText);
            return;
          }

          const parsed = parseTeiToSpans(tei, Object.keys(featuresById));
          finalize(
            parsed?.baseHtml || plainHtml,
            parsed?.spans || [],
            parsed?.text || normalizedPlainText,
          );
        } catch {
          finalize(plainHtml, [], normalizedPlainText);
        }
      };

      loadHighlights();

      return () => {
        isMounted = false;
      };
    }, [featuresById, itemKey, normalizedPlainText, plainHtml]);

    const handleSelectionUpdate = useCallback(() => {
      setSelectionState(
        getSelectionState(editable, containerRef.current, normalizedText),
      );
    }, [editable, normalizedText]);

    const handleAnnotationRequest = () => {
      if (!selectionState || !onRequestAddAnnotation) {
        return;
      }
      onRequestAddAnnotation(selectionState);
      setSelectionState(null);
      window.getSelection()?.removeAllRanges();
    };

    const handleRemoveHighlight =
      editable && onRemoveHighlight
        ? (state: HighlightTooltipState) => {
            onRemoveHighlight({
              id: state.id,
              featureKey: state.featureKey,
              start: state.start,
              end: state.end,
              text: normalizedText.slice(state.start, state.end),
              label: state.label,
              normalized: state.normalized,
              color: state.color,
            });
            setTooltipState(null);
          }
        : undefined;

    const selectionTooltip =
      editable && selectionState && onRequestAddAnnotation ? (
        <div
          data-highlight-action
          style={{
            position: "fixed",
            left: selectionState.x,
            top: selectionState.y,
            transform: "translate(-50%, -100%)",
            background: "white",
            border: "1px solid #d9d9d9",
            boxShadow: "0 6px 16px rgba(0, 0, 0, 0.2)",
            borderRadius: "0.4rem",
            padding: "0.4rem 0.6rem",
            fontSize: "0.8rem",
            display: "flex",
            alignItems: "center",
            gap: "0.5rem",
            zIndex: 12001,
          }}
        >
          <span>Tag selection as annotation</span>
          <button
            type="button"
            onClick={handleAnnotationRequest}
            style={{
              border: "none",
              background: "#333",
              color: "white",
              padding: "0.25rem 0.5rem",
              borderRadius: "0.25rem",
              cursor: "pointer",
              fontSize: "0.8rem",
            }}
          >
            Tag
          </button>
        </div>
      ) : null;

    if (!isReady) {
      return (
        <div style={{ position: "relative", whiteSpace: "pre-wrap" }}>
          <div
            style={{ position: "relative" }}
            dangerouslySetInnerHTML={{ __html: plainHtml }}
          />
        </div>
      );
    }

    return (
      <div
        style={{ position: "relative", whiteSpace: "pre-wrap" }}
        onMouseMove={(event) => {
          if (
            (event.target as HTMLElement | null)?.closest?.(
              "[data-highlight-action]",
            )
          ) {
            return;
          }
          handleHighlightTooltipMouseMove(event, setTooltipState);
        }}
        onMouseLeave={() => {
          if (!tooltipPinned) {
            handleHighlightTooltipMouseLeave(setTooltipState);
          }
        }}
        onMouseUp={handleSelectionUpdate}
        onKeyUp={handleSelectionUpdate}
        ref={containerRef}
      >
        {renderedLayers.map((layer, index) => (
          <div
            key={index}
            data-highlight-layer={`layer-${index}`}
            style={{
              color: "transparent",
              position: "absolute",
              inset: 0,
              pointerEvents: "none",
              userSelect: "none",
              zIndex: index,
            }}
            dangerouslySetInnerHTML={{ __html: layer }}
          />
        ))}
        <div
          data-highlight-layer="base"
          style={{
            position: "relative",
            zIndex: renderedLayers.length,
            pointerEvents:
              editable || renderedLayers.length === 0 ? "auto" : "none",
          }}
          dangerouslySetInnerHTML={{ __html: renderedHtml }}
        />
        {editable &&
          renderedHitLayers.map((layer, index) => (
            <div
              key={`hit-${index}`}
              data-highlight-layer={`hit-${index}`}
              style={{
                color: "transparent",
                position: "absolute",
                inset: 0,
                pointerEvents: "auto",
                userSelect: "text",
                zIndex: renderedLayers.length + index + 1,
              }}
              dangerouslySetInnerHTML={{ __html: layer }}
            />
          ))}
        <FeatureHighlightTooltip
          tooltipState={
            tooltipState?.featureKey && featuresById[tooltipState.featureKey]
              ? tooltipState
              : null
          }
          onRemove={handleRemoveHighlight}
          onTooltipEnter={editable ? () => setTooltipPinned(true) : undefined}
          onTooltipLeave={
            editable
              ? () => {
                  setTooltipPinned(false);
                  setTooltipState(null);
                }
              : undefined
          }
        />
        {selectionTooltip}
      </div>
    );
  },
);
