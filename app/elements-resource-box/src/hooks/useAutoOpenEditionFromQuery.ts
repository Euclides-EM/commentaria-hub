import { useCallback, useEffect, useMemo, useRef } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import type { Item } from "../types";
import { getEdition } from "../api/editionApi";
import { mapEditionsToItems } from "../utils/dataUtils";

const EDITION_KEY_QUERY_PARAM = "editionKey";

const normalizeKey = (value: string) => value.trim().toLowerCase();

export function useAutoOpenEditionFromQuery(
  items: Item[] | null | undefined,
  onOpen: (item: Item) => void,
) {
  const location = useLocation();
  const navigate = useNavigate();
  const lastHandledKeyRef = useRef<string | null>(null);
  const inFlightKeyRef = useRef<string | null>(null);

  const editionKey = useMemo(() => {
    const value = new URLSearchParams(location.search).get(
      EDITION_KEY_QUERY_PARAM,
    );
    if (!value) {
      return null;
    }
    const trimmed = value.trim();
    return trimmed.length > 0 ? trimmed : null;
  }, [location.search]);

  const clearEditionKeyParam = useCallback(() => {
    const params = new URLSearchParams(location.search);
    if (!params.has(EDITION_KEY_QUERY_PARAM)) {
      return;
    }
    params.delete(EDITION_KEY_QUERY_PARAM);
    const nextSearch = params.toString();
    navigate(
      {
        pathname: location.pathname,
        search: nextSearch ? `?${nextSearch}` : "",
        hash: location.hash,
      },
      { replace: true },
    );
  }, [location.hash, location.pathname, location.search, navigate]);

  useEffect(() => {
    if (!editionKey) {
      return;
    }
    if (
      lastHandledKeyRef.current === editionKey ||
      inFlightKeyRef.current === editionKey
    ) {
      return;
    }

    const matchingItem = items?.find(
      (item) => normalizeKey(item.key) === normalizeKey(editionKey),
    );
    if (matchingItem) {
      onOpen(matchingItem);
      lastHandledKeyRef.current = editionKey;
      clearEditionKeyParam();
      return;
    }

    inFlightKeyRef.current = editionKey;
    let active = true;

    (async () => {
      try {
        const edition = await getEdition(editionKey).catch(() => null);
        if (!active || !edition) {
          return;
        }
        const mapped = mapEditionsToItems([edition])[0];
        if (mapped) {
          onOpen(mapped);
        }
      } finally {
        if (!active) {
          return;
        }
        lastHandledKeyRef.current = editionKey;
        inFlightKeyRef.current = null;
        clearEditionKeyParam();
      }
    })();

    return () => {
      active = false;
      if (inFlightKeyRef.current === editionKey) {
        inFlightKeyRef.current = null;
      }
    };
  }, [clearEditionKeyParam, editionKey, items, onOpen]);
}
