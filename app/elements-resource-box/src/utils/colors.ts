import { css } from "@emotion/react";

export const SEA_COLOR = "#465966";
export const LAND_COLOR = "#c7b1ae";

export const MARKER_1 = "#4a076a";
export const MARKER_2 = "#af097e";
export const MARKER_3 = "#de3797";
export const MARKER_4 = "#fb9fb5";
export const MARKER_5 = "#fdc5bf";
export const MARKER_STROKE = "rgba(253,237,178,0.65)";
export const MARKER_STROKE_HL = "rgba(0, 0, 0, 0.65)";

export const TRANSPARENT_WHITE = css`rgba(255, 255, 255, 0.65)`;
export const TRANSPARENT_BLACK = css`rgba(0, 0, 0, 0.6)`;

export const PANE_COLOR = "#d5c99e";
export const PANE_COLOR_ALT = "#ACA280FF";
export const PANE_BORDER = "#615730FF";

export const RANGE_FILL = "#282722";

export const ButtonStyle = css`
  height: 1.5rem;
  width: 1.5rem;
  font-size: 1.5rem;
  background-color: ${TRANSPARENT_WHITE};
  color: ${SEA_COLOR};
  border-radius: 20%;
  padding: 0.4rem;
  cursor: pointer;
`;

export const getStableColor = (input: string): string => {
  if (input === "Yes") {
    input = "true";
  }
  if (
    ["No Digital Facsimile", "Uncategorized", "None", "No Title Page"].includes(
      input,
    )
  ) {
    return "#9f9f9f";
  }
  if (input === "Other") {
    return "#757575";
  }
  if (Number.isInteger(parseInt(input))) {
    input = `${input}-${input}`;
  }

  let hash = 0;
  const multiplier = 53;

  for (let i = 0; i < input.length; i++) {
    hash = input.charCodeAt(i) + hash * multiplier;
    hash = (hash << 5) + hash + input.charCodeAt(i);
    hash = (hash << 5) + hash + input.charCodeAt(i);
    hash = (hash << 5) + hash + input.charCodeAt(i);
  }

  const h = Math.abs(hash + 200) % 360;

  const s = 60 + (Math.abs(hash >> 8) % 20);

  const l = 55 + (Math.abs(hash >> 16) % 15);

  return `hsl(${h}, ${s}%, ${l}%)`;
};
