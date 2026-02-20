import React, { useMemo, useState, useEffect } from "react";
import { isNil, startCase, uniq, groupBy } from "lodash";
import { Filter, FilterValue } from "./Filter";
import { Item } from "../../types";
import { authorDisplayName } from "../../utils/dataUtils.ts";
import { ItemProperty } from "../../constants/itemProperties.ts";
import styled from "@emotion/styled";
import { NO_CITY } from "../../constants";

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

const toOption = (v: string) => ({ label: v, value: v });

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
            .filter(Boolean)
            .sort(
              config.customCompareFn ||
                ((a, b) => {
                  if (field === "authors") {
                    return authorDisplayName(a as string).localeCompare(
                      authorDisplayName(b as string),
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
            )
            .map((n) =>
              n.toString().replace("(?)", "").replace("?", "").trim(),
            ),
        ).map(
          field === "authors"
            ? (n) => ({
                label: authorDisplayName(n),
                value: n,
              })
            : toOption,
        );
      } else {
        byFilter[field] = uniq(
          data
            .map((t) => t[field]?.toString() || "")
            .filter((v) => v && v !== NO_CITY)
            .sort(config.customCompareFn),
        ).map(toOption);
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
