export type HighlightSpan = {
  id: string;
  start: number;
  end: number;
  featureKey: string;
  normalized: string;
  source: "tei";
};
