import { Link, useLocation } from "react-router-dom";
import styled from "@emotion/styled";
import { FaExternalLinkAlt } from "react-icons/fa";
import { css } from "@emotion/react";
import {
  CATALOGUE_ROUTE,
  GALLERY_ROUTE,
  HOME_ROUTE,
  ITEM_EDIT_ROUTE,
  MAP_ROUTE,
  TITLE_PAGES_ROUTE,
  TRENDS_ROUTE,
} from "./routes.ts";
import { MACTUTOR_URL } from "../../constants";
import { ActionsMenu } from "./ActionsMenu.tsx";
import { inEuclidesMode } from "../../utils/mode.ts";
import {
  preserveQueryParams,
  useNavigateWithQuery,
} from "../../utils/navigationUtils";
import { withAppBasePath } from "../../utils/basePath";

const separatorStyles = css`
  &::after {
    content: "·";
    color: #666666;
    font-size: 1.3rem;
    line-height: 1;
    margin-left: 0.6rem;
  }
`;

const StyledNavItem = styled.li<{
  active: boolean;
  mobile: boolean;
  hideSeparator?: boolean;
}>`
  display: flex;
  align-items: center;
  position: relative;

  a {
    text-decoration: none;
    color: ${({ active }) =>
      active ? "#ffffff" : inEuclidesMode() ? "#cecece" : "#aaaaaa"};
    font-weight: ${({ active }) => (active ? "bold" : "normal")};
    font-size: 1rem;
    line-height: 1;
    transition: color 0.3s ease;
    display: block;

    &:hover {
      color: white;
    }
  }

  ${({ mobile }) =>
    mobile &&
    css`
      a {
        padding: 0.5rem 0;
        font-size: 0.95rem;
      }
    `};
  ${({ mobile, hideSeparator }) =>
    !mobile &&
    !hideSeparator &&
    css`
      ${separatorStyles};
    `};
`;

const StyledExternalIcon = styled(FaExternalLinkAlt)`
  font-size: 0.8rem;
  margin-left: 0.5rem;
`;

const DevNavItem = styled.li<{ mobile: boolean; hideSeparator?: boolean }>`
  display: flex;
  align-items: center;

  ${({ mobile }) =>
    mobile &&
    css`
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid #555;
      margin-bottom: 1rem;
    `};
  ${({ mobile }) =>
    !mobile &&
    css`
      margin-left: auto;
      margin-right: 2rem;
    `};
  ${({ mobile, hideSeparator }) =>
    !mobile &&
    !hideSeparator &&
    css`
      ${separatorStyles};
    `};
`;

interface NavItemProps {
  to: string;
  active: boolean;
  external?: boolean;
  children: React.ReactNode;
  className?: string;
  mobile: boolean;
  hideSeparator?: boolean;
}

function NavItem({
  to,
  active,
  external = false,
  children,
  className,
  mobile,
  hideSeparator = false,
}: NavItemProps) {
  const navigateWithQuery = useNavigateWithQuery();

  if (external) {
    return (
      <StyledNavItem
        active={active}
        className={className}
        mobile={mobile}
        hideSeparator={hideSeparator}
      >
        <Link to={to} target="_blank" rel="noreferrer noopener">
          {children} <StyledExternalIcon />
        </Link>
      </StyledNavItem>
    );
  }

  return (
    <StyledNavItem
      active={active}
      className={className}
      mobile={mobile}
      hideSeparator={hideSeparator}
    >
      <Link
        to={to}
        onClick={(e) => {
          e.preventDefault();
          navigateWithQuery(to);
        }}
      >
        {children}
      </Link>
    </StyledNavItem>
  );
}

const NavList = styled.ul<{ mobile: boolean }>`
  list-style: none;
  margin: 0;
  padding: 0;
  ${({ mobile }) =>
    !mobile &&
    css`
      display: flex;
      align-items: center;
      flex: 1;
      min-width: 0;
      gap: 0.6rem;
    `};
`;

export const NavItems = ({ mobile }: { mobile: boolean }) => {
  const location = useLocation();

  return (
    <>
      <NavList mobile={mobile}>
        <NavItem
          to={HOME_ROUTE}
          active={location.pathname === HOME_ROUTE}
          mobile={mobile}
        >
          Home
        </NavItem>
        <NavItem
          to={CATALOGUE_ROUTE}
          active={location.pathname === CATALOGUE_ROUTE}
          mobile={mobile}
        >
          Catalogue
        </NavItem>
        <NavItem
          to={GALLERY_ROUTE}
          active={location.pathname === GALLERY_ROUTE}
          mobile={mobile}
        >
          Gallery
        </NavItem>
        {!inEuclidesMode() && (
          <NavItem
            to={TITLE_PAGES_ROUTE}
            active={location.pathname === TITLE_PAGES_ROUTE}
            mobile={mobile}
          >
            Title Pages
          </NavItem>
        )}
        <NavItem
          to={TRENDS_ROUTE}
          active={location.pathname === TRENDS_ROUTE}
          mobile={mobile}
        >
          Explorer
        </NavItem>
        <NavItem
          to={MAP_ROUTE}
          active={location.pathname === MAP_ROUTE}
          mobile={mobile}
          hideSeparator={inEuclidesMode()}
        >
          Map
        </NavItem>
        {!inEuclidesMode() && (
          <NavItem
            to={MACTUTOR_URL}
            active={false}
            external
            mobile={mobile}
            hideSeparator
          >
            MacTutor Index Graph
          </NavItem>
        )}
        <DevNavItem mobile={mobile} hideSeparator>
          <ActionsMenu
            mobile={mobile}
            onShowCreateModal={() => {
              window.location.href = withAppBasePath(
                preserveQueryParams(ITEM_EDIT_ROUTE, window.location.search),
              );
            }}
          />
        </DevNavItem>
      </NavList>
    </>
  );
};
