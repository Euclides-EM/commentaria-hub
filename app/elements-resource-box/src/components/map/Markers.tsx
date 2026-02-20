import { useCallback, useEffect, useRef, useState } from "react";
import { Marker, Point } from "react-simple-maps";
import {
  MARKER_1,
  MARKER_5,
  MARKER_STROKE,
  MARKER_STROKE_HL,
  SEA_COLOR,
} from "../../utils/colors";
import styled from "@emotion/styled";
import { getHeatColor } from "./HeatMap";
import { CityName } from "./CityDetails";
import { CITY_MARKER_ID } from "./Tour";
import { TOOLTIP_MARKER_ARROW } from "./MapTooltips.tsx";
import { Item } from "../../types";

type CityMarkersProps = {
  cities: Record<string, Point>;
  data: Record<string, Item[]>;
  selectedCity: string | undefined;
  setSelectedCity: React.Dispatch<React.SetStateAction<string | undefined>>;
};

const StyledCircle = styled.circle<{ selected: boolean; hovered: boolean }>`
  stroke: ${({ selected, hovered }) =>
    selected || hovered ? MARKER_STROKE_HL : MARKER_STROKE};
  stroke-width: 2px;
  cursor: pointer;
`;

const StyledText = styled.text`
  stroke: ${MARKER_5};
  fill: ${MARKER_1};
  font-size: 1.2rem;
  stroke-width: 4px;
  font-weight: bold;
  paint-order: stroke;
  stroke-linejoin: round;
  cursor: pointer;
`;

const ArrowStyle = styled.text`
  font-size: 2rem;
  font-weight: bolder;
  cursor: default;
  fill: ${SEA_COLOR};
`;

const Arrow = () => (
  <ArrowStyle id={TOOLTIP_MARKER_ARROW} x={10} y={5}>
    ⟶
  </ArrowStyle>
);

const OptionalArrow = ({ city }: { city: string }) => {
  if (city === "Beijing") {
    return <Arrow />;
  }
  return undefined;
};

export const CityMarkers = ({
  cities,
  data,
  selectedCity,
  setSelectedCity,
}: CityMarkersProps) => {
  const refsByCity = useRef<Record<string, SVGCircleElement | null>>({});
  const [hoveredCity, setHoveredCity] = useState<string | undefined>();

  const setRef = useCallback(
    (element: SVGCircleElement | null, key: string) => {
      refsByCity.current[key] = element;
    },
    [],
  );

  const handleCircleHover = useCallback((event: Event) => {
    const circle = event.target as SVGCircleElement;
    const city = circle.id;
    setHoveredCity(city);
  }, []);

  const handleCircleHoverStop = useCallback(
    () => setHoveredCity(undefined),
    [],
  );

  const getFillColor = useCallback((value: number) => getHeatColor(value), []);

  const reffedCities = Object.keys(refsByCity);
  useEffect(() => {
    const circles = Object.values(refsByCity.current).filter((c) => c);
    circles.forEach((circle) => {
      circle!.addEventListener("mouseenter", handleCircleHover);
      circle!.addEventListener("mouseleave", handleCircleHoverStop);
    });
    return () => {
      circles.forEach((circle) => {
        circle!.removeEventListener("mouseenter", handleCircleHover);
        circle!.removeEventListener("mouseleave", handleCircleHoverStop);
      });
    };
  }, [reffedCities, handleCircleHover, handleCircleHoverStop]);

  return (
    <>
      {Object.keys(cities)
        .filter((city) => data[city]?.length > 0)
        .map((city) => (
          <Marker
            key={`${city}_circle`}
            id={`${CITY_MARKER_ID}${city}`}
            coordinates={cities[city]}
          >
            <StyledCircle
              fill={getFillColor(data[city]?.length || 0)}
              selected={city === hoveredCity}
              hovered={city === selectedCity}
              onClick={() => setSelectedCity(city)}
              id={city}
              ref={(element) => setRef(element, city)}
              r={8}
            />
            <OptionalArrow city={city} />
          </Marker>
        ))}
      {hoveredCity && (
        <Marker key={`${hoveredCity}_hover`} coordinates={cities[hoveredCity]}>
          <StyledText x={-8} y={-12}>
            {CityName(hoveredCity)} · {data[hoveredCity]?.length || 0}
          </StyledText>
        </Marker>
      )}
    </>
  );
};
