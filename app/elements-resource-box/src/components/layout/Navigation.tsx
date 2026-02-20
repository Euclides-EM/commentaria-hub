import { useLocation } from "react-router-dom";
import styled from "@emotion/styled";
import {
  HOME_ROUTE,
  MAIN_CONTENT_ID,
  NAVBAR_HEIGHT,
  NO_FILTER_ROUTES,
} from "./routes.ts";
import { MARKER_5 } from "../../utils/colors.ts";
import { useLayoutEffect } from "react";
import { BsBoundingBoxCircles } from "react-icons/bs";
import { FilterButton } from "./FilterButton";
import { MobileNavigation } from "./MobileNavigation";
import { NavItems } from "./NavItem";
import { useIsMobile } from "./isMobile.ts";
import { inEuclidesMode } from "../../utils/mode.ts";
import { useNavigateWithQuery } from "../../utils/navigationUtils";

const Placeholder = styled.div`
  height: ${NAVBAR_HEIGHT}px;
`;

const NavContainer = styled.nav`
  position: fixed;
  display: flex;
  width: 100vw;
  padding: 1rem;
  height: calc(${NAVBAR_HEIGHT}px - 2rem);
  background-color: ${inEuclidesMode() ? "DarkSlateGray" : "black"};
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  z-index: 1001;
`;

const NavContent = styled.div`
  display: flex;
  align-items: center;
  width: 100%;
`;

const VerticalLine = styled.div`
  width: 1px;
  height: 20px;
  background-color: #aaaaaa;
  margin: 0 2rem;
`;

const SiteTitle = styled.div`
  color: ${MARKER_5};
  white-space: nowrap;
  cursor: pointer;
`;

const StyledBoxIcon = styled(BsBoundingBoxCircles)`
  margin-left: 0.5rem;
  color: ${MARKER_5};
  cursor: pointer;
`;

const FixedFilterButtonContainer = styled.div`
  position: fixed;
  top: calc(${NAVBAR_HEIGHT}px + 1rem);
  left: 1rem;
  z-index: 90;
`;

export function Navigation() {
  const location = useLocation();
  const navigateWithQuery = useNavigateWithQuery();
  const isMobile = useIsMobile();

  useLayoutEffect(() => {
    document.getElementById(MAIN_CONTENT_ID)?.scrollTo(0, 0);
  }, [location.pathname]);

  if (isMobile) {
    return (
      <>
        <MobileNavigation />
        <Placeholder />
        {!NO_FILTER_ROUTES.includes(location.pathname) && (
          <FixedFilterButtonContainer>
            <FilterButton />
          </FixedFilterButtonContainer>
        )}
      </>
    );
  }

  return (
    <>
      <NavContainer>
        <NavContent>
          <SiteTitle onClick={() => navigateWithQuery(HOME_ROUTE)}>
            {inEuclidesMode()
              ? "Commentaria in Euclidem"
              : "Euclid's Elements: A Resource Box"}
          </SiteTitle>
          <StyledBoxIcon onClick={() => navigateWithQuery(HOME_ROUTE)} />
          <VerticalLine />
          <NavItems mobile={false} />
        </NavContent>
      </NavContainer>
      <Placeholder />
      {!NO_FILTER_ROUTES.includes(location.pathname) && (
        <FixedFilterButtonContainer>
          <FilterButton />
        </FixedFilterButtonContainer>
      )}
    </>
  );
}
