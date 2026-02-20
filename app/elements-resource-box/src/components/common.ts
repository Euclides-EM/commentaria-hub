import styled from "@emotion/styled";
import { css } from "@emotion/react";

export const Container = styled.div`
  position: relative;
  top: 0;
  left: 0;
  width: 100%;
  margin: 2rem 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 2rem;

  .basic-multi-select {
    color: black;
    min-width: 12rem;
    max-width: 40rem;
    @media (max-width: 600px) {
      width: 80vw;
    }
  }
  .features-multi-select {
    width: 40rem;
  }
`;

export const Row = styled.div<{
  justifyStart?: boolean;
  gap?: number;
  rowGap?: number;
  noWrap?: boolean;
  noWrapAlsoOnMobile?: boolean;
  column?: boolean;
  width?: string;
}>`
  display: flex;
  flex-direction: ${({ column }) => (column ? "column" : "row")};
  justify-content: ${({ justifyStart }) => (justifyStart ? "start" : "center")};
  align-items: center;
  gap: ${({ gap }) => (gap !== undefined ? gap : 2)}rem;
  ${({ rowGap }) => rowGap && `row-gap: ${rowGap}`};
  width: ${({ width }) => (width ? width : "100%")};
  max-width: 96vw;
  flex-wrap: ${({ noWrap }) => (noWrap ? "nowrap" : "wrap")};
  @media (max-width: 600px) {
    flex-wrap: ${({ noWrapAlsoOnMobile }) =>
      noWrapAlsoOnMobile ? "nowrap" : "wrap"};
    gap: ${({ gap }) => (gap !== undefined ? gap : 1)}rem;
  }
`;

export const ResetButton = styled.button`
  background-color: aliceblue;
  color: #282828ff;
  border: unset;
  border-radius: 0.25rem;
  padding: 0.25rem 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;

  &:hover {
    background-color: #e3ecf5ff;
  }
`;

export const ToggleButton = styled(ResetButton)<{ isOpen: boolean }>`
  display: flex;
  align-content: center;
  justify-content: center;
  gap: 0.5rem;
  width: 100%;
`;

export const Column = styled.div<{ minWidth?: string; alignItems?: string }>`
  display: flex;
  flex-direction: column;
  justify-content: start;
  align-items: ${(props) => props.alignItems || "center"};
  gap: 1rem;
  min-width: ${({ minWidth }) => minWidth || "unset"};
  max-width: 90vw;
`;

export const Text = styled.div<{
  size: number;
  bold?: boolean;
  color?: string;
  clickable?: boolean;
}>`
  font-family: "Frank Ruhl Libre", serif;
  font-size: ${({ size }) => size}rem;
  font-weight: ${({ bold }) => (bold ? "bold" : "normal")};
  text-align: center;
  color: ${({ color }) => color || "unset"};
  ${({ clickable }) =>
    clickable &&
    css`
      cursor: pointer;
    `}
`;

export const Tile = css`
  border-radius: 1rem;
  clip-path: inset(0 round 1rem);
`;

export const ScrollbarStyle = css`
  ::-webkit-scrollbar-track {
    background-color: inherit;
  }

  ::-webkit-scrollbar {
    width: 0.5rem;
  }

  ::-webkit-scrollbar-thumb {
    background-color: #666;
    border-radius: 0.5rem;
  }

  * {
    scrollbar-width: thin;
    scrollbar-color: #666 inherit;
  }
`;

export const TextTile = styled.div<{ alignCenter: boolean }>`
  ${Tile};
  ${ScrollbarStyle};
  background-color: aliceblue;
  color: black;
  padding: 1rem;
  overflow: auto;
  height: 90%;
  width: 90%;
  line-height: 1.8;
  text-align: ${({ alignCenter }) => (alignCenter ? "center" : "start")};
  position: relative;
`;

export const ExpandIcon = styled.div`
  position: absolute;
  top: -0.2rem;
  right: 0.5rem;
  cursor: pointer;
  font-size: 1.2rem;
  z-index: 100;
  @media (max-width: 600px) {
    display: none;
  }
`;

export const ImageExpandIcon = styled.div`
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  cursor: pointer;
  font-size: 1.2rem;
  z-index: 10;
  background-color: aliceblue;
  color: black;
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 50%;
  display: flex;
  justify-content: center;
  align-items: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  @media (max-width: 600px) {
    display: none;
  }
`;

export const ScrollToTopButton = styled(ResetButton)`
  position: fixed;
  bottom: 1rem;
  right: 1rem;
  z-index: 100;
  width: 2rem;
  height: 2rem;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 1.2rem;
  border-radius: 50%;
  box-shadow: 0 2px 5px rgba(0, 0, 0, 0.2);
  transition:
    opacity 0.3s,
    transform 0.3s;
`;

export const NoImageTile = styled.div`
  ${Tile};
  height: 80%;
  width: 80%;
  background-color: rgba(240, 248, 255, 0.2);
  color: black;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
`;

export const ImageTile = styled.div`
  position: relative;
  max-height: 100%;
  max-width: 100%;
  display: flex;
  justify-content: center;
  align-items: start;
`;

export const StyledImage = styled.img<{ large?: boolean; clickable: boolean }>`
  ${Tile};
  width: min-content;
  max-height: ${({ large }) => (large ? "100%" : "90%")};
  max-width: ${({ large }) => (large ? "100%" : "90%")};
  cursor: ${({ clickable }) => (clickable ? "pointer" : "default")};
`;

export const LanguagesInfo = styled.div`
  font-size: 0.9rem;
  color: lightgray;
  text-align: center;
`;

export { LazyImage } from "./common/LazyImage";
