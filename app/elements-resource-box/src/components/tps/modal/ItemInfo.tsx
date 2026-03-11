import { useContext, useState } from "react";
import styled from "@emotion/styled";
import { FaBookReader, FaCheck, FaQuoteLeft } from "react-icons/fa";
import { Item } from "../../../types";
import { Row } from "../../common";
import { ModalTextColumn } from "./ModalComponents";
import { authorDisplayName } from "../../../utils/dataUtils";
import { formatBookRanges, joinArr } from "../../../utils/util";
import { withAppBasePath } from "../../../utils/basePath";
import { NO_AUTHOR } from "../../../constants";
import { LAND_COLOR } from "../../../utils/colors";
import { TOOLTIP_SCAN } from "../../map/MapTooltips";
import { SiMaterialdesign } from "react-icons/si";
import pluralize from "pluralize";
import { ITEM_EDIT_ROUTE } from "../../layout/routes.ts";
import { AuthContext } from "../../../contexts/Auth.ts";
import { useQuery } from "@tanstack/react-query";
import { getCommentariaHubPreferredTranscriptionUrl } from "../../../utils/commentariaHub.ts";

const InfoTitle = styled.div`
  font-size: 0.8rem;
  color: darkgray;
  min-width: 4rem;
`;

const StyledAnchor = styled.a`
  font-size: 1rem;
  svg {
    color: ${LAND_COLOR};
    margin-bottom: -2px;
  }
`;

const AnchorsRow = styled.div`
  display: flex !important;
  flex-direction: row;
  gap: 0.5rem;
`;

const CitationButton = styled.button<{ copied?: boolean }>`
  background: none;
  border: none;
  color: ${({ copied }) => (copied ? "#28a745" : LAND_COLOR)};
  cursor: pointer;
  padding: 0.25rem;
  margin-left: 0.5rem;
  border-radius: 3px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s ease;

  &:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }

  svg {
    font-size: 0.8rem;
  }
`;

const StyledDiagramIcon = styled(SiMaterialdesign)`
  color: white !important;
  background-color: ${LAND_COLOR};
  width: 20px;
  height: 20px;
  border-radius: 4px;
`;

const EditLink = styled.a`
  display: block;
  margin-left: 0;
  padding: 0.5rem;
  background-color: #f0f0f0;
  border: 1px solid #ddd;
  text-decoration: none;
  color: #333;
  text-align: center;
  font-size: 0.8rem;
  border-radius: 0.5rem;
  height: fit-content;
  width: max-content;
  justify-self: end;

  &:hover {
    background-color: #e0e0e0;
  }
`;

const CommentariaHubLink = styled(EditLink)`
  max-width: 200px;
  white-space: normal;
  overflow-wrap: anywhere;
`;

const ActionsRow = styled.div`
  display: flex !important;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-content: flex-end;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
`;

const getAuthorLastName = (author: string) => {
  const displayName = authorDisplayName(author);
  return displayName.split(",")[0].trim();
};

const generateCitation = (item: Item) => {
  const year = item.year || "s.d.";
  const firstAuthor = item.authors[0];
  if (!firstAuthor || firstAuthor === NO_AUTHOR) {
    return `s.n. ${year}`;
  }

  const lastNames = item.authors.map((a) => getAuthorLastName(a));
  if (lastNames.length === 1) {
    return `${lastNames[0]} ${year}`;
  }
  if (lastNames.length > 3) {
    return `${lastNames[0]} et al. ${year}`;
  }
  return `${lastNames.slice(0, lastNames.length - 1).join(", ")}, and ${lastNames[lastNames.length - 1]} ${year}`;
};

const copyCitation = async (
  item: Item,
  setCopied: (copied: boolean) => void,
) => {
  const citation = generateCitation(item);
  await navigator.clipboard.writeText(citation);
  setCopied(true);
  setTimeout(() => setCopied(false), 2000);
};

