import { useContext } from "react";
import styled from "@emotion/styled";
import { useQuery } from "@tanstack/react-query";
import { AiFillEdit } from "react-icons/ai";
import { FaFilePdf } from "react-icons/fa";
import { SiMaterialdesign } from "react-icons/si";
import { FacsimilesService } from "@hub-api";
import { AuthContext } from "../contexts/Auth.ts";
import { LAND_COLOR } from "../utils/colors.ts";
import { withAppBasePath } from "../utils/basePath.ts";
import { openAuthenticatedFacsimilePDF } from "../utils/facsimilePdf.ts";
import { ITEM_EDIT_ROUTE } from "./layout/routes.ts";
import { TOOLTIP_SCAN } from "./map/MapTooltips.tsx";
import { FacsimileLinks } from "./FacsimileLinks.tsx";
import type { Item } from "../types";

const AnchorsRow = styled.div`
  display: flex !important;
  flex-direction: row;
  gap: 0.5rem;
`;

const StyledAnchor = styled.a`
  font-size: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: ${LAND_COLOR};
  text-decoration: none;
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

const StyledDiagramIcon = styled(SiMaterialdesign)`
  color: white !important;
  background-color: ${LAND_COLOR};
  width: 20px;
  height: 20px;
  border-radius: 4px;
`;

export const ItemLinksRow = ({
  item,
  showDiagramsLink = true,
  showEditLink = true,
}: {
  item: Item;
  showDiagramsLink?: boolean;
  showEditLink?: boolean;
}) => {
  const { token } = useContext(AuthContext);
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
  const shouldShow =
    item.facsimiles.length > 0 ||
    (Boolean(token) && hasMainScan) ||
    (showDiagramsLink && item.diagramCropsAvailable) ||
    (showEditLink && Boolean(token));

  if (!shouldShow) {
    return null;
  }

  return (
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
      {showEditLink && token && (
        <StyledAnchor
          href={withAppBasePath(`${ITEM_EDIT_ROUTE}?key=${item.key}`)}
          target="_blank"
          rel="noopener noreferrer"
          title="Edit Item"
        >
          <AiFillEdit />
        </StyledAnchor>
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
  );
};
