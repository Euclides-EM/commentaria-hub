import styled from "@emotion/styled";
import { CATALOGUE_ROUTE, NAVBAR_HEIGHT } from "../components/layout/routes";
import { useNavigateWithQuery } from "../utils/navigationUtils";

const BackgroundContainer = styled.div`
  position: fixed;
  top: ${NAVBAR_HEIGHT}px;
  left: 0;
  width: 100vw;
  height: calc(100vh - ${NAVBAR_HEIGHT}px);
  background:
    linear-gradient(rgba(0, 0, 0, 0.8), rgba(0, 0, 0, 0)),
    url("public/athens.jpg");
  background-size: cover;
  background-repeat: no-repeat;
  background-position: center bottom;
  z-index: -1;
`;

const ContentContainer = styled.div`
  font-family: "Manrope", "Helvetica Neue", Arial, sans-serif;
  height: calc(100vh - ${NAVBAR_HEIGHT}px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: start;
  padding-top: 20vh;
  text-align: center;
  position: relative;
  z-index: 1;
  overflow: hidden;
  box-sizing: border-box;
`;

const Title = styled.h1`
  font-size: 4.5rem;
  color: white;
  margin: 0;
  margin-bottom: 0.8rem;
  text-shadow: 2px 2px 8px rgba(0, 0, 0, 0.8);
  font-weight: 700;
  letter-spacing: -0.02em;

  @media (max-width: 768px) {
    font-size: 2rem;
  }
`;

const Subtitle = styled.p`
  font-size: 2rem;
  color: rgba(255, 255, 255, 0.85);
  margin: 0;
  max-width: 60%;
  line-height: 1.4;
  text-shadow: 1px 1px 4px rgba(0, 0, 0, 0.8);

  @media (max-width: 768px) {
    font-size: 1.5rem;
  }
`;

const BrowseButton = styled.button`
  margin-top: 4rem;
  background: white;
  opacity: 0.8;
  color: #333;
  border: none;
  padding: 1.1rem 2.6rem;
  font-size: 1.2rem;
  border-radius: 6px;
  font-weight: 900;
  cursor: pointer;
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);

  &:hover {
    opacity: 1;
    transform: translateY(-2px);
    box-shadow: 0 6px 12px rgba(0, 0, 0, 0.2);
  }

  @media (max-width: 768px) {
    margin-top: 1.5rem;
    font-size: 0.95rem;
    padding: 0.7rem 1.8rem;
  }
`;

const Features = styled.div`
  margin-top: 10vh;
  display: flex;
  gap: 1.5rem;
  color: rgba(255, 255, 255, 0.9);
  font-size: 1.5rem;
  text-shadow: 1px 1px 3px rgba(0, 0, 0, 0.8);
  width: 100%;
  padding: 1rem;
  justify-content: center;
  align-content: center;
  background-color: rgba(34, 34, 34, 0.6);

  @media (max-width: 768px) {
    padding: 0.5rem;
    flex-direction: column;
    gap: 0.3rem;
    font-size: 1.2rem;
    bottom: 1rem;
  }

  span:not(:last-child)::after {
    content: "·";
    margin-left: 2rem;

    @media (max-width: 768px) {
      display: none;
    }
  }
`;

export const HomeCommentaria = () => {
  const navigateWithQuery = useNavigateWithQuery();

  return (
    <>
      <BackgroundContainer />
      <ContentContainer>
        <Title>Commentaria in Euclidem</Title>
        <Subtitle>
          A platform for studying modern editions and commentaries of Euclid's
          Elements
        </Subtitle>
        <BrowseButton onClick={() => navigateWithQuery(CATALOGUE_ROUTE)}>
          Browse the Catalogue
        </BrowseButton>
        <Features>
          <span>Searchable Texts</span>
          <span>Structured Metadata</span>
          <span>Comparison and Analysis Tools</span>
        </Features>
      </ContentContainer>
    </>
  );
};
