import styled from "@emotion/styled";
import { LuFilter } from "react-icons/lu";
import { useEditFilter } from "../../contexts/FilterEditContext";
import { TOOLTIP_FILTERS } from "../map/MapTooltips";
import React, { useRef } from "react";
import { FILTER_TOGGLE_BUTTON_ID } from "../map/Tour.tsx";
import { useAppliedFilter } from "../../contexts/FilterAppliedContext.tsx";

const FilterButtonContainer = styled.div`
  font-size: 1.5rem;
  cursor: pointer;
  color: white;
  display: flex;
  align-items: center;
  background-color: black;
  padding: 0.5rem;
  border-radius: 50%;
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.6);
  width: fit-content;

  svg {
    transform: translateY(2px);
  }
`;

export const FilterButton = () => {
  const { filterOpen, setFilterOpen } = useEditFilter();
  const { data } = useAppliedFilter();
  const buttonRef = useRef<HTMLDivElement>(null);

  if (!data) {
    return null;
  }

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setFilterOpen((f) => !f);
  };

  return (
    <FilterButtonContainer
      ref={buttonRef}
      onClick={handleClick}
      data-tooltip-id={TOOLTIP_FILTERS}
      data-tooltip-content={filterOpen ? "Hide Filters" : "Show Filters"}
      id={FILTER_TOGGLE_BUTTON_ID}
    >
      <LuFilter />
    </FilterButtonContainer>
  );
};
