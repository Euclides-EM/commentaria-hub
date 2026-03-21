import styled from "@emotion/styled";
import { PANE_COLOR, SEA_COLOR } from "../../utils/colors";
import { HerigoneMar25Proposition } from "./HerigoneMar25Proposition";

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
  max-width: 44rem;
`;

export function HerigoneMar25Page() {
  return (
    <Wrapper>
      <Container>
        <Title>Hérigone, Book II.4</Title>
        <SubTitle>
          A step-by-step reconstruction of the geometric proof from 1634
        </SubTitle>
        <div>Mia Joskowicz, March 2025</div>
        <Spacer />
        <Spacer size="sm" />
        <HerigoneMar25Proposition />
      </Container>
    </Wrapper>
  );
}
