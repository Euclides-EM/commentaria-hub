import styled from "@emotion/styled";
import { PANE_COLOR, SEA_COLOR } from "../../utils/colors";
import { HerigoneMar26Proposition } from "./HerigoneMar26Proposition.tsx";

const Wrapper = styled.div`
  height: 100vh;
  height: 100dvh;
  width: 100vw;
  overflow: hidden;
`;

const Container = styled.div`
  display: flex;
  height: 100vh;
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

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    padding: 0;
    overflow: hidden;
    justify-content: center;
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

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    height: 100%;
    box-sizing: border-box;
    padding: 0.5rem 0.6rem;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 0.55rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    width: 100dvh;
    height: 100vw;
    box-sizing: border-box;
    padding: 0.5rem 0.6rem;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 0.55rem;
    transform: rotate(90deg);
    transform-origin: center;
  }
`;

const HeadingBlock = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: clamp(0.15rem, 0.45vh, 0.4rem);

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    align-items: flex-start;
    text-align: left;
    width: clamp(8.25rem, 19vw, 10.75rem);
    flex: 0 0 auto;
    gap: 0.15rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    margin-left: 0.15rem;
    align-items: flex-start;
    text-align: left;
    width: clamp(8.25rem, 22dvh, 10.75rem);
    flex: 0 0 auto;
    gap: 0.15rem;
  }
`;

const Credit = styled.div`
  @media only screen and (max-height: 500px) and (orientation: landscape) {
    margin-top: 1rem;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    margin-top: 1rem;
  }
`;

const CardSlot = styled.div`
  width: 100%;
  display: flex;
  justify-content: center;

  @media only screen and (max-height: 500px) and (orientation: landscape) {
    min-width: 0;
    flex: 1 1 auto;
    align-items: center;
  }

  @media only screen and (max-width: 500px) and (orientation: portrait) {
    min-width: 0;
    flex: 1 1 auto;
    align-items: center;
  }
`;

export function HerigoneMar26Page() {
  return (
    <Wrapper>
      <Container>
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
