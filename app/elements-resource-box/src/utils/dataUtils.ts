import { Item } from "../types";
import { startCase, uniq } from "lodash";
import { ItemTypes } from "../constants";
import type { model_Edition } from "@hub-api";
import { toItemImageUrl } from "./util.ts";

type EditionWithVisualElementsTypes = model_Edition & {
  visualElementsTypes?: string[];
};

const firstOrNull = <T>(arr: T[]): T | null => (arr.length > 0 ? arr[0] : null);

const toBookRanges = (books: number[]) => {
  return Array.from(new Set(books))
    .sort((a, b) => a - b)
    .reduce<{ start: number; end: number }[]>((ranges, book) => {
      const previous = ranges[ranges.length - 1];
      if (!previous || book > previous.end + 1) {
        ranges.push({ start: book, end: book });
        return ranges;
      }
      previous.end = book;
      return ranges;
    }, []);
};

export const mapEditionsToItems = (editions: model_Edition[]): Item[] => {
  return editions
    .filter((edition) => edition.key)
    .map((edition) => {
      const editionWithVisualElementsTypes =
        edition as EditionWithVisualElementsTypes;
      const shelfmarks = edition.shelfmarks || [];
      const books = Array.isArray(edition.books)
        ? edition.books.filter((value): value is number =>
            Number.isFinite(value),
          )
        : [];
      return {
        key: edition.key!,
        year: edition.year || null,
        yearFrom: edition.manuscriptYearFrom ?? null,
        yearTo: edition.manuscriptYearTo ?? null,
        yearIsApproximate: Boolean(edition.manuscriptYearIsApproximate),
        materialType: edition.isManuscript ? "Manuscript" : "Print",
        cities: edition.cities || [],
        languages: (edition.languages || [])
          .map((lang) => startCase(lang.trim().toLowerCase()))
          .filter(Boolean),
        editors: (edition.editor || [])
          .map((name) => name.trim())
          .filter(Boolean),
        publishers: (edition.publisher || [])
          .map((name) => name.trim())
          .filter(Boolean),
        repository:
          firstOrNull(
            shelfmarks
              .map((s) => s.repository?.trim())
              .filter(Boolean) as string[],
          ) || null,
        tpImageName:
          firstOrNull(
            shelfmarks.map((s) => s.title_page_img).filter(Boolean) as string[],
          ) ||
          firstOrNull(
            shelfmarks
              .map((s) => s.frontispiece_img)
              .filter(Boolean) as string[],
          ),
        shortTitle: edition.shortTitle || null,
        title: edition.title || null,
        titleEn: edition.title_EN || null,
        imprint: edition.imprint || null,
        imprintEn: edition.imprint_EN || null,
        facsimiles: shelfmarks.filter((s) => s.scan),
        type: edition.isElements ? ItemTypes.elements : ItemTypes.secondary,
        format: edition.format || null,
        elementsBooks: toBookRanges(books),
        elementsBooksExpanded: books,
        additionalContent: edition.additionalContent || [],
        volumesCount: edition.volumes ?? null,
        titlePageStatus: edition.titlePageStatus || "Unknown",
        study_corpora: edition.corpus || [],
        notes: edition.notes || null,
        diagramCropsAvailable: edition.diagramCropsAvailable || null,
        hasDiagrams: edition.hasDiagrams,
        visualElementsTypes: uniq(
          (editionWithVisualElementsTypes.visualElementsTypes ||
            (edition.visualElements || [])
              .map((v) => v.visual_element_type)
              .filter(Boolean)) as string[],
        ),
        reprintOf: edition.reprintOf || null,
      } satisfies Item;
    });
};

export const personDisplayName = (person: string) => {
  person = person.replace("(?)", "").replace("?", "").trim();
  const parts = person.split(/\s+/).filter(Boolean);
  if (parts.length === 1) {
    return person;
  }

  const separators = [
    "de",
    "la",
    "del",
    "della",
    "di",
    "da",
    "do",
    "dos",
    "das",
    "du",
    "van",
    "von",
    "der",
    "den",
    "ter",
    "ten",
    "op",
    "af",
    "al",
    "le",
    "el",
    "of",
    "lefèvre",
  ];
  const lowerParts = parts.map((p) => p.toLowerCase());

  let sepIndex = -1;
  for (let i = 1; i < lowerParts.length; i++) {
    if (separators.includes(lowerParts[i])) {
      sepIndex = i;
      break;
    }
  }

  if (sepIndex !== -1) {
    const lastName = parts.slice(sepIndex).join(" ").trim();
    const firstNames = parts.slice(0, sepIndex).join(" ").trim();
    return `${lastName}, ${firstNames}`;
  } else {
    const lastName = parts[parts.length - 1];
    const firstNames = parts.slice(0, -1).join(" ").trim();
    return `${lastName}, ${firstNames}`;
  }
};

export function openScan(item: Item) {
  if (item.facsimiles.length === 0) {
    return;
  }
  return window.open(item.facsimiles[0].scan!, "_blank")?.focus();
}

export function openImage(item: Item) {
  const imageUrl = toItemImageUrl(item.tpImageName);
  if (!imageUrl) {
    return;
  }
  return window.open(imageUrl, "_blank")?.focus();
}
