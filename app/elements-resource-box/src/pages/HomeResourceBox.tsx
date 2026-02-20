import styled from "@emotion/styled";
import { Container, Row, Text } from "../components/common";
import { LAND_COLOR, MARKER_4, MARKER_5, PANE_COLOR } from "../utils/colors.ts";
import {
  CATALOGUE_ROUTE,
  MAIN_CONTENT_ID,
  NAVBAR_HEIGHT,
  TITLE_PAGES_ROUTE,
} from "../components/layout/routes.ts";
import { FaDraftingCompass } from "react-icons/fa";
import { GiHolySymbol } from "react-icons/gi";
import { ReactNode, useEffect, useState } from "react";
import { css } from "@emotion/react";
import { EIP_URL, MACTUTOR_URL } from "../constants";
import { TbMathMaxMin } from "react-icons/tb";
import { useEditFilter } from "../contexts/FilterEditContext.tsx";
import { useNavigateWithQuery } from "../utils/navigationUtils";
import { withAppBasePath } from "../utils/basePath";

const SCAN_BACKGROUND = withAppBasePath("/scan-v2.png");

const ParallaxBackground = styled.div`
  position: fixed;
  top: ${NAVBAR_HEIGHT}px;
  left: 0;
  width: 100vw;
  height: 100%;
  background-image: url("${SCAN_BACKGROUND}");
  background-size: cover;
  background-position: top center;
  background-repeat: no-repeat;
  z-index: -1;
  transform: translateY(0);
  will-change: transform;
  @media (max-width: 600px) {
    display: none;
  }
`;

const StyledContainer = styled(Container)`
  padding-top: 2rem;
  z-index: 1;
  padding-bottom: 2rem;
  max-width: 70vw;
  margin: auto;
  @media (max-width: 600px) {
    max-width: 96vw;
  }
`;

const Title = styled.div`
  font-size: 6rem;
  margin: -2rem;
  max-width: 100%;
  @media (max-width: 600px) {
    font-size: 4rem;
    margin: 0;
  }
`;

const SubTitle = styled(Text)`
  @media (max-width: 600px) {
    font-size: 2rem;
  }
`;

const TextStyle = css`
  background-color: #282828;
  border-radius: 0.5rem;
  padding: 0.5rem;
  width: auto;
  font-size: 1.2rem;
`;

const CardText = styled.div`
  ${TextStyle};
`;

const StyledImage = styled.img`
  max-width: 30vw;
  max-height: 50vh;
  object-fit: contain;
  border-radius: 0.5rem;
  @media (max-width: 600px) {
    max-width: 90vw;
  }
`;

const Credits = styled(Row)`
  color: ${LAND_COLOR};
  ${TextStyle};
  margin-top: 4rem;
  text-align: center;

  a {
    color: ${MARKER_4};
  }
`;

const Card = ({
  title,
  icon,
  color,
  text,
  imageSrc,
  imageOnLeft,
  onClick,
}: {
  title?: string;
  icon?: ReactNode;
  onClick?: () => void;
  color?: string;
  imageSrc: string;
  imageOnLeft: boolean;
  text: string | ReactNode;
}) => {
  if (window.innerWidth < 600) {
    imageOnLeft = true;
  }
  return (
    <Row noWrap>
      {imageOnLeft && <StyledImage src={imageSrc} alt={title} />}
      <Row column gap={0.5}>
        {title && (
          <CardText>
            <Text
              color={color}
              size={1.8}
              onClick={onClick}
              clickable={!!onClick}
            >
              {title}
            </Text>
          </CardText>
        )}
        {icon && (
          <Text
            size={1.5}
            color={color}
            onClick={onClick}
            clickable={!!onClick}
          >
            {icon}
          </Text>
        )}
        <CardText>
          {typeof text === "string"
            ? text
                .trim()
                .split("\n")
                .map((line, i) => <div key={i}>{line}</div>)
            : text}
        </CardText>
      </Row>
      {!imageOnLeft && <StyledImage src={imageSrc} alt={title} />}
    </Row>
  );
};