export const ItemInfo = ({
  item,
  isRow,
  showDiagramsLink = true,
}: {
  item: Item;
  isRow?: boolean;
  showDiagramsLink?: boolean;
}) => {
  const [copied, setCopied] = useState(false);
  const { token } = useContext(AuthContext);
  const preferredTranscriptionLinkQuery = useQuery({
    queryKey: ["commentaria-hub-preferred-transcription", item.key],
    queryFn: () => getCommentariaHubPreferredTranscriptionUrl(item.key),
    enabled: !!item.key,
  });

  return (
    <ModalTextColumn isRow={isRow}>
      <Row justifyStart>
        <InfoTitle>Year: </InfoTitle>
        {item.year || "s.d."}
      </Row>
      <Row justifyStart>
        <InfoTitle>{pluralize("Author", item.authors.length)}:</InfoTitle>{" "}
        {joinArr(item.authors) || NO_AUTHOR}
        <CitationButton
          copied={copied}
          onClick={() => copyCitation(item, setCopied)}
          title={
            copied
              ? "Copied to clipboard!"
              : "Copy Chicago author-date in-text citation"
          }
        >
          {copied ? <FaCheck /> : <FaQuoteLeft />}
        </CitationButton>
      </Row>
      <Row justifyStart>
        <InfoTitle>{pluralize("City", item.cities.length)}:</InfoTitle>{" "}
        {joinArr(item.cities)}
      </Row>
      <Row justifyStart>
        <InfoTitle>{pluralize("Language", item.languages.length)}:</InfoTitle>{" "}
        {joinArr(item.languages)}
      </Row>
      {item.scanUrl && item.scanUrl.length > 0 && (
        <Row justifyStart>
          <InfoTitle>Facsimile:</InfoTitle>
          <AnchorsRow
            data-tooltip-id={TOOLTIP_SCAN}
            data-tooltip-content="View Facsimile Online"
            data-tooltip-place="left"
          >
            {item.scanUrl.map((url, index) => (
              <StyledAnchor
                key={index}
                href={url}
                target="_blank"
                rel="noopener noreferrer"
              >
                <FaBookReader />
              </StyledAnchor>
            ))}
            {showDiagramsLink && item.diagramCropsAvailable && (
              <StyledAnchor
                href={withAppBasePath(`/diagrams?key=${item.key}`)}
                target="_blank"
                rel="noopener noreferrer"
                title="View Diagrams"
              >
                <StyledDiagramIcon />
              </StyledAnchor>
            )}
          </AnchorsRow>
        </Row>
      )}
      {item.format && (
        <Row justifyStart>
          <InfoTitle>Format:</InfoTitle> {item.format}
        </Row>
      )}
      {item.volumesCount && (
        <Row justifyStart>
          <InfoTitle>{pluralize("Volume", item.volumesCount)}:</InfoTitle>{" "}
          {item.volumesCount}
        </Row>
      )}
      {item.elementsBooks && (
        <Row justifyStart>
          <InfoTitle>Books:</InfoTitle> {formatBookRanges(item.elementsBooks)}
        </Row>
      )}
      {item.class && (
        <>
          <Row justifyStart>
            <InfoTitle>Class:</InfoTitle> {item.class}
          </Row>
        </>
      )}
      {item.additionalContent && item.additionalContent.length > 0 && (
        <Row justifyStart>
          <InfoTitle>Additional Content:</InfoTitle>{" "}
          {joinArr(item.additionalContent)}
        </Row>
      )}

      {(preferredTranscriptionLinkQuery.data || token) && (
        <ActionsRow>
          {preferredTranscriptionLinkQuery.data && (
            <CommentariaHubLink
              href={preferredTranscriptionLinkQuery.data}
              target="_blank"
              rel="noopener noreferrer"
            >
              View transcription in Commentaria Hub
            </CommentariaHubLink>
          )}
          {token && (
            <EditLink
              href={withAppBasePath(`${ITEM_EDIT_ROUTE}?key=${item.key}`)}
              target="_blank"
              rel="noopener noreferrer"
            >
              Edit Item
            </EditLink>
          )}
        </ActionsRow>
      )}
    </ModalTextColumn>
  );
};
