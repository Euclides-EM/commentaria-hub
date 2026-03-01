import { useCallback, useEffect, useMemo, useState } from "react";
import { Point } from "react-simple-maps";
import styled from "@emotion/styled";
import { isEmpty } from "lodash";
import { useQuery } from "@tanstack/react-query";
import { CityMarkers } from "../components/map/Markers";
import { MapControls } from "../components/map/MapControls";
import { useElementSize } from "../utils/useElementSize";
import { useLocalStorage, useWindowSize } from "usehooks-ts";
import {
  LAND_COLOR,
  PANE_BORDER,
  PANE_COLOR_ALT,
  SEA_COLOR,
} from "../utils/colors";
import { MdClose } from "react-icons/md";
import { TOOLTIP_CLOSE_CITY_DETAILS } from "../components/map/MapTooltips.tsx";
import {
  CitiesMap,
  DEFAULT_POSITION,
  MAX_ZOOM,
} from "../components/map/CitiesMap";
import { Timeline } from "../components/map/Timeline";
import { HeatLegend } from "../components/map/HeatMap";
import { useTour } from "@reactour/tour";
import { useEditionsSearch } from "../hooks/useEditionsSearch";
import { ItemModal } from "../components/tps/modal/ItemModal.tsx";
import { CityDetails } from "../components/map/CityDetails.tsx";
import { NAVBAR_HEIGHT } from "../components/layout/routes.ts";
import { useIsMobile } from "../components/layout/isMobile.ts";
import { useEditFilter } from "../contexts/FilterEditContext.tsx";
import { FLOATING_CITY_ENTRY } from "../types";
import type { Item } from "../types";
import { GeoDataService } from "@hub-api";
import { useAutoOpenEditionFromQuery } from "../hooks/useAutoOpenEditionFromQuery.ts";

const Wrapper = styled.div`
  position: fixed;
  display: flex;
  flex-direction: row;
  height: calc(100vh - ${NAVBAR_HEIGHT}px);
  width: 100%;
  color: black;
  overflow: hidden;
`;

const MapWrapper = styled.div`
  height: calc(100vh - ${NAVBAR_HEIGHT}px);
  width: 100%;

  svg {
    display: inline-block;
    vertical-align: middle;
    background-color: ${SEA_COLOR};
  }

  path {
    fill: ${LAND_COLOR};
  }
`;

const ControlsRow = styled.div`
  position: relative;
  right: 1rem;
  bottom: 4rem;
  display: flex;
  flex-direction: row;
  gap: 1rem;
  align-items: center;
  justify-content: flex-end;
  @media (max-width: 768px) {
    width: 100vw;
    gap: 8px;
  }
`;

const MapSection = styled.div`
  width: 100%;
  height: 100vh;
`;

const CollapseFiltersButton = styled.div`
  width: 1.5rem;
  height: 1.5rem;
  font-size: 1.5rem;
  cursor: pointer;
  border-radius: 20%;
  background-color: white;
  opacity: 0.7;
  color: ${SEA_COLOR};
  align-self: start;
  svg {
    margin-bottom: 2px;
  }
`;

const Pane = styled.div<{
  borderRight: boolean;
  widthPercentage?: number;
  backgroundColor?: string;
}>`
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: calc(100vh - 6rem + 4px);
  width: ${({ widthPercentage }) => widthPercentage || 20}%;
  min-width: 256px;
  overflow-x: auto;
  background-color: ${({ backgroundColor }) =>
    backgroundColor || PANE_COLOR_ALT};
  padding: 1rem;
  margin-bottom: 10rem;
  ${({ borderRight }) =>
    borderRight ? "border-right" : "border-left"}: 2px ${PANE_BORDER} solid;
`;

const loadCitiesAsync = async (): Promise<Record<string, Point>> => {
  const apiCities = await GeoDataService.getCities();
  const mappedCities = apiCities.reduce<Record<string, Point>>((acc, city) => {
    if (
      !city.name ||
      !Number.isFinite(city.longitude) ||
      !Number.isFinite(city.latitude)
    ) {
      return acc;
    }
    acc[city.name] = [city.longitude!, city.latitude!];
    return acc;
  }, {});

  return {
    ...mappedCities,
    [FLOATING_CITY_ENTRY.name]: [
      FLOATING_CITY_ENTRY.longitude,
      FLOATING_CITY_ENTRY.latitude,
    ],
  };
};

