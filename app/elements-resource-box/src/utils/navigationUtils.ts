import { useNavigate } from "react-router-dom";
import { useCallback } from "react";
import { NO_FILTER_ROUTES } from "../components/layout/routes";

const SAVED_FILTER_QUERY_KEY = "saved-filter-query";

const saveFilterQuery = (query: string) => {
  if (!query) {
    return;
  }
  try {
    localStorage.setItem(SAVED_FILTER_QUERY_KEY, query);
  } catch {
    // no-op
  }
};

const getSavedFilterQuery = (): string => {
  try {
    return localStorage.getItem(SAVED_FILTER_QUERY_KEY) || "";
  } catch {
    return "";
  }
};

export const preserveQueryParams = (
  path: string,
  currentSearch: string,
): string => {
  if (NO_FILTER_ROUTES.includes(path)) {
    saveFilterQuery(currentSearch);
    return path;
  }

  if (path.includes("?")) {
    return path;
  }

  if (currentSearch) {
    return `${path}${currentSearch}`;
  }

  const savedQuery = getSavedFilterQuery();
  if (savedQuery) {
    return `${path}${savedQuery}`;
  }

  return path;
};

export const useNavigateWithQuery = () => {
  const navigate = useNavigate();

  return useCallback(
    (path: string) => {
      const pathWithQuery = preserveQueryParams(path, window.location.search);
      navigate(pathWithQuery);
    },
    [navigate],
  );
};
