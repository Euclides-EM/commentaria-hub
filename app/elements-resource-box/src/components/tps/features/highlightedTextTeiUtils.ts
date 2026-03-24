import { buildTextHtml } from "./highlightedTextRenderUtils";
import type { HighlightSpan } from "./highlightedTextTypes";
import type { feature_Feature } from "@hub-api";

const TEI_NS = "http://www.tei-c.org/ns/1.0";
const XML_NS = "http://www.w3.org/XML/1998/namespace";

const normalizeKey = (value: string) =>
  value.toLowerCase().replace(/[^a-z0-9]+/g, "");

const parseAnaRefs = (value: string | null) =>
  (value || "")
    .split(/\s+/)
    .map((entry) => entry.replace(/^#/, "").trim())
    .filter(Boolean);

const getXmlId = (element: Element) =>
  element.getAttributeNS(XML_NS, "id") ||
  element.getAttribute("xml:id") ||
  element.getAttribute("id") ||
  "";

const resolveFeatureKey = (
  categoryId: string,
  categoryLabel: string,
  featuresById: Record<string, feature_Feature>,
) => {
  const selectedFeatures = Object.keys(featuresById);
  if (selectedFeatures.length === 0) {
    return categoryId;
  }
  const byNormalized = new Map<string, string>();
  selectedFeatures.forEach((featureKey) => {
    const feature = featuresById[featureKey];
    [featureKey, feature?.name || ""]
      .map((value) => normalizeKey(value))
      .filter(Boolean)
      .forEach((value) => byNormalized.set(value, featureKey));
  });

  const candidates = [
    categoryId,
    categoryId.replace(/^cat_/, ""),
    categoryLabel,
  ]
    .map((candidate) => normalizeKey(candidate))
    .filter(Boolean);

  for (const candidate of candidates) {
    const matched = byNormalized.get(candidate);
    if (matched) {
      return matched;
    }
  }

  return categoryId;
};

export const parseTeiToSpans = (
  tei: string,
  featuresById: Record<string, feature_Feature>,
): { baseHtml: string; spans: HighlightSpan[]; text: string } | null => {
  const parser = new DOMParser();
  const doc = parser.parseFromString(tei, "text/xml");
  const parseError = doc.getElementsByTagName("parsererror")[0];
  if (parseError) {
    return null;
  }

  const selectedSet = new Set(Object.keys(featuresById));
  const body = doc.getElementsByTagNameNS(TEI_NS, "body")[0];
  if (!body) {
    return null;
  }

  const categoryMap = new Map<string, string>();
  const interpGroups = doc.getElementsByTagNameNS(TEI_NS, "interpGrp");
  Array.from(interpGroups).forEach((group) => {
    if (group.getAttribute("type") !== "highlight-categories") {
      return;
    }
    const interps = group.getElementsByTagNameNS(TEI_NS, "interp");
    Array.from(interps).forEach((interp) => {
      const id = getXmlId(interp);
      if (!id) return;
      categoryMap.set(id, interp.textContent?.trim() || id);
    });
  });

  const anchorPos: Record<string, number> = {};
  let rawText = "";
  const startsWithClosingPunctuation = (value: string) =>
    /^[\s]*[.,;:!?)\]\}]/.test(value);
  const trimTrailingSpaces = () => {
    rawText = rawText.replace(/[ \t]+$/g, "");
    const nextLength = rawText.length;
    Object.keys(anchorPos).forEach((anchorId) => {
      if (anchorPos[anchorId] > nextLength) {
        anchorPos[anchorId] = nextLength;
      }
    });
  };
  const appendText = (value: string) => {
    if (!value) {
      return;
    }
    let normalized = value.replace(/\s+/g, " ");
    if (!normalized) {
      return;
    }
    if (startsWithClosingPunctuation(normalized)) {
      trimTrailingSpaces();
      normalized = normalized.replace(/^\s+/, "");
    }
    rawText += normalized;
  };
  const appendLineBreak = () => {
    trimTrailingSpaces();
    if (!rawText.endsWith("\n")) {
      rawText += "\n";
    }
  };
  const transcriptionDivs = Array.from(
    body.getElementsByTagNameNS(TEI_NS, "div"),
  ).filter((div) => div.getAttribute("type") === "transcription");
  const roots = transcriptionDivs.length > 0 ? transcriptionDivs : [body];

  const walkNode = (node: ChildNode) => {
    if (node.nodeType === Node.TEXT_NODE) {
      appendText(node.textContent || "");
      return;
    }
    if (node.nodeType !== Node.ELEMENT_NODE) {
      return;
    }
    const element = node as Element;
    const name = element.localName;

    if (name === "anchor") {
      const anchorId = getXmlId(element);
      if (anchorId) {
        anchorPos[anchorId] = rawText.length;
      }
      return;
    }

    if (name === "lb") {
      appendLineBreak();
      return;
    }

    element.childNodes.forEach(walkNode);

    if (name === "p" || name === "l") {
      appendLineBreak();
    }
  };

  roots.forEach((root) => {
    root.childNodes.forEach(walkNode);
  });

  const spanGroups = doc.getElementsByTagNameNS(TEI_NS, "spanGrp");
  const filteredSpans: Array<{
    start: number;
    end: number;
    featureKey: string;
    normalized: string;
  }> = [];

  Array.from(spanGroups).forEach((group) => {
    const type = group.getAttribute("type");
    if (type !== "highlight" && type !== "highlights") {
      return;
    }

    const spans = group.getElementsByTagNameNS(TEI_NS, "span");
    Array.from(spans).forEach((span) => {
      const from = span.getAttribute("from")?.replace("#", "");
      const to = span.getAttribute("to")?.replace("#", "");
      if (!from || !to) return;

      const startIdx = anchorPos[from];
      const endIdx = anchorPos[to];
      if (startIdx === undefined || endIdx === undefined) {
        return;
      }

      const anaRefs = parseAnaRefs(span.getAttribute("ana"));
      const categoryId =
        anaRefs.find((ref) => ref.startsWith("cat_")) ||
        anaRefs[0] ||
        "uncategorized";
      const categoryLabel = categoryMap.get(categoryId) || categoryId;
      const featureKey = resolveFeatureKey(
        categoryId,
        categoryLabel,
        featuresById,
      );
      if (!featureKey) {
        return;
      }

      const start = Math.min(startIdx, endIdx);
      const end = Math.max(startIdx, endIdx);
      if (end <= start) {
        return;
      }

      const notes = span.getElementsByTagNameNS(TEI_NS, "note");
      let normalized = "";
      Array.from(notes).forEach((note) => {
        if (normalized) return;
        const noteAnaRefs = parseAnaRefs(note.getAttribute("ana"));
        const isNormalized =
          noteAnaRefs.includes("prop_normalized") ||
          noteAnaRefs.some((ref) => ref.endsWith("normalized"));
        if (!isNormalized) return;
        normalized = note.textContent?.trim() || "";
      });

      if (selectedSet.size === 0 || selectedSet.has(featureKey)) {
        filteredSpans.push({ start, end, featureKey, normalized });
      }
    });
  });

  const leadingWhitespace = rawText.match(/^\s*/)?.[0].length ?? 0;
  const trimmedText = rawText.trim();
  const baseHtml = buildTextHtml(trimmedText);

  const spans: HighlightSpan[] = filteredSpans
    .map((span) => ({
      ...span,
      start: span.start - leadingWhitespace,
      end: span.end - leadingWhitespace,
    }))
    .filter((span) => span.end > 0 && span.start < trimmedText.length)
    .map((span) => ({
      ...span,
      start: Math.max(0, span.start),
      end: Math.min(trimmedText.length, span.end),
    }))
    .map((span) => {
      let start = span.start;
      let end = span.end;
      while (start < end && /\s/.test(trimmedText[start])) {
        start += 1;
      }
      while (end > start && /\s/.test(trimmedText[end - 1])) {
        end -= 1;
      }
      return { ...span, start, end };
    })
    .filter((span) => span.end > span.start)
    .map((span) => ({
      ...span,
      id: `${span.featureKey}:${span.start}-${span.end}:${span.normalized || "raw"}`,
      source: "tei" as const,
    }));

  return { baseHtml, spans, text: trimmedText };
};