export const Map = () => {
  const { height } = useWindowSize();
  const isMobile = useIsMobile();
  const { setFilterOpen } = useEditFilter();
  const { items: filteredItems } = useEditionsSearch();
  const citiesQuery = useQuery({
    queryKey: ["cities"],
    queryFn: loadCitiesAsync,
  });
  const cities = citiesQuery.data ?? {};
  const [zoom, setZoom] = useLocalStorage<number>("zoom", 1);
  const {
    ref: mapSectionRef,
    size: { width: mapWidth },
    location: [mapX],
    refresh: refreshSize,
  } = useElementSize();
  const [position, setPosition] = useLocalStorage<Point>(
    "map-position",
    DEFAULT_POSITION,
  );
  const [selectedCity, setSelectedCity] = useState<string | undefined>();
  const [selectedRecordKey, setSelectedRecordId] = useState<
    string | undefined
  >();
  const [selectedItem, setSelectedItem] = useState<Item | undefined>();
  const [toured, setToured] = useLocalStorage<boolean>("map-toured", false);
  const { setIsOpen: setTourOpen } = useTour();

  useEffect(() => {
    setFilterOpen(true);
    return () => setFilterOpen(false);
  }, [setFilterOpen]);

  useEffect(() => {
    if (!toured) {
      setToured(true);
      setTourOpen(true);
    }
  }, [setTourOpen, setToured, toured]);

  const itemsByCity = useMemo(() => {
    const res = {} as Record<string, typeof filteredItems>;
    if (filteredItems) {
      filteredItems.forEach((item) => {
        item.cities.forEach((city) => {
          if (!res[city]) {
            res[city] = [];
          }
          res[city].push(item);
        });
      });
    }
    return res;
  }, [filteredItems]);

  const selectedRecord = useMemo(
    () =>
      selectedRecordKey && filteredItems
        ? filteredItems.filter((d) => d.key === selectedRecordKey)[0]
        : undefined,
    [filteredItems, selectedRecordKey],
  );
  const modalItem = selectedRecord || selectedItem;

  const openItemModal = useCallback((item: Item) => {
    setSelectedRecordId(item.key);
    setSelectedItem(item);
  }, []);

  useAutoOpenEditionFromQuery(filteredItems, openItemModal);

  useEffect(() => {
    refreshSize();
  }, [refreshSize, selectedCity, selectedRecordKey]);

  return (
    <Wrapper>
      <MapSection ref={mapSectionRef}>
        <MapWrapper>
          <CitiesMap
            height={height}
            width={mapWidth}
            zoom={zoom}
            position={position}
            setPosition={setPosition}
            setZoom={setZoom}
            markers={
              <CityMarkers
                cities={cities}
                data={itemsByCity as Record<string, Item[]>}
                selectedCity={selectedCity}
                setSelectedCity={(selected) =>
                  !isMobile && setSelectedCity(selected)
                }
              />
            }
          />
        </MapWrapper>
        <ControlsRow>
          <MapControls
            setZoom={setZoom}
            maxZoom={MAX_ZOOM}
            resetCenter={() => {
              setPosition(DEFAULT_POSITION);
              setZoom(1);
            }}
            openTour={() => setTourOpen(true)}
          />
          <Timeline />
        </ControlsRow>
        <HeatLegend
          offsetRight={mapX + mapWidth}
          total={filteredItems?.length || 0}
        />
      </MapSection>
      {!isEmpty(selectedCity) && (
        <Pane borderRight={false}>
          <CollapseFiltersButton
            onClick={() => setSelectedCity(undefined)}
            data-tooltip-id={TOOLTIP_CLOSE_CITY_DETAILS}
            data-tooltip-content="Close"
          >
            <MdClose />
          </CollapseFiltersButton>
          <CityDetails
            city={selectedCity!}
            data={itemsByCity[selectedCity!] || []}
            selectedRecordId={selectedRecordKey}
            setSelectedRecordKey={(key) => {
              setSelectedItem(undefined);
              setSelectedRecordId(key);
            }}
          />
        </Pane>
      )}
      {!isEmpty(modalItem) && (
        <ItemModal
          item={modalItem!}
          featuresById={{}}
          onClose={() => {
            setSelectedRecordId(undefined);
            setSelectedItem(undefined);
          }}
        />
      )}
    </Wrapper>
  );
};
