export type FeatureEditState = {
  name: string;
  description: string;
  color: string;
};

export type RevisionFormState = {
  execution_strategy: "prompt" | "regex";
  type: "annotation" | "ner";
  prompt: string;
  regex: string;
  note: string;
};

export type ActiveTab = "definitions" | "executions";

export const STUDY_CORPORA_FILTER = "Title pages";

export const defaultRevisionForm: RevisionFormState = {
  execution_strategy: "prompt",
  type: "annotation",
  prompt: "",
  regex: "",
  note: "",
};
