import React, { useCallback } from "react";
import {
  MdMyLocation,
  MdQuestionMark,
  MdZoomInMap,
  MdZoomOutMap,
} from "react-icons/md";
import styled from "@emotion/styled";
import {
  TOOLTIP_HELP,
  TOOLTIP_RESET,
  TOOLTIP_ZOOMIN,
  TOOLTIP_ZOOMOUT,
} from "./MapTooltips.tsx";
import { ZOOM_CONTROLS_ID } from "./Tour";
import { ButtonStyle } from "../../utils/colors.ts";
import { MIN_ZOOM } from "./CitiesMap";

type ZoomControlsProps = {
  className?: string;
  setZoom: React.Dispatch<React.SetStateAction<number>>;
  maxZoom: number;
  resetCenter: () => void;
  openTour: () => void;
};

const StyledDiv = styled.div`
  background-color: transparent;
  display: flex;
  flex-direction: row;
  gap: 1rem;

  div {
    ${ButtonStyle};
  }
  svg {
    margin-bottom: 2px;
  }
  @media (max-width: 768px) {
    gap: 6px;
  }
`;

export const MapControls = ({
  className,
  setZoom,
  maxZoom,
  resetCenter,
  openTour,
}: ZoomControlsProps) => {
  const handleZoomIn = useCallback(() => {
    setZoom((z) => Math.min(z < 1 ? z + 0.1 : z + 1, maxZoom));
  }, [maxZoom, setZoom]);

  const handleZoomOut = useCallback(() => {
    setZoom((z) => Math.max(z > 1 ? z - 1 : z - 0.1, MIN_ZOOM));
  }, [setZoom]);

  return (
    <StyledDiv id={ZOOM_CONTROLS_ID} className={className}>
      <div
        onClick={openTour}
        data-tooltip-id={TOOLTIP_HELP}
        data-tooltip-content="Help"
      >
        <MdQuestionMark />
      </div>
      <div
        onClick={resetCenter}
        data-tooltip-id={TOOLTIP_RESET}
        data-tooltip-content="Reset"
      >
        <MdMyLocation />
      </div>
      <div
        onClick={handleZoomIn}
        data-tooltip-id={TOOLTIP_ZOOMIN}
        data-tooltip-content="Zoom-In"
      >
        <MdZoomInMap />
      </div>
      <div
        onClick={handleZoomOut}
        data-tooltip-id={TOOLTIP_ZOOMOUT}
        data-tooltip-content="Zoom-Out"
      >
        <MdZoomOutMap />
      </div>
    </StyledDiv>
  );
};
