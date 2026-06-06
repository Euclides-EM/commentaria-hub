import React, {
  createContext,
  ReactNode,
  useContext,
  useMemo,
  useRef,
  useState,
  useCallback,
} from "react";
import { Item, MAX_YEAR, MIN_YEAR, MIN_YEAR_MS } from "../types";
import { FilterValue } from "../components/map/Filter";
import { mapEditionsToItems } from "../utils/dataUtils";
import {
  FilterState,
  filterQueryParsers,
  getFilterStateSignature,
  mergeFilterQueryWithDefaults,
} from "../utils/filterQueryState";
import { useQueryStates } from "nuqs";
import { useQuery } from "@tanstack/react-query";
import { listAllEditions } from "../api/editionApi";
import { injectEuclidesEditionConstraints } from "../utils/editionSearchQuery";
import { inEuclidesMode } from "../utils/mode";

export type { FilterState } from "../utils/filterQueryState";

type FilterAppliedContextType = {
  data: Item[];
  filters: Record<string, FilterValue[] | undefined>;
  filtersInclude: Record<string, boolean>;
  range: [number, number];
  includeUndated: boolean;
  textSearch: string;
  textSearchFields: (keyof Item)[];
  minYear: number;
  maxYear: number;
  applyFilters: (filterState: FilterState) => void;
  applyRange: (range: [number, number]) => void;
  resetFilters: (setters: {
    setFilters: React.Dispatch<
      React.SetStateAction<Record<string, FilterValue[] | undefined>>
    >;
    setFiltersInclude: React.Dispatch<
      React.SetStateAction<Record<string, boolean>>
    >;
    setRange: React.Dispatch<React.SetStateAction<[number, number]>>;
    setIncludeUndated: React.Dispatch<React.SetStateAction<boolean>>;
    setTextSearch: React.Dispatch<React.SetStateAction<string>>;
    setTextSearchFields: React.Dispatch<React.SetStateAction<(keyof Item)[]>>;
  }) => void;
  updateHasUnappliedChanges: (hasChanges: boolean) => void;
  hasUnappliedChanges: boolean;
};

const FilterAppliedContext = createContext<
  FilterAppliedContextType | undefined
>(undefined);

export const useAppliedFilter = () => {
  const context = useContext(FilterAppliedContext);
  if (!context) {
    throw new Error(
      "useAppliedFilter must be used within a FilterAppliedProvider",
    );
  }
  return context;
};

const includesManuscripts = (
  filters: Record<string, FilterValue[] | undefined>,
  filtersInclude: Record<string, boolean>,
) => {
  const materialType = filters.materialType;
  const include = filtersInclude.materialType ?? true;

  if (!materialType || materialType.length === 0) {
    return true;
  }

  const values = new Set(materialType.map((item) => item.value));
  const hasManuscript = values.has("Manuscript");
  const hasPrint = values.has("Print");

  if (include) {
    return hasManuscript;
  }

  return !hasPrint;
};

