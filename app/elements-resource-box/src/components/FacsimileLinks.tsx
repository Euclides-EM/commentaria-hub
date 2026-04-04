import styled from "@emotion/styled";
import type { model_EditionShelfmark } from "@hub-api";
import type { HTMLAttributes } from "react";
import { FaBookReader } from "react-icons/fa";
import { PANE_BORDER } from "../utils/colors";

const LinksRow = styled.div`
  display: flex;
  flex-direction: row;
  gap: 0.5rem;
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

export const FacsimileLinks = ({
  facsimiles,
  color,
  className,
  ...props
}: {
  facsimiles: model_EditionShelfmark[];
  color: string;
  className?: string;
} & HTMLAttributes<HTMLDivElement>) => {
  const hasAnyValidVolume = facsimiles.some(
    (facsimile) => toValidVolume(facsimile.volume) !== null,
  );
  const shouldShowBadges = facsimiles.some(
    (facsimile) => (getDisplayVolume(facsimile, hasAnyValidVolume) ?? 0) >= 2,
  );

  return (
    <LinksRow className={className} {...props}>
      {facsimiles
        .filter((facsimile) => facsimile.scan)
        .slice()
        .sort((a, b) => compareFacsimilesByVolume(a, b, hasAnyValidVolume))
        .map((facsimile) => {
          const volume = getDisplayVolume(facsimile, hasAnyValidVolume);
          return (
            <FacsimileAnchor
              key={facsimile.scan}
              href={facsimile.scan!}
              target="_blank"
              rel="noopener noreferrer"
              title="View Facsimile Online"
              color={color}
            >
              <FaBookReader />
              {shouldShowBadges && volume && (
                <VolumeBadge>{volume}</VolumeBadge>
              )}
            </FacsimileAnchor>
          );
        })}
    </LinksRow>
  );
};
