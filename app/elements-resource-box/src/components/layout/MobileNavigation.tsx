import { useLocation } from "react-router-dom";
import styled from "@emotion/styled";
import { HOME_ROUTE, MAIN_CONTENT_ID, NAVBAR_HEIGHT } from "./routes.ts";
import { MARKER_5 } from "../../utils/colors.ts";
import { useLayoutEffect, useState } from "react";
import { BsBoundingBoxCircles } from "react-icons/bs";
import { FaBars, FaTimes } from "react-icons/fa";
import { NavItems } from "./NavItem";
import { Row } from "../common.ts";
import { inEuclidesMode } from "../../utils/mode.ts";
import { useNavigateWithQuery } from "../../utils/navigationUtils";

const NavContainer = styled.nav`
  position: fixed;
  display: flex;
  width: 100vw;
  padding: 1rem;
  height: calc(${NAVBAR_HEIGHT}px - 2rem);
  background-color: black;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  z-index: 1001;
`;

const TitleContainer = styled.div`
  display: flex;
  align-items: center;
`;

const SiteTitle = styled.div`
  color: ${MARKER_5};
  white-space: nowrap;
  cursor: pointer;
  font-size: 0.9rem;
`;

const StyledBoxIcon = styled(BsBoundingBoxCircles)`
  margin-left: 0.5rem;
  color: ${MARKER_5};
  cursor: pointer;
`;

const MenuButton = styled.button`
  background: none;
  border: none;
  color: white;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
`;

const MobileMenu = styled.div<{ isOpen: boolean }>`
  position: fixed;
  top: ${NAVBAR_HEIGHT}px;
  left: 0;
  right: 0;
  background-color: black;
  z-index: 2002;
  padding: 0 1rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  transform: ${({ isOpen }) =>
    isOpen ? "translateY(0)" : "translateY(-100vh)"};
  transition: transform 0.3s ease-in-out;
`;

export function MobileNavigation() {
  const location = useLocation();
  const navigateWithQuery = useNavigateWithQuery();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  useLayoutEffect(() => {
    document.getElementById(MAIN_CONTENT_ID)?.scrollTo(0, 0);
    setIsMenuOpen(false);
  }, [location.pathname]);

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  return (
    <>
      <NavContainer>
        <Row gap={4}>
          <TitleContainer>
            <SiteTitle onClick={() => navigateWithQuery(HOME_ROUTE)}>
              {inEuclidesMode()
                ? "Commentaria in Euclidem"
                : "Euclid's Elements: A Resource Box"}
            </SiteTitle>
            <StyledBoxIcon onClick={() => navigateWithQuery(HOME_ROUTE)} />
          </TitleContainer>
          <MenuButton onClick={toggleMenu}>
            {isMenuOpen ? <FaTimes /> : <FaBars />}
          </MenuButton>
        </Row>
      </NavContainer>
      <MobileMenu isOpen={isMenuOpen}>
        <NavItems mobile />
      </MobileMenu>
    </>
  );
}
