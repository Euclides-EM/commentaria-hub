import { useContext } from "react";
import styled from "@emotion/styled";
import { useQuery } from "@tanstack/react-query";
import { AiFillEdit } from "react-icons/ai";
import { SiMaterialdesign } from "react-icons/si";
import { FacsimilesService } from "@hub-api";
import { AuthContext } from "../contexts/Auth.ts";
import { LAND_COLOR } from "../utils/colors.ts";
import { withAppBasePath } from "../utils/basePath.ts";
import { openAuthenticatedFacsimilePDF } from "../utils/facsimilePdf.ts";
import { ITEM_EDIT_ROUTE } from "./layout/routes.ts";
import { FacsimileLinks } from "./FacsimileLinks.tsx";
import { TOOLTIP_LINK } from "./map/MapTooltips.tsx";
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

const StyledDiagramIcon = styled(SiMaterialdesign)`
  color: white !important;
  background-color: ${LAND_COLOR};
  width: 20px;
  height: 20px;
  border-radius: 4px;
`;

const diagramsTitle = () => "View diagrams";

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
  const localScans =
    localFacsimilesQuery.data?.filter(
      (facsimile) => facsimile.id && facsimile.download_available,
    ) ?? [];
  const localScansWithDiagrams =
    localFacsimilesQuery.data?.filter(
      (facsimile) => facsimile.id && facsimile.diagram_crops_available,
    ) ?? [];
  const openLocalScan = (facsimileId: string, name?: string) => {
    if (!token) {
      return;
    }
    void openAuthenticatedFacsimilePDF(
      facsimileId,
      token,
      undefined,
      name ? `${name}.pdf` : undefined,
    ).catch((error) => {
      console.error("Failed to open scan:", error);
    });
  };
  const shouldShow =
    item.facsimiles.length > 0 ||
    (Boolean(token) && localScans.length > 0) ||
    (showDiagramsLink && item.diagramCropsAvailable) ||
    (showEditLink && Boolean(token));

  if (!shouldShow) {
    return null;
  }

  const editUrl = withAppBasePath(`${ITEM_EDIT_ROUTE}?key=${item.key}`);
  const diagramsUrl = withAppBasePath(`/diagrams?key=${item.key}`);

  return (
    <AnchorsRow>
      <FacsimileLinks
        facsimiles={item.facsimiles}
        localFacsimiles={localScans}
        onOpenLocalFacsimile={(facsimile) =>
          openLocalScan(facsimile.id!, facsimile.name)
        }
        color={LAND_COLOR}
      />
      {showEditLink && token && (
        <StyledAnchor
          href={editUrl}
          target="_blank"
          rel="noopener noreferrer"
          title="Edit item"
          data-tooltip-id={TOOLTIP_LINK}
          data-tooltip-content="Edit item"
        >
          <AiFillEdit />
        </StyledAnchor>
      )}
      {showDiagramsLink &&
        (localScansWithDiagrams.length > 0 || item.diagramCropsAvailable) && (
          <StyledAnchor
            href={diagramsUrl}
            target="_blank"
            rel="noopener noreferrer"
            title={diagramsTitle()}
            data-tooltip-id={TOOLTIP_LINK}
            data-tooltip-content={diagramsTitle()}
          >
            <StyledDiagramIcon />
          </StyledAnchor>
        )}
    </AnchorsRow>
  );
};
