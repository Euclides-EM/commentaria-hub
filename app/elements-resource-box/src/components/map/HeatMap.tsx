import {
  MARKER_1,
  MARKER_2,
  MARKER_3,
  MARKER_4,
  MARKER_5,
  SEA_COLOR,
  TRANSPARENT_WHITE,
} from "../../utils/colors";
import styled from "@emotion/styled";

export const TOP_N = 5;

const COLORS_HEAT_MAP = [MARKER_1, MARKER_2, MARKER_3, MARKER_4, MARKER_5];

const RANGES = [40, 15, 5, 3, 0];

if (COLORS_HEAT_MAP.length !== TOP_N) {
  throw Error("hit map is outdated");
}

export const getHeatColor = (value: number): string => {
  for (let i = 0; i < TOP_N - 1; i++) {
    if (value >= RANGES[i]) {
      return COLORS_HEAT_MAP[i];
    }
  }
  return COLORS_HEAT_MAP[TOP_N - 1];
};

const Legend = styled.div<{ offsetRight: number }>`
  position: absolute;
  left: calc(${({ offsetRight }) => offsetRight}px - 12rem);
  bottom: 4.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  background-color: ${TRANSPARENT_WHITE};
  padding: 1rem;
  border-radius: 0.5rem;
`;

const Circle = styled.div<{ color: string }>`
  height: 1rem;
  width: 1rem;
  border-radius: 50%;
  cursor: pointer;
  background-color: ${({ color }) => color};
`;

const Row = styled.div`
  display: flex;
  flex-direction: row;
  gap: 0.5rem;
  align-items: center;
`;

const Title = styled.div`
  font-weight: bolder;
  padding-bottom: 0.5rem;
`;

const TotalLabel = styled.span`
  background-color: ${SEA_COLOR};
  color: white;
  width: fit-content;
  padding: 0.5rem;
  border-radius: 0.5rem;
`;

const HeatEntry = ({ index, label }: { index: number; label: string }) => (
  <Row>
    <Circle color={COLORS_HEAT_MAP[index]} />
    <div>{label}</div>
  </Row>
);

const getLabel = (index: number) => {
  if (index === 0) {
    return `${RANGES[0]} or more`;
  }
  if (index === TOP_N - 1) {
    return `${RANGES[TOP_N - 2] - 1} or less`;
  }
  return `${RANGES[index]}-${RANGES[index - 1] - 1}`;
};

export const HeatLegend = ({
  total,
  offsetRight,
}: {
  total: number;
  offsetRight: number;
}) => (
  <Legend offsetRight={offsetRight}>
    <div>
      Total: <TotalLabel>{total}</TotalLabel>
    </div>
    <div />
    <Title>Heatmap Legend:</Title>
    {RANGES.map((value, index) => (
      <HeatEntry key={value} index={index} label={getLabel(index)} />
    ))}
  </Legend>
);
