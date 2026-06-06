import { useContext, useState } from "react";
import styled from "@emotion/styled";
import { FaCheck, FaFilePdf, FaQuoteLeft } from "react-icons/fa";
import { AiFillEdit } from "react-icons/ai";
import { Item } from "../../../types";
import { Row } from "../../common";
import { ModalTextColumn } from "./ModalComponents";
import { personDisplayName } from "../../../utils/dataUtils";
import { formatBookRanges, joinArr } from "../../../utils/util";
import { withAppBasePath } from "../../../utils/basePath";
import { NO_EDITOR } from "../../../constants";
import { LAND_COLOR } from "../../../utils/colors";
import { TOOLTIP_SCAN } from "../../map/MapTooltips";
import { SiMaterialdesign } from "react-icons/si";
import pluralize from "pluralize";
import { ITEM_EDIT_ROUTE } from "../../layout/routes.ts";
import { AuthContext } from "../../../contexts/Auth.ts";
import { useQuery } from "@tanstack/react-query";
import { getCommentariaHubPreferredTranscriptionUrl } from "../../../utils/commentariaHub.ts";
import { FacsimileLinks } from "../../FacsimileLinks.tsx";
import { FacsimilesService } from "@hub-api";
import { openAuthenticatedFacsimilePDF } from "../../../utils/facsimilePdf.ts";

const InfoTitle = styled.div`
  font-size: 0.8rem;
  color: darkgray;
  min-width: 4rem;
`;

const StyledAnchor = styled.a`
  font-size: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
`;

const IconButton = styled.button`
  border: none;
  background: transparent;
  color: ${LAND_COLOR};
  cursor: pointer;
  padding: 0;
  font-size: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
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

const ActionLink = styled.a`
  font-size: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: ${LAND_COLOR};
  text-decoration: none;
`;

const CommentariaHubLink = styled.a`
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
  max-width: 200px;
  white-space: normal;
  overflow-wrap: anywhere;

  &:hover {
    background-color: #e0e0e0;
  }
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

const getPersonLastName = (person: string) => {
  const displayName = personDisplayName(person);
  return displayName.split(",")[0].trim();
};

const generateCitation = (item: Item) => {
  const year = item.year || "s.d.";
  const firstEditor = item.editors[0];
  if (!firstEditor || firstEditor === NO_EDITOR) {
    return `s.n. ${year}`;
  }

  const lastNames = item.editors.map((a) => getPersonLastName(a));
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
  const localFacsimilesQuery = useQuery({
    queryKey: ["facsimiles", "download-available", item.key],
    queryFn: () => FacsimilesService.getFacsimilies({ editionId: [item.key] }),
    enabled: Boolean(token && item.key),
  });
  const hasMainScan = localFacsimilesQuery.data?.some(
    (facsimile) => facsimile.download_available,
  );
  const openMainScan = () => {
    if (!token) {
      return;
    }
    void openAuthenticatedFacsimilePDF(item.key, token).catch((error) => {
      console.error("Failed to open main scan:", error);
    });
  };

  return (
    <ModalTextColumn isRow={isRow}>
      <Row justifyStart>
        <InfoTitle>Year: </InfoTitle>
        {item.year || "s.d."}
      </Row>
      <Row justifyStart>
        <InfoTitle>{pluralize("Editor", item.editors.length)}:</InfoTitle>{" "}
        {joinArr(item.editors) || NO_EDITOR}
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
      {(item.facsimiles.length > 0 ||
        hasMainScan ||
        (showDiagramsLink && item.diagramCropsAvailable) ||
        token) && (
        <Row justifyStart>
          <InfoTitle>Links:</InfoTitle>
          <AnchorsRow
            data-tooltip-id={TOOLTIP_SCAN}
            data-tooltip-content="View Facsimile Online"
            data-tooltip-place="left"
          >
            <FacsimileLinks facsimiles={item.facsimiles} color={LAND_COLOR} />
            {hasMainScan && (
              <IconButton
                type="button"
                onClick={openMainScan}
                title="View main scan"
                aria-label="View main scan"
              >
                <FaFilePdf />
              </IconButton>
            )}
            {token && (
              <ActionLink
                href={withAppBasePath(`${ITEM_EDIT_ROUTE}?key=${item.key}`)}
                target="_blank"
                rel="noopener noreferrer"
                title="Edit Item"
              >
                <AiFillEdit />
              </ActionLink>
            )}
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
        </ActionsRow>
      )}
    </ModalTextColumn>
  );
};
