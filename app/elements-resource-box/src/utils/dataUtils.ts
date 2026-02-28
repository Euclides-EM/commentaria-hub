import { Item } from "../types";
import { startCase, uniq } from "lodash";
import { ItemTypes } from "../constants";
import type { model_Edition } from "@hub-api";
import { toItemImageUrl } from "./util.ts";

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
      const shelfmarks = edition.shelfmarks || [];
      const books = (edition.books || []).filter((value): value is number =>
        Number.isFinite(value),
      );
      return {
        key: edition.key!,
        year: edition.year || null,
        cities: edition.cities || [],
        languages: (edition.languages || [])
          .map((lang) => startCase(lang.trim().toLowerCase()))
          .filter(Boolean),
        authors: (edition.editor || [])
          .map((name) => name.trim())
          .filter(Boolean),
        publishers: (edition.publisher || [])
          .map((name) => name.trim())
          .filter(Boolean),
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
        scanUrl: shelfmarks
          .map((s) => s.scan?.trim())
          .filter(Boolean) as string[],
        type: edition.isElements ? ItemTypes.elements : ItemTypes.secondary,
        format: edition.format || -1,
        elementsBooks: toBookRanges(books),
        elementsBooksExpanded: books,
        additionalContent: edition.additionalContent || [],
        volumesCount: edition.volumes ?? null,
        class: edition.manuscriptClass || null,
        hasTitle: "TODO",
        study_corpora: edition.corpus || [],
        notes: edition.notes || null,
        diagramCropsAvailable: edition.diagramCropsAvailable || null,
        hasDiagrams: edition.hasDiagrams,
        visualElementsTypes: uniq(
          (edition.visualElements || [])
            .map((v) => v.visual_element_type)
            .filter(Boolean) as string[],
        ),
        reprintOf: edition.reprintOf || null,
      } satisfies Item;
    })
    .sort(
      (a, b) =>
        (a.year || "").localeCompare(b.year || "") ||
        a.key.localeCompare(b.key),
    );
};

export const authorDisplayName = (author: string) => {
  author = author.replace("(?)", "").replace("?", "").trim();
  const parts = author.split(/\s+/).filter(Boolean);
  if (parts.length === 1) {
    return author;
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
  if (!item.scanUrl || item.scanUrl.length === 0) {
    return;
  }
  return window.open(item.scanUrl[0], "_blank")?.focus();
}

export function openImage(item: Item) {
  const imageUrl = toItemImageUrl(item.tpImageName);
  if (!imageUrl) {
    return;
  }
  return window.open(imageUrl, "_blank")?.focus();
}
