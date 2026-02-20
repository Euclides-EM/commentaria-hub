import styled from "@emotion/styled";
import { LAND_COLOR, SEA_COLOR } from "../../utils/colors.ts";

export const Switch = styled.div`
  display: flex;
  flex-direction: row;
  font-size: 0.8rem;
  height: fit-content;
  align-self: end;

  div:first-child {
    border-radius: 5px 0 0 5px;
  }

  div:last-child {
    border-radius: 0 5px 5px 0;
  }
`;

export const SwitchOption = styled.div<{ selected: boolean }>`
  cursor: pointer;
  padding: 6px 10px;
  color: white;

  background-color: ${({ selected }) => (selected ? SEA_COLOR : LAND_COLOR)};

  font-weight: ${({ selected }) => (selected ? 600 : 400)};
  opacity: ${({ selected }) => (selected ? 1 : 0.65)};

  transition:
    background-color 0.2s ease,
    opacity 0.2s ease,
    box-shadow 0.2s ease;

  &:hover {
    opacity: 1;
  }
`;
