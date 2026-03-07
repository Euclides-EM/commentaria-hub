import { memo, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
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
import { parseTeiToSpans } from "./highlightedTextTeiUtils";
import type { HighlightSpan } from "./highlightedTextTypes";
import { AnnotationsService, feature_Feature } from "@hub-api";
import {
  TITLE_PAGES_ANNOTATION_ID,
  TITLE_PAGES_DATASET_ID,
} from "../../../constants";

export type { HighlightSpan };

type HighlightedTextProps = {
  text: string;
  featuresById: Record<string, feature_Feature>;
  itemKey: string;
};

export const HighlightedText = memo(
  ({ text, featuresById, itemKey }: HighlightedTextProps) => {
    const [tooltipState, setTooltipState] =
      useState<HighlightTooltipState | null>(null);
    const [tooltipPinned, setTooltipPinned] = useState(false);

    const plainHtml = useMemo(() => buildTextHtml(text), [text]);
    const normalizedPlainText = useMemo(
      () => normalizeDisplayText(text),
      [text],
    );
    const teiQuery = useQuery({
      queryKey: ["tei", itemKey],
      queryFn: () =>
        AnnotationsService.getDatasetsAnnotationsTei({
          dataSetId: TITLE_PAGES_DATASET_ID,
          id: TITLE_PAGES_ANNOTATION_ID,
          pageNumOrKey: itemKey,
        }),
      enabled: Boolean(itemKey),
    });
    const parsedTei = useMemo(() => {
      if (!teiQuery.data) {
        return null;
      }
      return parseTeiToSpans(teiQuery.data, featuresById);
    }, [featuresById, teiQuery.data]);
    const renderedHtml = parsedTei?.baseHtml || plainHtml;
    const teiSpans = useMemo(() => parsedTei?.spans || [], [parsedTei]);
    const displayText = parsedTei?.text || normalizedPlainText;
    const isReady = !itemKey || !teiQuery.isPending;

    const combinedSpans = useMemo(() => {
      const selectedSet = new Set(Object.keys(featuresById));
      const normalizedText = displayText || normalizedPlainText;

      return teiSpans
        .map((span) => ({
          ...span,
          start: Math.max(0, Math.min(span.start, normalizedText.length)),
          end: Math.max(0, Math.min(span.end, normalizedText.length)),
        }))
        .filter((span) => span.end > span.start)
        .filter(
          (span) => selectedSet.size === 0 || selectedSet.has(span.featureKey),
        );
    }, [displayText, featuresById, normalizedPlainText, teiSpans]);

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
            pointerEvents: renderedLayers.length === 0 ? "auto" : "none",
          }}
          dangerouslySetInnerHTML={{ __html: renderedHtml }}
        />
        {renderedHitLayers.map((layer, index) => (
          <div
            key={`hit-${index}`}
            data-highlight-layer={`hit-${index}`}
            style={{
              color: "transparent",
              position: "absolute",
              inset: 0,
              pointerEvents: "auto",
              userSelect: "none",
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
          onTooltipEnter={() => setTooltipPinned(true)}
          onTooltipLeave={() => {
            setTooltipPinned(false);
            setTooltipState(null);
          }}
        />
      </div>
    );
  },
);
