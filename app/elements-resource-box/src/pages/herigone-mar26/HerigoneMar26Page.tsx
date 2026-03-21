import styled from "@emotion/styled";
import { PANE_COLOR, SEA_COLOR } from "../../utils/colors";
import { HerigoneMar26Proposition } from "./HerigoneMar26Proposition.tsx";

const Wrapper = styled.div`
  height: 100dvh;
  width: 100vw;
  overflow: hidden;
`;

const Container = styled.div`
  display: flex;
  height: 100dvh;
  flex-direction: column;
  align-items: center;
  justify-content: start;
  color: white;
  padding: 2rem 1rem calc(1rem + env(safe-area-inset-bottom, 0px));
  background-color: ${SEA_COLOR};
  gap: clamp(0.15rem, 0.45vh, 0.4rem);
  text-align: center;
  box-sizing: border-box;
  overflow-y: auto;

  @media only screen and (max-width: 500px) {
    padding: 1rem 0.75rem calc(0.9rem + env(safe-area-inset-bottom, 0px));
    gap: clamp(0.1rem, 0.35vh, 0.3rem);
  }
`;

const Spacer = styled.div<{ size?: "sm" | "md" }>`
  height: ${({ size }) =>
    size === "sm"
      ? "clamp(0.05rem, 0.18vh, 0.15rem)"
      : "clamp(0.1rem, 0.35vh, 0.25rem)"};
  flex: 0 0 auto;
`;

const Title = styled.div`
  font-size: clamp(1.45rem, 2.8vw, 2rem);
  font-weight: bold;
`;

const SubTitle = styled.div`
  font-size: clamp(1rem, 1.8vw, 1.15rem);
  color: ${PANE_COLOR};
  max-width: 90vw;
`;

const Stage = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
`;

const HeadingBlock = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: clamp(0.15rem, 0.45vh, 0.4rem);
`;

const Credit = styled.div``;

const CardSlot = styled.div`
  width: 100%;
  display: flex;
  justify-content: center;
`;

export function HerigoneMar26Page() {
  return (
    <Wrapper>
      <Container data-herigone-scroll-container="true">
        <Stage>
          <HeadingBlock>
            <Title>Hérigone, Book II.4</Title>
            <SubTitle>
              If a straight line is cut in whatever way one wishes: the square
              of the whole is equal to the squares of the parts, and to twice
              the rectangle contained under those same parts.
            </SubTitle>
            <Credit>Mia Joskowicz, March 2026</Credit>
            <Spacer />
            <Spacer size="sm" />
          </HeadingBlock>
          <CardSlot>
            <HerigoneMar26Proposition />
          </CardSlot>
        </Stage>
      </Container>
    </Wrapper>
  );
}
