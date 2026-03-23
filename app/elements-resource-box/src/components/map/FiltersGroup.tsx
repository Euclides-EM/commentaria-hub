import React, { useEffect, useMemo, useState } from "react";
import { groupBy, isNil, startCase, uniq } from "lodash";
import { Filter, FilterValue } from "./Filter";
import { Item, STUDY_CORPUSES } from "../../types";
import { personDisplayName } from "../../utils/dataUtils.ts";
import { ItemProperty } from "../../constants/itemProperties.ts";
import styled from "@emotion/styled";

const GroupSeparator = styled.div`
  border-top: 1px solid #e0e0e0;
  margin: 0.5rem 0;
`;

const GroupHeader = styled.div`
  cursor: pointer;
  padding: 8px 0;
  font-weight: bold;
  font-size: 14px;
  color: #666;
  user-select: none;
  display: flex;
  align-items: center;
  gap: 8px;
`;

const GroupArrow = styled.span<{ $collapsed: boolean }>`
  transform: ${(props) =>
    props.$collapsed ? "rotate(-90deg)" : "rotate(0deg)"};
  transition: transform 0.2s ease;
  font-size: 12px;
`;

const GroupContent = styled.div`
  display: flex;
  flex-direction: column;
  gap: 1rem;
`;

type FiltersGroupProps = {
  data: Item[];
  fields: Partial<Record<keyof Item, ItemProperty>>;
  filters: Record<string, FilterValue[] | undefined>;
  setFilters: React.Dispatch<
    React.SetStateAction<Record<string, FilterValue[] | undefined>>
  >;
  filtersInclude: Record<string, boolean>;
  setFiltersInclude: React.Dispatch<
    React.SetStateAction<Record<string, boolean>>
  >;
};

const mapStudyCorpus = (s: string): string => {
  return STUDY_CORPUSES[s] || startCase(s.toLowerCase());
};

const toFormat = (value: string | undefined) => {
  if (isNil(value)) {
    return "";
  }
  return `${value}º`;
};

const optionDisplayName = (
  field: keyof Item,
  value: string | undefined | null,
): string => {
  if (field === "editors" || field === "publishers") {
    return personDisplayName(value || "");
  }
  if (field === "study_corpora") {
    return mapStudyCorpus(value || "");
  }
  if (field === "diagramCropsAvailable") {
    return value === "true" ? "Yes" : "No";
  }
  if (field === "hasDiagrams") {
    return isNil(value) ? "Uncatalogued" : value === "true" ? "Yes" : "No";
  }
  if (field === "format") {
    return toFormat(value || "");
  }

  return value?.toString().replace("(?)", "").replace("?", "").trim() || "";
};

const toOption = (field: keyof Item, v: string | undefined | null) => ({
  label: optionDisplayName(field, v),
  value: v || "false",
});

export const FiltersGroup = ({
  data,
  fields,
  filters,
  setFilters,
  filtersInclude,
  setFiltersInclude,
}: FiltersGroupProps) => {
  const keys = Object.keys(fields)
    .filter((key) => !fields[key as keyof Item]?.notFilterable)
    .map((field) => field as keyof Item);

  const groupOrder = useMemo(
    () => ["Common", "Elements", "Title Page", "Material", "Diagrams"],
    [],
  );
  const [collapsedGroups, setCollapsedGroups] = useState<
    Record<string, boolean>
  >({});

  useEffect(() => {
    const stored = localStorage.getItem("filter-groups-collapsed");
    if (stored) {
      setCollapsedGroups(JSON.parse(stored));
    } else {
      const defaultCollapsed = groupOrder.reduce(
        (acc, group) => ({ ...acc, [group]: true }),
        {} as Record<string, boolean>,
      );
      setCollapsedGroups(defaultCollapsed);
    }
  }, [groupOrder]);

  const toggleGroup = (groupName: string) => {
    setCollapsedGroups((prev) => {
      const newState = { ...prev, [groupName]: !prev[groupName] };
      localStorage.setItem("filter-groups-collapsed", JSON.stringify(newState));
      return newState;
    });
  };

  const groupedFields = useMemo(() => {
    const fieldEntries = keys.map((field) => ({
      field,
      config: fields[field]!,
      group: fields[field]?.filterGroup || "Common",
    }));

    return groupBy(fieldEntries, "group");
  }, [keys, fields]);

  const optionsByFilter = useMemo(() => {
    const byFilter: Record<string, FilterValue[]> = {};
    keys.forEach((field) => {
      const config = fields[field]!;
      if (config.isArray) {
        byFilter[field] = uniq(
          data
            .flatMap((t) => t[field] as (string | number)[])
            .filter((v) => v !== "" && v !== -1)
            .sort(
              config.customCompareFn ||
                ((a, b) => {
                  if (field === "editors" || field === "publishers") {
                    return personDisplayName(a as string).localeCompare(
                      personDisplayName(b as string),
                    );
                  }
                  if (typeof a === "string" && typeof b === "string") {
                    return a.localeCompare(b);
                  }
                  if (typeof a === "number" && typeof b === "number") {
                    return a - b;
                  }
                  return 0;
                }),
            ),
        ).map((v) => toOption(field, v as string));
      } else {
        byFilter[field] = uniq(
          data
            .map((t) => t[field]?.toString())
            .filter((v) => v !== "" && v !== "-1")
            .sort(config.customCompareFn),
        ).map((v) => toOption(field, v));
      }
    });
    return byFilter;
  }, [data, fields, keys]);

  const sortedGroups = Object.keys(groupedFields).sort((a, b) => {
    const aIndex = groupOrder.indexOf(a);
    const bIndex = groupOrder.indexOf(b);
    if (aIndex === -1 && bIndex === -1) return a.localeCompare(b);
    if (aIndex === -1) return 1;
    if (bIndex === -1) return -1;
    return aIndex - bIndex;
  });

  return (
    <div>
      {sortedGroups.map((groupName, groupIndex) => (
        <div key={groupName}>
          {groupIndex > 0 && <GroupSeparator />}
          <GroupHeader onClick={() => toggleGroup(groupName)}>
            <GroupArrow $collapsed={collapsedGroups[groupName]}>▼</GroupArrow>
            {groupName}
          </GroupHeader>
          {!collapsedGroups[groupName] && (
            <GroupContent>
              {groupedFields[groupName].map(({ field, config }) => (
                <Filter
                  field={field}
                  key={field}
                  label={config.displayName || startCase(field)}
                  value={filters[field]}
                  setValue={(values) =>
                    setFilters((f) => ({
                      ...f,
                      [field]: values ? [...values] : undefined,
                    }))
                  }
                  options={optionsByFilter[field]}
                  include={
                    isNil(filtersInclude[field]) ? true : filtersInclude[field]
                  }
                  setInclude={(include) =>
                    setFiltersInclude((f) => ({ ...f, [field]: include }))
                  }
                />
              ))}
            </GroupContent>
          )}
        </div>
      ))}
    </div>
  );
};
