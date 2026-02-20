import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  getStableColor,
  LAND_COLOR,
  MARKER_3,
  MARKER_4,
  MARKER_5,
  SEA_COLOR,
} from "../../utils/colors";
import { GroupByOption } from "./useTrendsData";
import { ChartContainer } from "./TrendsStyles";
import { itemProperties } from "../../constants/itemProperties.ts";

type TrendsBarChartProps = {
  chartData: Record<string, unknown>[];
  uniqueGroups: string[];
  groupBy: GroupByOption;
};

const lastValuesInSort = [
  "other",
  "uncategorized",
  "no digital facsimile",
  "none",
  "no title page",
];

export function TrendsBarChart({
  chartData,
  uniqueGroups,
  groupBy,
}: TrendsBarChartProps) {
  return (
    <ChartContainer>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={chartData}
          margin={{ top: 20, right: 30, left: 20, bottom: 70 }}
        >
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" angle={-45} textAnchor="end" height={70} />
          <YAxis />
          <Tooltip
            contentStyle={{
              backgroundColor: SEA_COLOR,
              border: `1px solid ${MARKER_3}`,
              borderRadius: "8px",
            }}
            itemStyle={{ color: "#ffffff" }}
            labelStyle={{
              color: MARKER_5,
              fontWeight: "bold",
              marginBottom: "10px",
              borderBottom: `1px solid ${MARKER_4}`,
              paddingBottom: "5px",
            }}
            cursor={{ fill: "rgba(74, 7, 106, 0.6)" }}
            itemSorter={() => -1}
            content={({ active, payload, label }) => {
              if (active && payload && payload.length) {
                let sortedPayload = [...payload];

                if (groupBy.key) {
                  const fieldConfig = itemProperties[groupBy.key as string];
                  if (fieldConfig && fieldConfig.customCompareFn) {
                    sortedPayload = sortedPayload.sort((a, b) =>
                      fieldConfig.customCompareFn!(a.name, b.name),
                    );
                  } else {
                    sortedPayload = sortedPayload.sort((a, b) => {
                      if (
                        lastValuesInSort.includes(
                          (a.name as string).toLowerCase(),
                        )
                      ) {
                        return 1;
                      }
                      if (
                        lastValuesInSort.includes(
                          (b.name as string).toLowerCase(),
                        )
                      ) {
                        return -1;
                      }
                      return (a.name as string).localeCompare(b.name as string);
                    });
                  }
                } else {
                  sortedPayload = sortedPayload.sort(
                    (a, b) => (b.value as number) - (a.value as number),
                  );
                }

                return (
                  <div
                    style={{
                      backgroundColor: SEA_COLOR,
                      padding: "10px",
                      border: `1px solid ${MARKER_3}`,
                      borderRadius: "8px",
                    }}
                  >
                    <p
                      style={{
                        color: MARKER_5,
                        fontWeight: "bold",
                        marginBottom: "10px",
                        borderBottom: `1px solid ${MARKER_4}`,
                        paddingBottom: "5px",
                      }}
                    >
                      {label}
                    </p>
                    {sortedPayload.map((entry, index) => {
                      const color =
                        entry.name === "Total Editions"
                          ? LAND_COLOR
                          : getStableColor(entry.name as string);

                      return (
                        <div
                          key={`item-${index}`}
                          style={{
                            display: "flex",
                            alignItems: "center",
                            color: "#ffffff",
                            margin: "4px 0",
                          }}
                        >
                          <div
                            style={{
                              width: "10px",
                              height: "10px",
                              backgroundColor: color,
                              marginRight: "8px",
                              borderRadius: "2px",
                            }}
                          />
                          <span style={{ marginRight: "8px" }}>
                            {entry.name}:
                          </span>
                          <span>
                            {Number.isInteger(entry.value)
                              ? entry.value
                              : parseFloat(Number(entry.value).toFixed(2))}
                          </span>
                        </div>
                      );
                    })}
                  </div>
                );
              }
              return null;
            }}
          />
          <Legend />
          {groupBy.key ? (
            [...uniqueGroups]
              .sort((a, b) => {
                const fieldConfig = itemProperties[groupBy.key as string];
                if (fieldConfig && fieldConfig.customCompareFn) {
                  return fieldConfig.customCompareFn(a, b);
                }
                if (lastValuesInSort.includes(a.toLowerCase())) {
                  return 1;
                }
                if (lastValuesInSort.includes(b.toLowerCase())) {
                  return -1;
                }
                return a.localeCompare(b);
              })
              .map((group) => (
                <Bar
                  key={group}
                  dataKey={group}
                  stackId="a"
                  fill={getStableColor(group)}
                  name={group}
                />
              ))
          ) : (
            <Bar
              dataKey="total"
              fill={LAND_COLOR}
              name="Total Editions"
              activeBar={{ fill: MARKER_3 }}
            />
          )}
        </BarChart>
      </ResponsiveContainer>
    </ChartContainer>
  );
}
