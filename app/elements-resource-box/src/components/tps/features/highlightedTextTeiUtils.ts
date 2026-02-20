import { TITLE_PAGES_DATASET_ID } from "../../../constants";
import { buildTextHtml } from "./highlightedTextRenderUtils";
import type { HighlightSpan } from "./highlightedTextTypes";
import { AnnotationsService } from '@hub-api';

const TEI_NS = "http://www.tei-c.org/ns/1.0";

const teiCache = new Map<string, string>();
const teiPromiseCache = new Map<string, Promise<string>>();

const getFeatureFromSpanId = (spanId: string) => {
  if (!spanId.startsWith("h-")) {
    return "";
  }
  const lastDash = spanId.lastIndexOf("-");
  if (lastDash <= 2) {
    return "";
  }
  return spanId.slice(2, lastDash);
};

export const getCachedOrFetchTei = async (itemKey: string) => {
  const cachedTei = teiCache.get(itemKey);
  if (cachedTei) {
    return cachedTei;
  }

  const inFlight =
    teiPromiseCache.get(itemKey) ||
    AnnotationsService.getDatasetsAnnotationsTei({
      dataSetId: TITLE_PAGES_DATASET_ID,
      id: itemKey,
    });
  teiPromiseCache.set(itemKey, inFlight);

  const tei = await inFlight;
  if (tei) {
    teiCache.set(itemKey, tei);
  }
  return tei;
};

export const parseTeiToSpans = (
  tei: string,
  selectedFeatures: string[],
): { baseHtml: string; spans: HighlightSpan[]; text: string } | null => {
  const parser = new DOMParser();
  const doc = parser.parseFromString(tei, "text/xml");
  const parseError = doc.getElementsByTagName("parsererror")[0];
  if (parseError) {
    return null;
  }

  const respStatements = doc.getElementsByTagNameNS(TEI_NS, "respStmt");
  const respToFeature: Record<string, string> = {};
  Array.from(respStatements).forEach((respStmt) => {
    const respId =
      respStmt.getAttribute("xml:id") || respStmt.getAttribute("id");
    if (!respId) return;
    const idnos = respStmt.getElementsByTagNameNS(TEI_NS, "idno");
    const featureIdno = Array.from(idnos).find(
      (node) => node.getAttribute("type") === "feature",
    );
    const featureKey = featureIdno?.textContent?.trim();
    if (featureKey) {
      respToFeature[respId] = featureKey;
    }
  });

  const selectedSet = new Set(selectedFeatures);
  const body = doc.getElementsByTagNameNS(TEI_NS, "body")[0];
  if (!body) {
    return null;
  }

  const anchorPos: Record<string, number> = {};
  let rawText = "";

  const walkNode = (node: ChildNode) => {
    if (node.nodeType === Node.TEXT_NODE) {
      rawText += node.textContent || "";
      return;
    }
    if (node.nodeType !== Node.ELEMENT_NODE) {
      return;
    }
    const element = node as Element;
    const name = element.localName;

    if (name === "anchor") {
      const anchorId =
        element.getAttribute("xml:id") || element.getAttribute("id");
      if (anchorId) {
        anchorPos[anchorId] = rawText.length;
      }
      return;
    }

    if (name === "lb") {
      rawText += "\n";
      return;
    }

    element.childNodes.forEach(walkNode);

    if (name === "p") {
      rawText += "\n";
    }
  };

  body.childNodes.forEach(walkNode);

  const spanGroups = doc.getElementsByTagNameNS(TEI_NS, "spanGrp");
  const filteredSpans: Array<{
    start: number;
    end: number;
    featureKey: string;
    normalized: string;
  }> = [];

  Array.from(spanGroups).forEach((group) => {
    if (group.getAttribute("type") !== "highlight") {
      return;
    }

    const spans = group.getElementsByTagNameNS(TEI_NS, "span");
    Array.from(spans).forEach((span) => {
      const from = span.getAttribute("from")?.replace("#", "");
      const to = span.getAttribute("to")?.replace("#", "");
      const resp = span.getAttribute("resp")?.replace("#", "");
      if (!from || !to || !resp) return;

      const startIdx = anchorPos[from];
      const endIdx = anchorPos[to];
      if (startIdx === undefined || endIdx === undefined) {
        return;
      }

      const spanId =
        span.getAttribute("xml:id") || span.getAttribute("id") || "";
      const featureKey =
        getFeatureFromSpanId(spanId) || respToFeature[resp] || "";
      if (!featureKey) {
        return;
      }

      const start = Math.min(startIdx, endIdx);
      const end = Math.max(startIdx, endIdx);
      if (end <= start) {
        return;
      }

      const normalized =
        Array.from(span.getElementsByTagNameNS(TEI_NS, "f"))
          .find((node) => node.getAttribute("name") === "normalized")
          ?.textContent?.trim() || "";

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
    .filter((span) => span.end > span.start)
    .map((span) => ({
      ...span,
      id: `${span.featureKey}:${span.start}-${span.end}:${span.normalized || "raw"}`,
      source: "tei" as const,
    }));

  return { baseHtml, spans, text: trimmedText };
};
