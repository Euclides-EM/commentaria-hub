import type { HighlightSelection } from "./highlightedTextTypes";

const HIGHLIGHT_LAYER_SELECTOR = "[data-highlight-layer]";

const getLayerContainer = (node: Node) => {
  if (node instanceof Element) {
    return node.closest(HIGHLIGHT_LAYER_SELECTOR);
  }
  return node.parentElement?.closest(HIGHLIGHT_LAYER_SELECTOR) || null;
};

const getOffsetInContainer = (
  layerContainer: Element,
  targetNode: Node,
  offset: number,
) => {
  const preRange = document.createRange();
  preRange.selectNodeContents(layerContainer);
  preRange.setEnd(targetNode, offset);
  return preRange.toString().length;
};

const getClosestMatchStart = (
  normalizedText: string,
  rawSelectedText: string,
  sliceStart: number,
) => {
  const searchWindow = 120;
  const windowStart = Math.max(0, sliceStart - searchWindow);
  const windowEnd = Math.min(normalizedText.length, sliceStart + searchWindow);
  const windowText = normalizedText.slice(windowStart, windowEnd);

  let bestIndex = -1;
  let bestDistance = Number.POSITIVE_INFINITY;
  let matchIndex = windowText.indexOf(rawSelectedText);

  while (matchIndex !== -1) {
    const absoluteIndex = windowStart + matchIndex;
    const distance = Math.abs(absoluteIndex - sliceStart);
    if (distance < bestDistance) {
      bestDistance = distance;
      bestIndex = absoluteIndex;
    }
    matchIndex = windowText.indexOf(rawSelectedText, matchIndex + 1);
  }

  return bestIndex;
};

export const getSelectionState = (
  editable: boolean,
  container: HTMLDivElement | null,
  normalizedText: string,
): HighlightSelection | null => {
  if (!editable || !container) {
    return null;
  }

  const selection = window.getSelection();
  if (!selection || selection.rangeCount === 0) {
    return null;
  }

  const range = selection.getRangeAt(0);
  if (range.collapsed || !container.contains(range.commonAncestorContainer)) {
    return null;
  }

  const startContainer = getLayerContainer(range.startContainer);
  const endContainer = getLayerContainer(range.endContainer);
  if (!startContainer || !endContainer) {
    return null;
  }

  const start = getOffsetInContainer(
    startContainer,
    range.startContainer,
    range.startOffset,
  );
  const end = getOffsetInContainer(
    endContainer,
    range.endContainer,
    range.endOffset,
  );
  if (start === end) {
    return null;
  }

  const rawSelectedText = selection.toString();
  const sliceStart = Math.min(start, end);
  const sliceEnd = Math.max(start, end);
  let selectedText = normalizedText.slice(sliceStart, sliceEnd);
  let adjustedStart = sliceStart;

  if (rawSelectedText && selectedText !== rawSelectedText) {
    const bestIndex = getClosestMatchStart(
      normalizedText,
      rawSelectedText,
      sliceStart,
    );
    if (bestIndex !== -1) {
      adjustedStart = bestIndex;
      selectedText = rawSelectedText;
    }
  }

  if (!selectedText.trim()) {
    return null;
  }

  const rect = range.getBoundingClientRect();
  return {
    start: adjustedStart,
    end: adjustedStart + selectedText.length,
    text: selectedText,
    x: rect.left + rect.width / 2,
    y: rect.top - 8,
  };
};
