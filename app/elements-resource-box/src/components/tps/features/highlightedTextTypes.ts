export type HighlightSpan = {
  id: string;
  start: number;
  end: number;
  featureKey: string;
  normalized: string;
  source: "tei" | "local";
};

export type HighlightSelection = {
  start: number;
  end: number;
  text: string;
  x: number;
  y: number;
};

export type HighlightAction = {
  id: string;
  featureKey: string;
  start: number;
  end: number;
  text: string;
  label: string;
  normalized: string;
  color: string;
};
