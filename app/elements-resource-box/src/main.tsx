import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import "react-tooltip/dist/react-tooltip.css";
import { App } from "./App.tsx";
import { Tooltip } from "react-tooltip";
import {
  MapTooltips,
  TOOLTIP_FEATURES_HIGHLIGHT,
} from "./components/map/MapTooltips.tsx";
import { FilterEditProvider } from "./contexts/FilterEditContext.tsx";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

const queryClient = new QueryClient();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <FilterEditProvider>
        <App />
        <Tooltip
          id={TOOLTIP_FEATURES_HIGHLIGHT}
          delayHide={200}
          clickable
          style={{
            zIndex: 10000,
            backgroundColor: "white",
            color: "black",
            padding: "1rem",
            fontSize: "1.2rem",
            maxWidth: 600,
          }}
        />
        <MapTooltips />
      </FilterEditProvider>
    </QueryClientProvider>
  </StrictMode>,
);
