import styled from "@emotion/styled";
import { css } from "@emotion/react";
import { Link, useNavigate } from "react-router-dom";
import { MARKER_4, MARKER_5, PANE_COLOR, SEA_COLOR } from "../../utils/colors";
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
    padding: 2rem 0.9rem calc(1.25rem + env(safe-area-inset-bottom, 0px));
    gap: clamp(0.2rem, 0.45vh, 0.45rem);
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
  cursor: pointer;

  @media only screen and (max-width: 500px) {
    font-size: 1.2rem;
  }
`;

const StyledAnchor = styled.a`
  ${linkStyle};
  color: ${MARKER_5};
  font-size: 1.8rem;
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

const Tile = styled.img`
  height: auto;
  width: auto;
  max-height: min(18rem, 27vh);
  max-width: min(24rem, 92vw);
  cursor: pointer;

  @media only screen and (max-width: 500px) {
    max-height: min(13rem, 24vh);
  }
`;

export function Nov24Page() {
  const navigate = useNavigate();
  const pdf = withAppBasePath(
    "/presentation/FrenchTranslationsOfEuclidsElementsInTheFirstHalfOfThe17thCentury.pdf",
  );
  const nov24Image = withAppBasePath("/presentation/nov24.png");
  const mapImage = withAppBasePath("/presentation/map.png");

  return (
    <Wrapper>
      <Container>
        <Title>
          French Translations of Euclid&apos;s Elements in the first half of the
          17th century
        </Title>
        <SubTitle>A Study of a Book in the Education Sphere</SubTitle>
        <div>Mia Joskowicz, November 2024</div>
        <Spacer />
        <StyledAnchor href={pdf} target="_blank" rel="noopener noreferrer">
          Presentation Slides
        </StyledAnchor>
        <Tile
          src={nov24Image}
          onClick={() => window.open(pdf, "_blank")?.focus()}
        />
        <Spacer size="sm" />
        <Block>
          <StyledLink to={MAP_ROUTE}>Elements Timeline Map</StyledLink>
          <Tile src={mapImage} onClick={() => navigate(MAP_ROUTE)} />
          <div>Best viewed on desktop</div>
        </Block>
        <Spacer />
        <Block>
          <StyledLink to="/sep24">
            Transformation of Mathematical Knowledge: German Editions of
            Euclid&apos;s Elements in the 16th-17th Centuries
          </StyledLink>
          <div>Talk, September 2024</div>
        </Block>
      </Container>
    </Wrapper>
  );
}
