import type { HighlightSpan } from "./highlightedTextTypes";
import { feature_Feature } from '@hub-api';

const OUTLINE_FEATURES = ["action_verbs"];

const escapeHtml = (value: string) =>
  value
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");

const escapeHtmlAttr = (value: string) =>
  escapeHtml(value).replaceAll('"', "&quot;").replaceAll("'", "&#39;");

const wrapNonInteractive = (value: string) =>
  value ? `<span style="pointer-events: none;">${value}</span>` : "";

export const buildTextHtml = (text: string) =>
  escapeHtml(text)
    .replaceAll("\n", "<br/>")
    .replace(/^(?:<br\/>|\s)+/, "")
    .replace(/(?:<br\/>|\s)+$/, "");

export const normalizeDisplayText = (value: string) => value.trim();

type DecoratedSpan = HighlightSpan & {
  length: number;
};

const sortSpansByLength = (spans: HighlightSpan[]): DecoratedSpan[] =>
  spans
    .map((span) => ({
      ...span,
      length: span.end - span.start,
    }))
    .sort((a, b) => b.length - a.length);

const buildSpanHtml = (
  trimmedText: string,
  span: HighlightSpan,
  style: string,
  featuresById: Record<string, feature_Feature>,
) => {
  const featureLabel = featuresById[span.featureKey].name || span.featureKey;
  const tooltip = escapeHtmlAttr(featureLabel);
  const tooltipColor = escapeHtmlAttr(
    featuresById[span.featureKey].color || "#f2f2f2",
  );
  const tooltipNormalized = escapeHtmlAttr(span.normalized);
  const tooltipNormalizedAttr = tooltipNormalized
    ? ` data-feature-normalized="${tooltipNormalized}"`
    : "";
  const featureKeyAttr = escapeHtmlAttr(span.featureKey);
  const highlightIdAttr = escapeHtmlAttr(span.id);

  const before = trimmedText.slice(0, span.start);
  const segment = trimmedText.slice(span.start, span.end);
  const after = trimmedText.slice(span.end);

  let html =
    wrapNonInteractive(escapeHtml(before)) +
    `<span style="${style}" data-feature-label="${tooltip}" data-feature-color="${tooltipColor}" data-feature-key="${featureKeyAttr}" data-feature-start="${span.start}" data-feature-end="${span.end}" data-highlight-id="${highlightIdAttr}"${tooltipNormalizedAttr}>${escapeHtml(segment)}</span>` +
    wrapNonInteractive(escapeHtml(after));
  html = html.replaceAll("\n", "<br/>");
  html = html.replace(/^(?:<br\/>|\s)+/, "");
  html = html.replace(/(?:<br\/>|\s)+$/, "");
  return html;
};

export const buildHighlightLayers = (
  trimmedText: string,
  spans: HighlightSpan[],
  featuresById: Record<string, feature_Feature>,
) => {
  if (
    !trimmedText ||
    spans.length === 0 ||
    !featuresById ||
    Object.keys(featuresById).length === 0
  ) {
    return [];
  }

  const sorted = sortSpansByLength(spans);

  return sorted
    .map((span, layerIndex) => {
      const color = featuresById[span.featureKey].color;
      if (!color) {
        return "white";
      }
      const useOutline = OUTLINE_FEATURES.includes(span.featureKey);
      const depth = sorted.reduce((count, other, idx) => {
        if (idx <= layerIndex) return count;
        if (span.start < other.end && span.end > other.start) {
          return count + 1;
        }
        return count;
      }, 0);
      const shadowSize = Math.min(6, 2 + depth * 2);
      const style = useOutline
        ? `outline: 2px solid ${color}; outline-offset: 2px; border-radius: 8px; pointer-events: auto;`
        : `background-color: ${color}; box-shadow: 0 0 0 ${shadowSize}px ${color}; border-radius: 8px; pointer-events: auto;`;
      return buildSpanHtml(trimmedText, span, style, featuresById);
    })
    .filter(Boolean);
};

export const buildHighlightHitLayers = (
  trimmedText: string,
  spans: HighlightSpan[],
  featuresById: Record<string, feature_Feature>,
) => {
  if (
    !trimmedText ||
    spans.length === 0 ||
    !featuresById ||
    Object.keys(featuresById).length === 0
  ) {
    return [];
  }

  const style =
    "pointer-events: auto; background-color: transparent; box-shadow: none; border-radius: 8px; outline: none;";

  return sortSpansByLength(spans)
    .map((span) => buildSpanHtml(trimmedText, span, style, featuresById))
    .filter(Boolean);
};
