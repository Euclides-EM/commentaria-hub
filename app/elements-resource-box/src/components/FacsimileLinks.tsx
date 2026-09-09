import styled from "@emotion/styled";
import type { model_EditionShelfmark, model_Facsimile } from "@hub-api";
import type { HTMLAttributes } from "react";
import { FaBookReader, FaFilePdf } from "react-icons/fa";
import { PANE_BORDER } from "../utils/colors";
import { TOOLTIP_LINK } from "./map/MapTooltips";

const LinksRow = styled.div`
  display: flex !important;
  flex-direction: row;
  gap: 0.5rem;
  flex-wrap: wrap;
  min-width: 0;
  max-width: 100%;
`;

const FacsimileAnchor = styled.a<{ color: string }>`
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  color: ${({ color }) => color};

  svg {
    color: ${({ color }) => color};
  }
`;

const FacsimileButton = styled.button<{ color: string }>`
  position: relative;
  border: none;
  background: transparent;
  color: ${({ color }) => color};
  cursor: pointer;
  padding: 0;
  font-size: 1rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
`;

const FacsimileGroup = styled.div`
  display: flex;
  align-items: center;
  gap: 0.5rem;
`;

const VolumeBadge = styled.span`
  position: absolute;
  right: -0.4rem;
  bottom: -0.4rem;
  min-width: 0.7rem;
  height: 0.7rem;
  border-radius: 50%;
  background-color: white;
  color: ${PANE_BORDER};
  border: 1px solid ${PANE_BORDER};
  font-size: 0.5rem;
  font-weight: 700;
  line-height: 0.7rem;
  text-align: center;
`;

const toValidVolume = (value: unknown) => {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return null;
  }
  return value >= 1 ? value : null;
};

const getDisplayVolume = (
  facsimile: model_EditionShelfmark,
  assumeUnsetAsOne: boolean,
) => toValidVolume(facsimile.volume) ?? (assumeUnsetAsOne ? 1 : null);

const compareFacsimilesByVolume = (
  a: model_EditionShelfmark,
  b: model_EditionShelfmark,
  assumeUnsetAsOne: boolean,
) => {
  const volumeA = getDisplayVolume(a, assumeUnsetAsOne);
  const volumeB = getDisplayVolume(b, assumeUnsetAsOne);

  if (volumeA === null && volumeB === null) {
    return (a.scan || "").localeCompare(b.scan || "");
  }
  if (volumeA === null) {
    return -1;
  }
  if (volumeB === null) {
    return 1;
  }
  return volumeA - volumeB || (a.scan || "").localeCompare(b.scan || "");
};

const facsimileTitle = (shelfmark: model_EditionShelfmark) =>
  `View facsimile at ${shelfmark.scan}`;

const localFacsimileTitle = (shelfmark: model_EditionShelfmark) =>
  `Download PDF facsimile for ${shelfmark.scan}`;

export const FacsimileLinks = ({
  facsimiles,
  localFacsimiles = [],
  onOpenLocalFacsimile,
  isAuthenticated = false,
  color,
  className,
  ...props
}: {
  facsimiles: model_EditionShelfmark[];
  localFacsimiles?: model_Facsimile[];
  onOpenLocalFacsimile?: (facsimile: model_Facsimile) => void;
  isAuthenticated?: boolean;
  color: string;
  className?: string;
} & HTMLAttributes<HTMLDivElement>) => {
  const hasAnyValidVolume = facsimiles.some(
    (facsimile) => toValidVolume(facsimile.volume) !== null,
  );
  const shouldShowBadges = facsimiles.some(
    (facsimile) => (getDisplayVolume(facsimile, hasAnyValidVolume) ?? 0) >= 2,
  );
  const visibleFacsimiles = facsimiles
    .filter((facsimile) => facsimile.scan)
    .slice()
    .sort((a, b) => compareFacsimilesByVolume(a, b, hasAnyValidVolume));

  return (
    <LinksRow className={className} {...props}>
      {visibleFacsimiles.map((facsimile) => {
        const volume = getDisplayVolume(facsimile, hasAnyValidVolume);
        const linkedLocalFacsimiles = localFacsimiles.filter(
          (localFacsimile) =>
            localFacsimile.shelfmark_id === facsimile.id &&
            (isAuthenticated || Boolean(facsimile.copyright?.trim())),
        );
        return (
          <FacsimileGroup key={facsimile.scan}>
            <FacsimileAnchor
              href={facsimile.scan!}
              target="_blank"
              rel="noopener noreferrer"
              title={facsimileTitle(facsimile)}
              data-tooltip-id={TOOLTIP_LINK}
              data-tooltip-content={facsimileTitle(facsimile)}
              color={color}
            >
              <FaBookReader />
              {shouldShowBadges && volume && (
                <VolumeBadge>{volume}</VolumeBadge>
              )}
            </FacsimileAnchor>
            {linkedLocalFacsimiles.map((localFacsimile) => (
              <FacsimileButton
                key={localFacsimile.id}
                type="button"
                onClick={() => onOpenLocalFacsimile?.(localFacsimile)}
                title={localFacsimileTitle(facsimile)}
                aria-label={localFacsimileTitle(facsimile)}
                data-tooltip-id={TOOLTIP_LINK}
                data-tooltip-content={localFacsimileTitle(facsimile)}
                color={color}
              >
                <FaFilePdf />
                {shouldShowBadges && volume && (
                  <VolumeBadge>{volume}</VolumeBadge>
                )}
              </FacsimileButton>
            ))}
          </FacsimileGroup>
        );
      })}
      {localFacsimiles
        .filter(
          (localFacsimile) =>
            isAuthenticated &&
            !visibleFacsimiles.some(
              (facsimile) => facsimile.id === localFacsimile.shelfmark_id,
            ),
        )
        .map((localFacsimile) => {
          const shelfmark = facsimiles.find(
            (facsimile) => facsimile.id === localFacsimile.shelfmark_id,
          );
          const volume = shelfmark
            ? getDisplayVolume(shelfmark, hasAnyValidVolume)
            : null;
          return (
            <FacsimileButton
              key={localFacsimile.id}
              type="button"
              onClick={() => onOpenLocalFacsimile?.(localFacsimile)}
              title={shelfmark ? localFacsimileTitle(shelfmark) : "Facsimile"}
              aria-label={
                shelfmark ? localFacsimileTitle(shelfmark) : "Facsimile"
              }
              data-tooltip-id={TOOLTIP_LINK}
              data-tooltip-content={
                shelfmark ? localFacsimileTitle(shelfmark) : "Facsimile"
              }
              color={color}
            >
              <FaFilePdf />
              {shouldShowBadges && volume && (
                <VolumeBadge>{volume}</VolumeBadge>
              )}
            </FacsimileButton>
          );
        })}
    </LinksRow>
  );
};
