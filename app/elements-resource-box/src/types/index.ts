import { NO_CITY } from "../constants";
import {
  feature_Feature,
  model_City,
  type model_EditionTitlePageStatus,
} from "@hub-api";

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
  editors: string[];
  publishers: string[];
  tpImageName: string | null;
  titlePageStatus: model_EditionTitlePageStatus;
  shortTitle: string | null;
  title: string | null;
  titleEn: string | null;
  imprint: string | null;
  imprintEn: string | null;
  scanUrl: string[];
  type: string;
  format: number | null;
  elementsBooks: Range[];
  elementsBooksExpanded: number[];
  additionalContent: string[];
  volumesCount: number | null;
  class: string | null;
  notes: string | null;
  study_corpora: string[];
  diagramCropsAvailable: boolean | null;
  hasDiagrams: boolean | undefined;
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

export const STUDY_CORPUSES: Record<string, string> = {
  dh: "DH Core",
  dotted_lines: "Dotted Lines",
  Angela_metadata: "Angela Metadata",
  origin_eip_csv: "Original EiP",
  axiomatics_16: "Axiomatics 16",
  tps_experiment: "Title Pages Experiment",
  tps_experiment_reviewed: "Title Pages Experiment [Reviewed]",
};