export const FilterAppliedProvider = ({
  children,
}: {
  children: ReactNode;
}) => {
  const editionsQuery = useQuery({
    queryKey: ["editions", "all", "filter-applied"],
    queryFn: () => listAllEditions(),
  });
  const data = useMemo<Item[]>(
    () => mapEditionsToItems(editionsQuery.data || []),
    [editionsQuery.data],
  );

  const getDefaultState = useCallback((): FilterState => {
    return {
      filters: {
        ...(inEuclidesMode()
          ? {}
          : {
              materialType: [
                {
                  label: "Print",
                  value: "Print",
                },
              ],
            }),
        type: [
          {
            label: "Elements",
            value: "Elements",
          },
        ],
      } as Record<string, FilterValue[] | undefined>,
      filtersInclude: {},
      range: [MIN_YEAR, MAX_YEAR] as [number, number],
      includeUndated: true,
      textSearch: "",
      textSearchFields: [
        "key",
        "shortTitle",
        "title",
        "titleEn",
      ] as (keyof Item)[],
    };
  }, []);
  const [queryFilters, setQueryFilters] = useQueryStates(filterQueryParsers, {
    history: "replace",
  });
  const mergedAppliedFilters = useMemo(
    () => mergeFilterQueryWithDefaults(queryFilters, getDefaultState()),
    [queryFilters, getDefaultState],
  );
  const appliedFiltersSignature = useMemo(
    () => getFilterStateSignature(mergedAppliedFilters),
    [mergedAppliedFilters],
  );
  const stableAppliedFiltersRef = useRef<{
    signature: string;
    value: FilterState;
  } | null>(null);
  if (
    !stableAppliedFiltersRef.current ||
    stableAppliedFiltersRef.current.signature !== appliedFiltersSignature
  ) {
    stableAppliedFiltersRef.current = {
      signature: appliedFiltersSignature,
      value: mergedAppliedFilters,
    };
  }
  const appliedFilters = stableAppliedFiltersRef.current.value;
  const minYear = useMemo(
    () =>
      !inEuclidesMode() &&
      includesManuscripts(appliedFilters.filters, appliedFilters.filtersInclude)
        ? MIN_YEAR_MS
        : MIN_YEAR,
    [appliedFilters.filters, appliedFilters.filtersInclude],
  );
  const maxYear = MAX_YEAR;

  const [hasUnappliedChanges, setHasUnappliedChanges] = useState(false);

  const resetFilters = useCallback(
    (setters: {
      setFilters: React.Dispatch<
        React.SetStateAction<Record<string, FilterValue[] | undefined>>
      >;
      setFiltersInclude: React.Dispatch<
        React.SetStateAction<Record<string, boolean>>
      >;
      setRange: React.Dispatch<React.SetStateAction<[number, number]>>;
      setIncludeUndated: React.Dispatch<React.SetStateAction<boolean>>;
      setTextSearch: React.Dispatch<React.SetStateAction<string>>;
      setTextSearchFields: React.Dispatch<React.SetStateAction<(keyof Item)[]>>;
    }) => {
      const defaultState = getDefaultState();

      setters.setFilters(() => defaultState.filters);
      setters.setFiltersInclude(() => ({}));
      setters.setRange(defaultState.range);
      setters.setIncludeUndated(defaultState.includeUndated);
      setters.setTextSearch(defaultState.textSearch);
      setters.setTextSearchFields(defaultState.textSearchFields);
      setQueryFilters(defaultState);
      setHasUnappliedChanges(false);
    },
    [getDefaultState, setQueryFilters],
  );

  const applyFilters = useCallback(
    (filterState: FilterState) => {
      setQueryFilters(filterState);
      setHasUnappliedChanges(false);
    },
    [setQueryFilters],
  );

  const applyRange = useCallback(
    (range: [number, number]) => {
      setQueryFilters({ range });
      setHasUnappliedChanges(false);
    },
    [setQueryFilters],
  );

  const updateHasUnappliedChanges = useCallback((hasChanges: boolean) => {
    setHasUnappliedChanges(hasChanges);
  }, []);

  const value = useMemo(
    () => ({
      data,
      filters: appliedFilters.filters,
      filtersInclude: appliedFilters.filtersInclude,
      range: appliedFilters.range,
      includeUndated: appliedFilters.includeUndated,
      textSearch: appliedFilters.textSearch,
      textSearchFields: appliedFilters.textSearchFields,
      minYear,
      maxYear,
      applyFilters,
      applyRange,
      resetFilters,
      updateHasUnappliedChanges,
      hasUnappliedChanges,
    }),
    [
      data,
      appliedFilters.filters,
      appliedFilters.filtersInclude,
      appliedFilters.range,
      appliedFilters.includeUndated,
      appliedFilters.textSearch,
      appliedFilters.textSearchFields,
      minYear,
      maxYear,
      applyFilters,
      applyRange,
      resetFilters,
      updateHasUnappliedChanges,
      hasUnappliedChanges,
    ],
  );

  return (
    <FilterAppliedContext.Provider value={value}>
      {children}
    </FilterAppliedContext.Provider>
  );
};
