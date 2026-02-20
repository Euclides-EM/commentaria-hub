import {
  Cell,
  Legend,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
} from "recharts";
import { getStableColor, MARKER_3, SEA_COLOR } from "../../utils/colors";
import { GroupByOption } from "./useTrendsData";
import { PieChartContainer } from "./TrendsStyles";
import { Text } from "../../components/common";

type TrendsPieChartProps = {
  pieChartData: Array<{
    name: string;
    value: number;
    percent: number;
  }>;
  groupBy: GroupByOption;
};

export function TrendsPieChart({ pieChartData, groupBy }: TrendsPieChartProps) {
  if (!groupBy.key || pieChartData.length === 0) {
    return null;
  }

  return (
    <>
      <Text size={1.5} style={{ marginTop: "-4rem" }}>
        Distribution of {groupBy.label}
      </Text>
      <PieChartContainer>
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={pieChartData}
              dataKey="value"
              nameKey="name"
              cx="50%"
              cy="50%"
              outerRadius="70%"
              labelLine={true}
              label={({ name, percent }) =>
                `${name}: ${Number.isInteger(percent) ? percent : parseFloat(percent.toFixed(2))}%`
              }
            >
              {pieChartData.map((entry) => (
                <Cell key={entry.name} fill={getStableColor(entry.name)} />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{
                backgroundColor: SEA_COLOR,
                border: `1px solid ${MARKER_3}`,
                borderRadius: "8px",
                padding: "10px",
              }}
              itemStyle={{ color: "#ffffff" }}
              formatter={(value, name) => {
                const entry = pieChartData.find((item) => item.name === name);
                const color = getStableColor(name as string);
                return [
                  <div style={{ display: "flex", alignItems: "center" }}>
                    <span
                      style={{
                        display: "inline-block",
                        width: "12px",
                        height: "12px",
                        backgroundColor: color,
                        marginRight: "8px",
                        borderRadius: "2px",
                      }}
                    />
                    <span>
                      {name}:{" "}
                      {Number.isInteger(value)
                        ? value
                        : parseFloat(Number(value).toFixed(2))}{" "}
                      (
                      {entry?.percent
                        ? Number.isInteger(entry.percent)
                          ? entry.percent
                          : parseFloat(entry.percent.toFixed(2))
                        : 0}
                      %)
                    </span>
                  </div>,
                  <span style={{ display: "flex", alignItems: "center" }}>
                    <span
                      style={{
                        display: "inline-block",
                        width: "12px",
                        height: "12px",
                        backgroundColor: color,
                        marginRight: "8px",
                        borderRadius: "2px",
                      }}
                    />
                    {name}
                  </span>,
                ];
              }}
            />
            <Legend
              layout="vertical"
              verticalAlign="middle"
              align="right"
              formatter={(value) => {
                const entry = pieChartData.find((item) => item.name === value);
                return (
                  <span style={{ display: "flex", alignItems: "center" }}>
                    <span>
                      {value}:{" "}
                      {entry?.value
                        ? Number.isInteger(entry.value)
                          ? entry.value
                          : parseFloat(Number(entry.value).toFixed(2))
                        : 0}{" "}
                      (
                      {entry?.percent
                        ? Number.isInteger(entry.percent)
                          ? entry.percent
                          : parseFloat(entry.percent.toFixed(2))
                        : 0}
                      %)
                    </span>
                  </span>
                );
              }}
            />
          </PieChart>
        </ResponsiveContainer>
      </PieChartContainer>
    </>
  );
}
