import { Proposition } from "./Proposition";

const propositionVer1PathPrefix = "/b2p5/ver1/p";
const propositionVer1Text = [
  "If a straight line is cut into equal and unequal segments,",
  "then the rectangle contained by the unequal segments of the whole,",
  "together with the square on the straight line between the points of section,",
  "equals the square on the half.",
];

const propositionVer1Mapping: Record<number, number> = {
  1: 0,
  2: 0,
  3: 0,
  4: 1,
  5: 1,
  6: 1,
  7: 1,
  8: 1,
  9: 1,
  10: 1,
  11: 1,
  13: 2,
  14: 2,
  15: 2,
  16: 2,
  17: 2,
  18: 2,
  19: 2,
  20: 2,
  21: 2,
  22: 2,
  23: 2,
  24: 3,
  25: 3,
  26: 3,
  27: 3,
  28: 3,
  29: 3,
  30: 3,
  31: 3,
  32: 3,
  33: 3,
  34: 3,
  35: 3,
  36: 3,
  37: 4,
};

export function Proposition5Book2V1() {
  return (
    <Proposition
      title="Proposition 5, Book II"
      description={propositionVer1Text}
      stepImagePrefix={propositionVer1PathPrefix}
      stepsToDescriptionIndex={propositionVer1Mapping}
    />
  );
}
