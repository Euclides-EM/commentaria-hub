import { memo } from "react";
import styled from "@emotion/styled";
import { createPortal } from "react-dom";
import { HighlightTooltipState } from "./highlightTooltipUtils";

type FeatureHighlightTooltipProps = {
  tooltipState: HighlightTooltipState | null;
  onRemove?: (highlight: HighlightTooltipState) => void;
  onTooltipEnter?: () => void;
  onTooltipLeave?: () => void;
};

const TooltipContainer = styled("div", {
  shouldForwardProp: (prop) => prop !== "x" && prop !== "y",
})<{ x: number; y: number }>`
  position: fixed;
  left: ${({ x }) => `${x}px`};
  top: ${({ y }) => `${y}px`};
  z-index: 12000;
  pointer-events: auto;
  background-color: white;
  color: black;
  padding: 1rem;
  border-radius: 0.35rem;
  font-size: 0.8rem;
  line-height: 1.3;
  max-width: 22rem;
  border: 1px solid #d9d9d9;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
`;

const FeatureLabel = styled("span", {
  shouldForwardProp: (prop) => prop !== "featureColor",
})<{ featureColor: string }>`
  display: inline-block;
  background-color: ${({ featureColor }) => featureColor};
  color: black;
  padding: 0.15rem 0.4rem;
  border-radius: 0.25rem;
  font-weight: 600;
`;

const NormalizedText = styled.div`
  margin-top: 0.35rem;
`;

const TooltipActions = styled.div`
  margin-top: 0.6rem;
  display: flex;
  justify-content: flex-end;
`;

const TooltipButton = styled.button`
  border: 1px solid #cfcfcf;
  background: white;
  color: #333;
  border-radius: 0.3rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  cursor: pointer;
`;

const FeatureHighlightTooltip = memo(
  ({
    tooltipState,
    onRemove,
    onTooltipEnter,
    onTooltipLeave,
  }: FeatureHighlightTooltipProps) => {
    if (!tooltipState) {
      return null;
    }

    const tooltip = (
      <TooltipContainer
        x={tooltipState.x}
        y={tooltipState.y}
        data-highlight-action
        onMouseEnter={onTooltipEnter}
        onMouseLeave={onTooltipLeave}
      >
        <FeatureLabel featureColor={tooltipState.color}>
          {tooltipState.label}
        </FeatureLabel>
        {tooltipState.normalized ? (
          <NormalizedText>
            {`Normalized: ${tooltipState.normalized}`}
          </NormalizedText>
        ) : null}
        {onRemove && (
          <TooltipActions>
            <TooltipButton type="button" onClick={() => onRemove(tooltipState)}>
              Remove
            </TooltipButton>
          </TooltipActions>
        )}
      </TooltipContainer>
    );

    if (typeof document === "undefined") {
      return tooltip;
    }

    return createPortal(tooltip, document.body);
  },
);

export default FeatureHighlightTooltip;
