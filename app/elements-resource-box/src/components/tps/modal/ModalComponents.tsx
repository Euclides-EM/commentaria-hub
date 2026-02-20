import styled from "@emotion/styled";
import { Row, ScrollbarStyle } from "../../common.ts";
import { css } from "@emotion/react";

export const Modal = styled.div`
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
`;

export const ModalContent = styled.div<{ hasImage: boolean }>`
  ${ScrollbarStyle};
  background-color: aliceblue;
  min-height: 24rem;
  max-height: 80vh;
  min-width: 90vw;
  max-width: ${({ hasImage }) => (hasImage ? "90vw" : "60vw")};
  color: black;
  border-radius: 1rem;
  padding: 2rem;
  overflow: auto;
  display: flex;
  flex-direction: column;
  position: relative;
  justify-content: center;
  align-content: center;
`;

export const ModalClose = styled.div`
  position: absolute;
  top: 0.5rem;
  right: 0.8rem;
  font-size: 1rem;
  cursor: pointer;
`;

export const ModalTitle = styled(Row)`
  color: darkgray;
  font-size: 1rem;
  margin-bottom: 0.5rem;
`;

export const ModalTitleRow = styled(Row)`
  font-size: 1.2rem;
  gap: 1rem;
  margin-bottom: 0.5rem;
  justify-content: start;
`;

export const ModalTextContainer = styled.div`
  ${ScrollbarStyle};
  display: flex;
  flex-direction: row;
  gap: 2rem;
  overflow: auto;
  max-height: calc(90vh - 6rem);
`;

export const TextColumnsContainer = styled.div`
  ${ScrollbarStyle};
  display: flex;
  flex-direction: row;
  gap: 2rem;
  overflow-y: auto;
  flex: 2;
  overflow-x: visible;
`;

export const ModalTextColumn = styled.div<{
  isImage?: boolean;
  isTextContent?: boolean;
  isRow?: boolean;
  alignCenter?: boolean;
}>`
  flex: 1;
  overflow-y: ${({ isImage }) => (isImage ? "hidden" : "visible")};
  line-height: 1.8;
  padding: 0 0.5rem;
  min-width: 250px;
  max-width: 90vw;
  width: auto;
  text-align: ${({ alignCenter }) => (alignCenter ? "center" : "start")};
  ${({ isRow }) =>
    isRow &&
    css`
      max-width: 90%;
      display: flex;
      flex-direction: row;
      gap: 1rem;
      margin-bottom: 1rem;
      div {
        display: block;
      }
    `}
`;
