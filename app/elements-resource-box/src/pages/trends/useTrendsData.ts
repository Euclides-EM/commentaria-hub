import React, { useMemo } from "react";
import { useLocalStorage } from "usehooks-ts";
import { Item, Range } from "../../types";
import { itemProperties } from "../../constants/itemProperties.ts";
import { useAppliedFilter } from "../../contexts/FilterAppliedContext";
import { useEditionsSearch } from "../../hooks/useEditionsSearch";
import { isNil } from "lodash";

export type GroupByOption = {
  key: keyof Item | "";
  label: string;
  isArray?: boolean;
  origIsArray?: boolean;
  isTitlePageImageFeature?: boolean;
  isTitlePageTextFeature?: boolean;
};

export type TimeWindow = {
  start: number;
  end: number;
  count: number;
  groups: Record<string, number>;
};

function countItem(
  groupBy: GroupByOption,
  value: unknown,
  addValue: (key: string, count: number) => void,
) {
  if (groupBy.isArray && Array.isArray(value)) {
    if (value.length === 0) {
      addValue("None", 1);
    }
    value.forEach((val) => {
      let strVal: string;
      if (groupBy.key === "elementsBooks") {
        const range = val as Range;
        strVal =
          range.start === range.end
            ? range.start.toString()
            : `${range.start}-${range.end}`;
      } else {
        strVal = String(val);
      }
      addValue(strVal, 1 / value.length);
    });
  } else {
    let strVal: string;
    if (groupBy.origIsArray) {
      strVal =
        (value as string[] | null)?.length === 0
          ? ""
          : (value as string[] | null)?.join(" & ") || "";
    } else {
      strVal = isNil(value) ? "" : String(value);
    }
    if (strVal === "") {
      strVal = groupBy.isTitlePageImageFeature
        ? "No Digital Facsimile"
        : groupBy.isTitlePageTextFeature
          ? "No Title Page"
          : "Uncategorized";
    }
    addValue(strVal, 1);
  }
}

export function useTrendsData() {
  const { range } = useAppliedFilter();
  const { items: filteredItems } = useEditionsSearch();
  const [groupBy, setGroupBy] = useLocalStorage<GroupByOption>(
    "trends-group-by",
    {
      key: "",
      label: "None",
    },
  );
  const [windowSize, setWindowSize] = useLocalStorage<number>(
    "trends-window-size",
    10,
  );

  const groupByOptions = useMemo(() => {
    const options: GroupByOption[] = [{ key: "", label: "None" }];

    Object.entries(itemProperties)
      .sort(([, a], [, b]) => {
        return a.displayName.localeCompare(b.displayName);
      })
      .forEach(([key, config]) => {
        if (!config.notGroupable) {
          options.push({
            key: key as keyof Item,
            label: config.displayName || key,
            isArray: (config.isArray || false) && !config.groupByJoinArray,
            origIsArray: config.isArray || false,
            isTitlePageImageFeature: config.isTitlePageImageFeature,
            isTitlePageTextFeature: config.isTitlePageTextFeature,
          });
        }
      });

    return options;
  }, []);

  const handleGroupByChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const value = e.target.value;
    const option = groupByOptions.find((opt) => opt.key === value) || {
      key: "",
      label: "None",
    };
    setGroupBy(option);
  };

  const timeWindows = useMemo(() => {
    if (!filteredItems || !filteredItems.length) return [];

    const [minYear, maxYear] = range;
    const windowsCount = Math.ceil((maxYear - minYear) / windowSize);

    const windows: TimeWindow[] = Array.from(
      { length: windowsCount },
      (_, i) => {
        const start = minYear + i * windowSize;
        const end = Math.min(start + windowSize - 1, maxYear);
        return {
          start,
          end,
          count: 0,
          groups: {},
        };
      },
    );

    filteredItems?.forEach((item) => {
      if (!item.year) return;

      const year = parseInt(item.year.split("/")[0]);
      const windowIndex = Math.floor((year - minYear) / windowSize);

      if (windowIndex >= 0 && windowIndex < windows.length) {
        windows[windowIndex].count += 1;

        if (groupBy.key) {
          const value = item[groupBy.key];
          countItem(groupBy, value, (key, count) => {
            windows[windowIndex].groups[key] =
              (windows[windowIndex].groups[key] || 0) + count;
          });
        }
      }
    });

    return windows;
  }, [filteredItems, range, windowSize, groupBy]);

  const chartData = useMemo(() => {
    if (!timeWindows.length) return [];

    return timeWindows.map((window) => {
      const data: Record<string, unknown> = {
        name: `${window.start}-${window.end}`,
        total: window.count,
      };

      if (groupBy.key) {
        Object.entries(window.groups).forEach(([group, count]) => {
          data[group] = count;
        });
      }

      return data;
    });
  }, [timeWindows, groupBy]);

  const uniqueGroups = useMemo(() => {
    if (!groupBy.key) return [];

    const groups = new Set<string>();
    timeWindows.forEach((window) => {
      Object.keys(window.groups).forEach((group) => groups.add(group));
    });

    return Array.from(groups);
  }, [timeWindows, groupBy]);

  const pieChartData = useMemo(() => {
    if (!groupBy.key || !filteredItems || !filteredItems.length) return [];

    const groupCounts: Record<string, number> = {};
    let totalCount = 0;

    filteredItems?.forEach((item) => {
      const value = item[groupBy.key as keyof Item];

      countItem(groupBy, value, (key, count) => {
        groupCounts[key] = (groupCounts[key] || 0) + count;
        totalCount += count;
      });
    });

    return Object.entries(groupCounts)
      .map(([name, value]) => ({
        name,
        value,
        percent: Math.round((value / totalCount) * 100),
      }))
      .sort((a, b) => b.value - a.value);
  }, [filteredItems, groupBy]);

  return {
    filteredItems,
    range,
    groupBy,
    setGroupBy,
    windowSize,
    setWindowSize,
    groupByOptions,
    handleGroupByChange,
    chartData,
    uniqueGroups,
    pieChartData,
  };
}
