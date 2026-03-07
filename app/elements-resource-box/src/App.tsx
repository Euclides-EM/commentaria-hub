import {
  createBrowserRouter,
  createRoutesFromElements,
  Navigate,
  Route,
  RouterProvider,
} from "react-router-dom";
import { HomeResourceBox } from "./pages/HomeResourceBox.tsx";
import { Gallery } from "./pages/Gallery.tsx";
import { Layout } from "./components/layout/Layout";
import { TourProvider } from "@reactour/tour";
import { tourSteps } from "./components/map/Tour.tsx";
import { PANE_COLOR_ALT } from "./utils/colors.ts";
import { Map } from "./pages/Map.tsx";
import { UpsertEdition } from "./pages/UpsertEdition.tsx";
import {
  CATALOGUE_ROUTE,
  DIAGRAMS_ROUTE,
  GALLERY_ROUTE,
  HOME_ROUTE,
  ITEM_EDIT_ROUTE,
  LEGACY_EDITIONS_ROUTE,
  MAP_ROUTE,
  PRESENTATION_ROUTE,
  TITLE_PAGES_ROUTE,
  TRENDS_ROUTE,
} from "./components/layout/routes.ts";
import { Catalogue } from "./pages/Catalogue.tsx";
import { Trends } from "./pages/trends/Trends.tsx";
import { Presentation } from "./pages/Presentation.tsx";
import { Diagrams } from "./pages/Diagrams.tsx";
import { useLocalStorage } from "usehooks-ts";
import { AuthContext } from "./contexts/Auth.ts";
import { inEuclidesMode } from "./utils/mode.ts";
import { HomeCommentaria } from "./pages/HomeCommentaria.tsx";
import { getRouterBasename } from "./utils/basePath.ts";
import { configureHubApi } from "./api/hubApiConfig.ts";
import { NuqsAdapter } from "nuqs/adapters/react-router/v7";
import { FilterAppliedProvider } from "./contexts/FilterAppliedContext.tsx";

export function App() {
  const [authToken, setAuthToken] = useLocalStorage<string | null>(
    "resource-box-auth",
    null,
  );
  configureHubApi(authToken);

  const router = createBrowserRouter(
    createRoutesFromElements(
      <Route
        element={
          <NuqsAdapter>
            <FilterAppliedProvider>
              <Layout />
            </FilterAppliedProvider>
          </NuqsAdapter>
        }
      >
        <Route
          path={HOME_ROUTE}
          element={inEuclidesMode() ? <HomeCommentaria /> : <HomeResourceBox />}
        />
        <Route
          path={GALLERY_ROUTE}
          element={<Gallery titlePagesModeOn={false} />}
        />
        <Route
          path={TITLE_PAGES_ROUTE}
          element={<Gallery titlePagesModeOn={true} />}
        />
        <Route
          path={LEGACY_EDITIONS_ROUTE}
          element={<Navigate replace to={GALLERY_ROUTE} />}
        />
        <Route path={CATALOGUE_ROUTE} element={<Catalogue />} />
        <Route path={TRENDS_ROUTE} element={<Trends />} />
        <Route path={PRESENTATION_ROUTE} element={<Presentation />} />
        <Route path={DIAGRAMS_ROUTE} element={<Diagrams />} />
        <Route path={ITEM_EDIT_ROUTE} element={<UpsertEdition />} />
        <Route
          path={MAP_ROUTE}
          element={
            <TourProvider
              steps={tourSteps}
              styles={{
                maskArea: (base) => ({ ...base, rx: 8 }),
                popover: (base) => ({
                  ...base,
                  "--reactour-accent": PANE_COLOR_ALT,
                  borderRadius: "0.5rem",
                }),
              }}
            >
              <Map />
            </TourProvider>
          }
        />
        <Route path="*" element={<Navigate replace to={HOME_ROUTE} />} />
      </Route>,
    ),
    { basename: getRouterBasename() },
  );

  return (
    <AuthContext.Provider
      value={{
        token: authToken,
        setToken: setAuthToken,
      }}
    >
      <RouterProvider router={router} />
    </AuthContext.Provider>
  );
}
