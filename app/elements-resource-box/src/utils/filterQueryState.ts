import { parseAsBoolean, parseAsJson, parseAsString } from "nuqs";
import { FilterValue } from "../components/map/Filter";
import { Item } from "../types";

export type FilterState = {
  filters: Record<string, FilterValue[] | undefined>;
  filtersInclude: Record<string, boolean>;
  range: [number, number];
  includeUndated: boolean;
  textSearch: string;
  textSearchFields: (keyof Item)[];
};

type NullableFilterQueryState = {
  filters: Record<string, FilterValue[] | undefined> | null;
  filtersInclude: Record<string, boolean> | null;
  range: [number, number] | null;
  includeUndated: boolean | null;
  textSearch: string | null;
  textSearchFields: (keyof Item)[] | null;
};

const isObjectRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

const isFilterValue = (value: unknown): value is FilterValue =>
  isObjectRecord(value) &&
  typeof value.label === "string" &&
  typeof value.value === "string";

const isFiltersState = (
  value: unknown,
): value is Record<string, FilterValue[] | undefined> => {
  if (!isObjectRecord(value)) {
    return false;
  }
  return Object.values(value).every(
    (entry) =>
      entry === undefined ||
      (Array.isArray(entry) && entry.every((item) => isFilterValue(item))),
  );
};

const isFiltersIncludeState = (
  value: unknown,
): value is Record<string, boolean> => {
  if (!isObjectRecord(value)) {
    return false;
  }
  return Object.values(value).every((entry) => typeof entry === "boolean");
};

const isRangeState = (value: unknown): value is [number, number] =>
  Array.isArray(value) &&
  value.length === 2 &&
  value.every((entry) => typeof entry === "number" && Number.isFinite(entry));

const isTextSearchFieldsState = (value: unknown): value is (keyof Item)[] =>
  Array.isArray(value) && value.every((entry) => typeof entry === "string");

export const filterQueryParsers = {
  filters: parseAsJson<Record<string, FilterValue[] | undefined>>((value) =>
    isFiltersState(value) ? value : null,
  ),
  filtersInclude: parseAsJson<Record<string, boolean>>((value) =>
    isFiltersIncludeState(value) ? value : null,
  ),
  range: parseAsJson<[number, number]>((value) =>
    isRangeState(value) ? value : null,
  ),
  includeUndated: parseAsBoolean,
  textSearch: parseAsString,
  textSearchFields: parseAsJson<(keyof Item)[]>((value) =>
    isTextSearchFieldsState(value) ? value : null,
  ),
};

export const mergeFilterQueryWithDefaults = (
  queryState: NullableFilterQueryState,
  defaults: FilterState,
): FilterState => ({
  filters: queryState.filters ?? defaults.filters,
  filtersInclude: queryState.filtersInclude ?? defaults.filtersInclude,
  range: queryState.range ?? defaults.range,
  includeUndated: queryState.includeUndated ?? defaults.includeUndated,
  textSearch: queryState.textSearch ?? defaults.textSearch,
  textSearchFields: queryState.textSearchFields ?? defaults.textSearchFields,
});
