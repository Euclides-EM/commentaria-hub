import type { model_Pseudonym } from "@hub-api";

type PseudonymSource = "Editors" | "Publishers";

type PseudonymEntry = {
  name: string;
  aliases: Set<string>;
  tupleAliases: Set<string>;
  normalizedAliases: Set<string>;
};

type PseudonymSourceIndex = Map<string, PseudonymEntry>;

export type PseudonymAutocompleteIndex = Partial<
  Record<PseudonymSource, PseudonymSourceIndex>
>;

const normalizeText = (value: string) =>
  value
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, " ")
    .trim()
    .replace(/\s+/g, " ");

const addAlias = (target: Set<string>, value?: string) => {
  if (!value) {
    return;
  }
  const normalized = normalizeText(value);
  if (!normalized) {
    return;
  }
  target.add(normalized);
};

export const buildPseudonymAutocompleteIndex = (
  pseudonyms: model_Pseudonym[] = [],
): PseudonymAutocompleteIndex => {
  const index: PseudonymAutocompleteIndex = {};
  const grouped = new Map<
    PseudonymSource,
    Map<
      string,
      {
        name: string;
        firstOrOther: Set<string>;
        lastOrOther: Set<string>;
        aliases: Set<string>;
      }
    >
  >();

  for (const pseudonym of pseudonyms) {
    const source = pseudonym.source as PseudonymSource | undefined;
    const name = pseudonym.name?.trim();
    if (!source || !name || (source !== "Editors" && source !== "Publishers")) {
      continue;
    }
    const groupForSource = grouped.get(source) ?? new Map();
    grouped.set(source, groupForSource);
    const person = groupForSource.get(name) ?? {
      name,
      firstOrOther: new Set<string>(),
      lastOrOther: new Set<string>(),
      aliases: new Set<string>(),
    };
    groupForSource.set(name, person);
    addAlias(person.aliases, name);
    addAlias(person.aliases, pseudonym.pseudonym);
    if (pseudonym.position === "first" || pseudonym.position === "other") {
      addAlias(person.firstOrOther, pseudonym.pseudonym);
    }
    if (pseudonym.position === "last" || pseudonym.position === "other") {
      addAlias(person.lastOrOther, pseudonym.pseudonym);
    }
  }

  for (const [source, people] of grouped) {
    const sourceIndex = new Map<string, PseudonymEntry>();
    for (const [name, person] of people) {
      const tupleAliases = new Set<string>();
      for (const left of person.firstOrOther) {
        for (const right of person.lastOrOther) {
          addAlias(tupleAliases, `${left} ${right}`);
          addAlias(tupleAliases, `${right} ${left}`);
        }
      }
      sourceIndex.set(name, {
        name,
        aliases: person.aliases,
        tupleAliases,
        normalizedAliases: new Set([...person.aliases, ...tupleAliases]),
      });
    }
    index[source] = sourceIndex;
  }

  return index;
};

export const createPseudonymAutocompleteMatcher = (
  pseudonymIndex: PseudonymAutocompleteIndex | undefined,
  source: PseudonymSource,
) => {
  const sourceIndex = pseudonymIndex?.[source];
  return (optionValue: string, inputValue: string) => {
    const normalizedInput = normalizeText(inputValue);
    if (!normalizedInput) {
      return true;
    }
    const normalizedOption = normalizeText(optionValue);
    if (normalizedOption.includes(normalizedInput)) {
      return true;
    }
    const person = sourceIndex?.get(optionValue);
    if (!person) {
      return false;
    }
    if (normalizeText(person.name).includes(normalizedInput)) {
      return true;
    }
    for (const alias of person.normalizedAliases) {
      if (alias.includes(normalizedInput)) {
        return true;
      }
    }
    return false;
  };
};
