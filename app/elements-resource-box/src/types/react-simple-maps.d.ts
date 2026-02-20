import {
  ComposableMapProps,
  ZoomableGroupProps,
  GeographiesProps,
  GeographyProps,
  MarkerProps,
} from "react-simple-maps";

declare module "react-simple-maps" {
  export const ComposableMap: React.FC<ComposableMapProps>;
  export const ZoomableGroup: React.FC<ZoomableGroupProps>;
  export const Geographies: React.FC<GeographiesProps>;
  export const Geography: React.FC<GeographyProps>;
  export const Marker: React.FC<MarkerProps>;
}
