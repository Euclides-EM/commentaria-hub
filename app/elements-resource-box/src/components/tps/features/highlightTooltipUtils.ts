import { Dispatch, MouseEvent as ReactMouseEvent, SetStateAction } from "react";

export type HighlightTooltipState = {
  label: string;
  normalized: string;
  color: string;
  featureKey: string;
  start: number;
  end: number;
  id: string;
  x: number;
  y: number;
};

type HighlightTooltipInfo = {
  label: string;
  normalized: string;
  color: string;
  featureKey: string;
  start: number;
  end: number;
  id: string;
};

const TOOLTIP_OFFSET_X = 12;
const TOOLTIP_OFFSET_Y = 14;
const FEATURE_LABEL_SELECTOR = "[data-feature-label]";
const DEFAULT_FEATURE_COLOR = "#f2f2f2";

const getTooltipInfoFromElement = (
  element: Element | null,
): HighlightTooltipInfo | null => {
  if (!element) {
    return null;
  }
  const highlighted = element.closest(FEATURE_LABEL_SELECTOR);
  if (!(highlighted instanceof HTMLElement)) {
    return null;
  }
  const label = highlighted.dataset.featureLabel || "";
  if (!label) {
    return null;
  }
  const featureKey = highlighted.dataset.featureKey || "";
  const start = Number.parseInt(highlighted.dataset.featureStart || "", 10);
  const end = Number.parseInt(highlighted.dataset.featureEnd || "", 10);
  const id = highlighted.dataset.highlightId || "";
  if (!featureKey || Number.isNaN(start) || Number.isNaN(end) || !id) {
    return null;
  }
  return {
    label,
    normalized: highlighted.dataset.featureNormalized || "",
    color: highlighted.dataset.featureColor || DEFAULT_FEATURE_COLOR,
    featureKey,
    start,
    end,
    id,
  };
};

const getTooltipPositionFromElement = (
  element: Element,
): { x: number; y: number } => {
  const rect = element.getBoundingClientRect();
  return {
    x: rect.left + rect.width / 2 + TOOLTIP_OFFSET_X,
    y: rect.top + TOOLTIP_OFFSET_Y,
  };
};

export const handleHighlightTooltipMouseMove = (
  event: ReactMouseEvent<HTMLElement>,
  setTooltipState: Dispatch<SetStateAction<HighlightTooltipState | null>>,
) => {
  const elements = document.elementsFromPoint(event.clientX, event.clientY);
  const hitElement =
    elements.find((el) => el.matches?.(FEATURE_LABEL_SELECTOR)) ||
    elements.find((el) => el.closest?.(FEATURE_LABEL_SELECTOR)) ||
    (event.target instanceof Element ? event.target : null);
  const info = getTooltipInfoFromElement(hitElement);
  if (!info) {
    setTooltipState((prev) => (prev ? null : prev));
    return;
  }
  const position = hitElement
    ? getTooltipPositionFromElement(hitElement)
    : {
        x: event.clientX + TOOLTIP_OFFSET_X,
        y: event.clientY + TOOLTIP_OFFSET_Y,
      };
  setTooltipState((prev) => {
    if (
      prev &&
      prev.label === info.label &&
      prev.normalized === info.normalized &&
      prev.color === info.color &&
      prev.featureKey === info.featureKey &&
      prev.start === info.start &&
      prev.end === info.end &&
      prev.id === info.id &&
      prev.x === position.x &&
      prev.y === position.y
    ) {
      return prev;
    }
    return {
      label: info.label,
      normalized: info.normalized,
      color: info.color,
      featureKey: info.featureKey,
      start: info.start,
      end: info.end,
      id: info.id,
      x: position.x,
      y: position.y,
    };
  });
};

export const handleHighlightTooltipMouseLeave = (
  setTooltipState: Dispatch<SetStateAction<HighlightTooltipState | null>>,
) => {
  setTooltipState(null);
};