export function HomeResourceBox() {
  const navigateWithQuery = useNavigateWithQuery();
  const [imageHeight, setImageHeight] = useState(0);
  const { setFilterOpen } = useEditFilter();

  useEffect(() => {
    setFilterOpen(false);
  }, [setFilterOpen]);

  useEffect(() => {
    const img = new Image();
    img.src = withAppBasePath("/scan-v2.png");
    const bgElement = document.getElementById("parallax-bg");
    if (!bgElement) {
      return;
    }

    img.onload = () => {
      const imageAspectRatio = img.naturalHeight / img.naturalWidth;
      const topOffset = NAVBAR_HEIGHT;

      const mainContent = document.getElementById(MAIN_CONTENT_ID);

      const handleScroll = () => {
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;
        const scrollTop = mainContent ? mainContent.scrollTop : 0;
        const imageHeight = viewportWidth * imageAspectRatio;
        setImageHeight(imageHeight);

        const effectiveBackgroundHeight = viewportHeight - topOffset;
        const imageScrollRange = imageHeight - effectiveBackgroundHeight;
        const pageScrollRange = mainContent
          ? mainContent.scrollHeight - mainContent.clientHeight
          : 1;

        const scrollRatio = scrollTop / pageScrollRange;
        const translateY = scrollRatio * imageScrollRange;

        bgElement.style.transform = `translateY(-${translateY}px)`;
      };

      handleScroll();

      mainContent?.addEventListener("scroll", handleScroll, { passive: true });
      window.addEventListener("resize", handleScroll);

      return () => {
        mainContent?.removeEventListener("scroll", handleScroll);
        window.removeEventListener("resize", handleScroll);
      };
    };
  }, []);

  return (
    <>
      <ParallaxBackground
        id="parallax-bg"
        style={{
          height: imageHeight,
        }}
      />
      <StyledContainer>
        <Row column gap={1}>
          <Title className="gothic">Euclid's Elements</Title>
          <SubTitle size={3} color={MARKER_5}>
            A RESOURCE BOX
          </SubTitle>
        </Row>
        <Card
          title="The Project"
          icon={<FaDraftingCompass />}
          color={PANE_COLOR}
          imageSrc={withAppBasePath("/frontispiece.png")}
          imageOnLeft={true}
          text={
            <div>
              This web application is a living resource book that accompanies my
              PhD research in the area of early modern transmission of Euclid’s{" "}
              <i>Elements</i>. This project is rooted in two realizations I had
              early in my PhD journey. First, a defining feature of a PhD is
              that there is no source book or textbook to rely on; if there
              were, it would hardly count as a PhD. Second, reading mathematical
              works taught me that one of the best ways to understand something
              is to explain and demonstrate it. As an engineer, coding is a
              primary way for me to express myself, so I began building this
              site as my evolving handbook. It’s a playground, not a finished
              product, and it keeps changing alongside the research itself.
            </div>
          }
        />
        <Card
          imageSrc={withAppBasePath("/map.png")}
          imageOnLeft={false}
          text={
            <div>
              The Catalog, Editions Gallery, and Map tabs offer three different
              ways to explore the editions of Euclidean and related texts. The
              Catalog lists editions in a table, each with key details and a
              link to a digital facsimile when available. The Gallery presents
              the same information with thumbnails and an option to expand and
              view the full details on each edition. The Map places the editions
              geographically along a timeline. All three tabs share a filtering
              function that preserves the search parameters as you navigate
              between them. The corpus is based primarily on the{" "}
              <a href={EIP_URL} target="_blank" rel="noopener noreferrer">
                Euclid in print (1482–1703) catalog
              </a>
              , supplemented with references from scholarship, USTC, the BnF
              collection, and Google Books. The distinction between edition,
              reimpression, version and the classification of a book as a
              translation of Elements or "other" is not always clear and can be
              challenged.
            </div>
          }
        />
        <Card
          title="Title Pages Experiment"
          onClick={() => navigateWithQuery(CATALOGUE_ROUTE)}
          icon={<GiHolySymbol />}
          color={MARKER_5}
          imageSrc={`${import.meta.env.VITE_BACKEND_URL}/store/data/tps/Paris_1622_tp.jpeg`}
          imageOnLeft={true}
          text={
            <div>
              The Gallery tab includes a toggle that lets you view an experiment
              in analyzing the title pages of the editions. This experiment is
              based on the idea that title pages serve multiple roles. They
              designate the book’s identity and can be viewed as an instrument
              of intellectual and commercial positioning, they can reflect
              contemporary aesthetic and typographical conventions and they
              advertise the book’s contents to its intended audience and encode
              broader intellectual and disciplinary priorities. In the case of
              mathematical works such as <i>Elements</i>, title pages can reveal
              the pedagogical and epistemological frameworks and networks within
              which mathematics was studied, taught, and disseminated. Thus,
              they can offer insight into both the circulation of{" "}
              <i>Elements</i> and the social, intellectual, and commercial
              forces shaping its transmission.
            </div>
          }
        />
        <Card
          onClick={() => navigateWithQuery(TITLE_PAGES_ROUTE)}
          imageSrc={withAppBasePath("/modal.png")}
          imageOnLeft={false}
          text={
            <div>
              The experiment involved segmenting the title page's text into
              distinct elements to identify patterns and variations. The
              transcription and segmentation of the title pages were done partly
              by an LLM and partly by automated processes. As a result, the data
              may contain some unusual errors but also fewer human mistakes and
              inconsistencies. The hands-on work with multiple models was part
              of the experience I aimed to gain from this exploration.
            </div>
          }
        />
        <Card
          title="MacTutor Index Graph"
          onClick={() => window.open(MACTUTOR_URL, "_blank")}
          icon={<TbMathMaxMin />}
          color={MARKER_4}
          imageSrc={withAppBasePath("/mactutor.png")}
          imageOnLeft={true}
          text={
            <div>
              Another LLM experiment is the MacTutor Index Graph. This project
              grew out of my attempts to make sense of the many mathematicians
              and practitioners active during the period, leading me to rely on{" "}
              <a
                href="https://mathshistory.st-andrews.ac.uk/"
                target="_blank"
                rel="noopener noreferrer"
              >
                MacTutor Index
              </a>
              , an online resource which contains biographies of over 3,000
              mathematicians. Using an LLM to process that website, I extracted
              key information from each biography and mapped the connections
              between individuals, displaying everything as a graph. This
              project is still in its early stages and the LLM output has not
              yet been fully verified and refined.
            </div>
          }
        />

        <Credits column gap={0.5}>
          <div>
            Created and maintained by Mia Joskowicz as part of her PhD project
            with the assistance of Liri Sokol, 2025,{" "}
            <a
              href="https://creativecommons.org/licenses/by-nc/4.0/"
              target="_blank"
              rel="noopener noreferrer"
            >
              License
            </a>
          </div>
        </Credits>
      </StyledContainer>
    </>
  );
}
