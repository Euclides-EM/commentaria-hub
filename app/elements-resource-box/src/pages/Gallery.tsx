import { useCallback, useEffect, useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import useLocalStorageState from "use-local-storage-state";
import { Mode } from "../types";
import {
  Column,
  Container,
  ResetButton,
  Row,
  ScrollToTopButton,
  Text,
} from "../components/common";
import { TILE_HEIGHT, TILE_WIDTH, TITLE_PAGES_DATASET_ID } from "../constants";
import { IoWarning } from "react-icons/io5";
import styled from "@emotion/styled";
import { LAND_COLOR } from "../utils/colors.ts";
import { Stats } from "../components/Stats.tsx";
import { MAIN_CONTENT_ID } from "../components/layout/routes.ts";
import { groupByMap } from "../utils/util.ts";
import { useEditionsSearchInfinite } from "../hooks/useEditionsSearch.ts";
import { feature_Feature } from "@hub-api";
import { FeaturesService } from "@hub-api";
import { MultiSelect } from "../components/tps/filters/MultiSelect.tsx";
import { ItemView } from "../components/tps/features/ItemView.tsx";

const NoteLine = styled(Row)`
  opacity: 0.8;
`;

export function Gallery({ titlePagesModeOn }: { titlePagesModeOn: boolean }) {
  const { items, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } =
    useEditionsSearchInfinite({
      pageSize: 25,
      orderBy: [
        { field: "year", descending: false },
        { field: "key", descending: false },
      ],
    });

  const [selectedFeatureIds, setSelectedFeatureIds] = useLocalStorageState<
    string[]
  >("tp-features", {
    defaultValue: [],
  });
  const [showScrollTop, setShowScrollTop] = useState(false);
  const featuresQuery = useQuery({
    queryKey: ["title-pages", "features", TITLE_PAGES_DATASET_ID],
    queryFn: () =>
      FeaturesService.getDatasetsFeatures({
        dataSetId: TITLE_PAGES_DATASET_ID,
      }),
    enabled: titlePagesModeOn,
  });

  const {
    availableFeatures,
    featureColors,
    featureTooltips,
    defaultFeatureIds,
  } = useMemo(() => {
    const response = featuresQuery.data ?? [];
    const sortedFeatures = response
      .filter((feature) => feature.id && feature.name)
      .slice()
      .sort((a, b) =>
        (a.name || "").localeCompare(b.name || "", undefined, {
          sensitivity: "base",
        }),
      );
    const colorMap: Record<string, string> = {};
    const tooltipMap: Record<string, string> = {};
    response.forEach((feature: feature_Feature) => {
      if (!feature.id) {
        return;
      }
      if (feature.color) {
        colorMap[feature.id] = feature.color;
      }
      if (feature.description) {
        tooltipMap[feature.id] = feature.description;
      }
    });
    return {
      availableFeatures: sortedFeatures,
      featureColors: colorMap,
      featureTooltips: tooltipMap,
      defaultFeatureIds: sortedFeatures
        .filter((feature) => feature.is_default)
        .map((feature) => feature.id as string),
    };
  }, [featuresQuery.data]);

  const featureNameById = useMemo(() => {
    const map: Record<string, string> = {};
    availableFeatures.forEach((feature) => {
      if (feature.id && feature.name) {
        map[feature.id] = feature.name;
      }
    });
    return map;
  }, [availableFeatures]);

  const sortedFeatureIds = useMemo(() => {
    return availableFeatures.map((feature) => feature.id as string);
  }, [availableFeatures]);

  const sortFeatures = useCallback(
    (ids: string[]) =>
      [...ids].sort((a, b) =>
        (featureNameById[a] || "").localeCompare(
          featureNameById[b] || "",
          undefined,
          { sensitivity: "base" },
        ),
      ),
    [featureNameById],
  );
  const mode: Mode = titlePagesModeOn ? "texts" : "images";

  useEffect(() => {
    if (availableFeatures.length === 0) {
      return;
    }
    setSelectedFeatureIds((prev) => {
      const normalized = prev
        .map((value) => {
          if (availableFeatures.some((f) => f.id === value)) {
            return value;
          }
          const matching = availableFeatures.filter(
            (feature) => feature.name === value,
          );
          if (matching.length === 1) {
            return matching[0].id as string;
          }
          return null;
        })
        .filter((value): value is string => Boolean(value));
      if (normalized.length > 0) {
        return sortFeatures(normalized);
      }
      if (defaultFeatureIds.length > 0) {
        return sortFeatures(defaultFeatureIds);
      }
      return sortFeatures(
        availableFeatures.map((feature) => feature.id as string),
      );
    });
  }, [
    availableFeatures,
    defaultFeatureIds,
    setSelectedFeatureIds,
    sortFeatures,
  ]);

  const handleScroll = useCallback(() => {
    const el = document.getElementById(MAIN_CONTENT_ID);
    if (!el) {
      return;
    }
    const scrollTop = el.scrollTop;
    setShowScrollTop(scrollTop > 200);

    const remaining = el.scrollHeight - (el.scrollTop + el.clientHeight);
    if (remaining < 300 && hasNextPage && !isFetchingNextPage) {
      fetchNextPage();
    }
  }, [fetchNextPage, hasNextPage, isFetchingNextPage]);

  const scrollToTop = () => {
    document.getElementById(MAIN_CONTENT_ID)?.scrollTo({
      top: 0,
      behavior: "smooth",
    });
  };

  const filteredBySearchItems = items;

  useEffect(() => {
    const el = document.getElementById(MAIN_CONTENT_ID);
    el?.addEventListener("scroll", handleScroll);
    handleScroll();
    return () => {
      el?.removeEventListener("scroll", handleScroll);
    };
  }, [handleScroll]);

  useEffect(() => {
    const el = document.getElementById(MAIN_CONTENT_ID);
    if (!el || !hasNextPage || isFetchingNextPage) {
      return;
    }
    if (el.scrollHeight - el.clientHeight < 300) {
      fetchNextPage();
    }
  }, [fetchNextPage, items?.length, hasNextPage, isFetchingNextPage]);

  return (
    <Container
      style={{
        position: "relative",
        margin: "2rem 0",
        minHeight: "calc(100vh - 6rem)",
      }}
    >
      {showScrollTop && (
        <ScrollToTopButton onClick={scrollToTop} title="Scroll to top">
          ↑
        </ScrollToTopButton>
      )}
      <Column minWidth="min(820px, 90%)">
        <Stats />
        {titlePagesModeOn && (
          <>
            <Row justifyStart noWrap>
              <Column alignItems="end">
                <span>Highlight Segments:</span>
              </Column>
              <MultiSelect
                name="Features"
                value={selectedFeatureIds}
                options={sortedFeatureIds}
                labelFn={(featureId) => featureNameById[featureId] || featureId}
                onChange={(f) =>
                  setSelectedFeatureIds(sortFeatures(f as string[]))
                }
                colors={featureColors}
                tooltips={featureTooltips}
                className="features-multi-select"
              />
              <ResetButton
                onClick={() =>
                  setSelectedFeatureIds(
                    sortFeatures(
                      defaultFeatureIds.length > 0
                        ? defaultFeatureIds
                        : sortedFeatureIds,
                    ),
                  )
                }
              >
                Reset
              </ResetButton>
            </Row>
            <NoteLine gap={0.5} noWrap noWrapAlsoOnMobile>
              <IoWarning /> Highlighted features were partially identified using
              an LLM and may not be accurate.
            </NoteLine>
          </>
        )}
      </Column>
      <Row rowGap={6}>
        {filteredBySearchItems == null ||
        (filteredBySearchItems?.length || 0) > 0 ? (
          filteredBySearchItems
            ?.sort((a, b) => {
              if (!a.year) return 1;
              if (!b.year) return -1;
              return a.year.localeCompare(b.year);
            })
            .map((item) => (
              <ItemView
                key={item.key}
                height={TILE_HEIGHT}
                width={TILE_WIDTH}
                item={item}
                mode={mode}
                featuresById={
                  titlePagesModeOn
                    ? groupByMap(
                        availableFeatures.filter((feat) =>
                          selectedFeatureIds.includes(feat.id!),
                        ),
                        (feat) => feat.id!,
                      )
                    : null
                }
              />
            ))
        ) : (
          <Text size={1.5} color={LAND_COLOR}>
            No matches. Try adjusting the filters or search.
          </Text>
        )}
      </Row>
      <Text size={1}>
        {isLoading
          ? "Loading editions..."
          : isFetchingNextPage
            ? "Loading more editions..."
            : hasNextPage
              ? "Scroll to load more editions."
              : "All matching editions loaded."}
      </Text>
      <Text size={1} style={{ marginTop: "auto" }}>
        À la Croisée des Hyperliens, chez le scribe fatigué et son félin
        passivement investi, MMXXV.
      </Text>
      <div />
    </Container>
  );
}
