import { NO_CITY } from "../constants";
import { feature_Feature, model_City } from '@hub-api';

export type Mode = "texts" | "images";

export type FilterGroup =
  | "General"
  | "Elements"
  | "Title Page"
  | "Material"
  | "Diagrams";

export type Range = {
  start: number;
  end: number;
};

type YesNoBool = "Yes" | "No";

export const MIN_YEAR = 1482;
export const MAX_YEAR = 1883;

export const FLOATING_CITY_ENTRY: Required<model_City> = {
  name: NO_CITY,
  longitude: -16,
  latitude: 42,
};

export type Item = {
  key: string;
  reprintOf?: string | null;
  year: string | null;
  cities: string[];
  languages: string[];
  authors: string[];
  publishers: string[];
  tpImageName: string | null;
  hasTitle: string;
  shortTitle: string | null;
  title: string | null;
  titleEn: string | null;
  imprint: string | null;
  imprintEn: string | null;
  scanUrl: string[];
  type: string;
  format: string | null;
  elementsBooks: Range[];
  elementsBooksExpanded: number[];
  additionalContent: string[];
  volumesCount: number | null;
  class: string | null;
  notes: string | null;
  study_corpora: string[];
  diagramsExtracted: YesNoBool | null;
  hasDiagrams: string;
  visualElementsTypes: string[];
};

export type RadioProps = {
  name: string;
  options: [string, string];
  value: boolean;
  onChange: (value: boolean) => void;
};

export type ItemProps = {
  item: Item;
  height: number;
  width: number;
  mode: Mode;
  featuresById: Record<string, feature_Feature> | null;
};
