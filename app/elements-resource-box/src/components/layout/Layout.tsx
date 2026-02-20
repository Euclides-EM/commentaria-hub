import { Outlet, useLocation } from "react-router-dom";
import { Navigation } from "./Navigation";
import styled from "@emotion/styled";
import { FilterPane } from "../map/FilterPane";
import { MAIN_CONTENT_ID, MAP_ROUTE, NAVBAR_HEIGHT } from "./routes.ts";
import { useEditFilter } from "../../contexts/FilterEditContext";

const LayoutContainer = styled.div`
  height: 100vh;
  width: 100vw;
`;

const ContentRow = styled.div`
  display: flex;
  height: calc(100vh - ${NAVBAR_HEIGHT}px);
`;

const MainContent = styled.main`
  flex: 1;
  min-width: 0;
  overflow-y: auto;

  div {
    overscroll-behavior-y: auto;
  }
`;

export function Layout() {
  const { filterOpen, setFilterOpen } = useEditFilter();
  const location = useLocation();
  const isMapRoute = location.pathname === MAP_ROUTE;

  return (
    <LayoutContainer>
      <Navigation />
      {isMapRoute && <FilterPane overlay />}
      <ContentRow>
        {!isMapRoute && <FilterPane />}
        <MainContent
          id={MAIN_CONTENT_ID}
          onClick={() => filterOpen && setFilterOpen(false)}
        >
          <Outlet />
        </MainContent>
      </ContentRow>
    </LayoutContainer>
  );
}
