import {
  ComposableMap,
  Geographies,
  Geography,
  Point,
  ZoomableGroup,
} from "react-simple-maps";
import { ReactElement } from "react";

const geoUrl = "https://cdn.jsdelivr.net/npm/world-atlas@2/land-110m.json";

type MapProps = {
  height: number;
  width: number;
  zoom: number;
  position: Point;
  setPosition: React.Dispatch<React.SetStateAction<Point>>;
  setZoom?: React.Dispatch<React.SetStateAction<number>>;
  markers: ReactElement;
};

export const MAX_ZOOM = 10;
export const MIN_ZOOM = 1;
export const DEFAULT_POSITION: Point = [-2, 45];

export const CitiesMap = ({
  height,
  width,
  zoom,
  position,
  setPosition,
  setZoom,
  markers,
}: MapProps) => (
  <ComposableMap
    height={height}
    width={width}
    projection="geoAzimuthalEqualArea"
    projectionConfig={{
      rotate: [-40.0, -48.0, 0],
      scale: 2200,
    }}
  >
    <ZoomableGroup
      zoom={zoom}
      minZoom={MIN_ZOOM}
      maxZoom={MAX_ZOOM}
      center={position}
      onMoveEnd={(e) => {
        setPosition(e.coordinates);
        if (setZoom) {
          const clampedZoom = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, e.zoom));
          if (clampedZoom !== zoom) {
            setZoom(clampedZoom);
          }
        }
      }}
    >
      <Geographies geography={geoUrl}>
        {({ geographies }) =>
          geographies.map((geo) => (
            <Geography key={geo.rsmKey} geography={geo} />
          ))
        }
      </Geographies>
      {markers}
    </ZoomableGroup>
  </ComposableMap>
);
