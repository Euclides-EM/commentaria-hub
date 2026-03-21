import styled from "@emotion/styled";
import { css } from "@emotion/react";
import { Link } from "react-router-dom";
import { Proposition5Book2V1 } from "./Proposition5Book2V1";
import { MARKER_4, PANE_COLOR, SEA_COLOR } from "../../utils/colors";
import { MAP_ROUTE } from "../../components/layout/routes";
import { withAppBasePath } from "../../utils/basePath";

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
  padding: 7.25rem 1rem calc(2.25rem + env(safe-area-inset-bottom, 0px));
  background-color: ${SEA_COLOR};
  gap: clamp(0.3rem, 0.8vh, 0.7rem);
  text-align: center;
  box-sizing: border-box;

  @media only screen and (max-width: 500px) {
    padding: 4.75rem 0.9rem calc(1.5rem + env(safe-area-inset-bottom, 0px));
  }
`;

const Spacer = styled.div<{ size?: "sm" | "md" }>`
  height: ${({ size }) =>
    size === "sm"
      ? "clamp(0.1rem, 0.35vh, 0.25rem)"
      : "clamp(0.2rem, 0.6vh, 0.45rem)"};
  flex: 0 0 auto;
`;

const Title = styled.div`
  font-size: 2rem;
  font-weight: bold;

  @media only screen and (max-width: 500px) {
    font-size: 1.5rem;
  }
`;

const SubTitle = styled.div`
  font-size: 1.2rem;
  color: ${PANE_COLOR};

  @media only screen and (max-width: 500px) {
    font-size: 1.1rem;
  }
`;

const linkStyle = css`
  font-size: 1.5rem;
  color: ${MARKER_4};

  @media only screen and (max-width: 500px) {
    font-size: 1.2rem;
  }
`;

const StyledAnchor = styled.a`
  ${linkStyle};
`;

const StyledLink = styled(Link)`
  ${linkStyle};
`;

const Block = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: start;
  gap: clamp(0.25rem, 0.6vh, 0.5rem);
`;

export function Sep24Page() {
  const pdfPath = withAppBasePath(
    "/presentation/TransformationOfMathematicalKnowledge.pdf",
  );

  return (
    <Wrapper>
      <Container>
        <Title>Transformation of Mathematical Knowledge</Title>
        <SubTitle>
          German Editions of Euclid&apos;s Elements in the 16th-17th Centuries
        </SubTitle>
        <div>Mia Joskowicz, September 2024</div>
        <Spacer />
        <StyledAnchor href={pdfPath} target="_blank" rel="noopener noreferrer">
          Presentation Slides
        </StyledAnchor>
        <Spacer size="sm" />
        <Block>
          <StyledLink to={MAP_ROUTE}>Elements Timeline Map</StyledLink>
          <div>Best viewed on desktop</div>
        </Block>
        <Spacer />
        <Proposition5Book2V1 />
      </Container>
    </Wrapper>
  );
}
