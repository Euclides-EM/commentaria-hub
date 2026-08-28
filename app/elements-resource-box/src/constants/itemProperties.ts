import { FilterGroup, Range } from "../types";
import { NO_CITY } from "./index.ts";

export type ItemProperty = {
  displayName: string;
  filterGroup?: FilterGroup | "General";
  isArray?: boolean;
  filterOptions?: { label: string; value: string }[];
  customCompareFn?: (a: unknown, b: unknown) => number;
  isTitlePageImageFeature?: boolean;
  isTitlePageTextFeature?: boolean;
  groupByJoinArray?: boolean;
  notFilterable?: boolean;
  notGroupable?: boolean;
};

function parseRangeIfNeeded(a: Range | string): Range {
  if (!(typeof a === "string")) {
    return a;
  }
  if (a === "None") {
    return { start: 100, end: 100 };
  }
  const parts = (a as string).split("-");
  return {
    start: parseInt(parts[0]),
    end: parseInt(parts[parts.length - 1]),
  };
}

export const itemProperties: {
  [key: string]: ItemProperty;
} = {
  study_corpora: {
    displayName: "Study Corpus",
    isArray: true,
  },
  type: {
    displayName: "Book Classification",
  },
  languages: {
    displayName: "Languages",
    isArray: true,
    groupByJoinArray: true,
  },
  cities: {
    displayName: "Cities",
    isArray: true,
    customCompareFn: ((a: string, b: string) => {
      if (a === NO_CITY) {
        return -1;
      }
      if (b === NO_CITY) {
        return 1;
      }
      return a.localeCompare(b, undefined, { sensitivity: "base" });
    }) as (a: unknown, b: unknown) => number,
    groupByJoinArray: true,
  },
  editors: {
    displayName: "Editors",
    isArray: true,
    groupByJoinArray: true,
  },
  publishers: {
    displayName: "Publishers",
    isArray: true,
    groupByJoinArray: true,
  },
  elementsBooks: {
    displayName: "Elements Books (ranges)",
    filterGroup: "Elements",
    isArray: true,
    notFilterable: true,
    customCompareFn: ((a: Range | string, b: Range | string) => {
      const rangeA: Range = parseRangeIfNeeded(a);
      const rangeB: Range = parseRangeIfNeeded(b);
      if (rangeA.start === rangeB.start) {
        return rangeA.end - rangeB.end;
      }
      return rangeA.start - rangeB.start;
    }) as (a: unknown, b: unknown) => number,
  },
  elementsBooksExpanded: {
    displayName: "Elements Books",
    filterGroup: "Elements",
    isArray: true,
    customCompareFn: ((a: string, b: string): number => {
      if (a === "None") return 1;
      if (b === "None") return -1;
      const numA = parseInt(a);
      const numB = parseInt(b);
      return numA - numB;
    }) as (a: unknown, b: unknown) => number,
  },
  additionalContent: {
    displayName: "Additional Content",
    filterGroup: "Elements",
    isArray: true,
  },
  diagramCropsAvailable: {
    displayName: "Diagrams Extracted",
    filterGroup: "Diagrams",
  },
  hasDiagrams: {
    displayName: "Has Diagrams",
    filterGroup: "Diagrams",
  },
  visualElementsTypes: {
    displayName: "Visual Elements Types",
    isArray: true,
    filterGroup: "Diagrams",
  },
  volumesCount: {
    displayName: "Number of Volumes",
    filterGroup: "Material",
  },
  format: {
    displayName: "Format",
    filterGroup: "Material",
    customCompareFn: ((a: string, b: string): number => {
      const numA = parseInt(a);
      const numB = parseInt(b);
      if (Number.isNaN(numA)) return Number.isNaN(numB) ? 0 : 1;
      if (Number.isNaN(numB)) return -1;
      return numA - numB;
    }) as (a: unknown, b: unknown) => number,
  },
  shelfmarkProperties: {
    displayName: "Shelfmark Properties",
    filterGroup: "Material",
    isArray: true,
    filterOptions: [
      { label: "Shelfmark available", value: "shelfmark_available" },
      { label: "Facsimile available", value: "facsimile_available" },
      {
        label: "Copyright status unknown",
        value: "copyright_status_unknown",
      },
      {
        label: "External transcription available",
        value: "external_transcription_available",
      },
      {
        label: "Internal transcription available",
        value: "internal_transcription_available",
      },
      {
        label: "External structural metadata available",
        value: "external_structural_metadata_available",
      },
      {
        label: "Internal structural metadata available",
        value: "internal_structural_metadata_available",
      },
    ],
  },
  titlePageStatus: {
    displayName: "Has Title Page",
    filterGroup: "Title Page",
  },
  subjectCategoryValues: {
    displayName: "Subject Categories",
    filterGroup: "Subject Categories",
    isArray: true,
  },
};
