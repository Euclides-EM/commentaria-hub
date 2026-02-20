import { RevisionFormState } from "./types";

export const getRevisionDefaults = (revision?: {
  execution_strategy?: "prompt" | "regex";
  type?: "annotation" | "ner";
  prompt?: string;
  regex?: string;
  note?: string;
}): RevisionFormState => ({
  execution_strategy: revision?.execution_strategy ?? "prompt",
  type: revision?.type ?? "annotation",
  prompt: revision?.prompt ?? "",
  regex: revision?.regex ?? "",
  note: revision?.note ?? "",
});

export const formatDate = (value?: string) => {
  if (!value) {
    return "Unknown";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
};
