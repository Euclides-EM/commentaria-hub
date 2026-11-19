import { reduce } from "lodash";
import { Range } from "../types";
import { TITLE_PAGES_DATASET_ID } from "../constants";

export const groupByMap = <TFrom, TKey extends string, TValue = TFrom>(
  data: TFrom[],
  keyBy: (entry: TFrom) => TKey,
  valueBy: ((entry: TFrom) => TValue) | undefined = undefined,
): Record<TKey, TValue> => {
  return reduce(
    data,
    (acc, entry) => ({
      ...acc,
      [keyBy(entry)]: valueBy ? valueBy(entry) : entry,
    }),
    {} as Record<TKey, TValue>,
  );
};

export const joinArr = (arr: string[]): string => {
  if (arr.length === 0) return "";
  if (arr.length === 1) return arr[0];
  if (arr.length === 2) return `${arr[0]} and ${arr[1]}`;
  return `${arr.slice(0, -1).join(", ")}, and ${arr[arr.length - 1]}`;
};

export const formatBookRanges = (ranges: Range[]): string => {
  if (!ranges.length) {
    return "";
  }

  const merged = ranges
    .map((range) => ({
      start: Math.min(range.start, range.end),
      end: Math.max(range.start, range.end),
    }))
    .sort((a, b) => a.start - b.start)
    .reduce<Range[]>((acc, range) => {
      const previous = acc[acc.length - 1];
      if (!previous || range.start > previous.end + 1) {
        acc.push({ ...range });
        return acc;
      }
      previous.end = Math.max(previous.end, range.end);
      return acc;
    }, []);

  return merged
    .map((range) =>
      range.start === range.end
        ? `${range.start}`
        : `${range.start}-${range.end}`,
    )
    .join(", ");
};

const bookFormatNames: Record<string, string> = {
  "2": "Folio",
  "4": "Quarto",
  "6": "Sexto",
  "8": "Octavo",
  "12": "Duodecimo",
  "16": "Sextodecimo",
  "18": "Octodecimo",
};

export const formatBookFormat = (value: string | number | null | undefined) => {
  if (value === null || value === undefined || value === "") {
    return "";
  }

  const format = String(value);
  if (!/^\d+$/.test(format)) {
    return format;
  }
  const name = bookFormatNames[format];
  return name ? `${format}º (${name})` : `${format}º`;
};

export function isValidUrl(s: string) {
  try {
    new URL(s);
  } catch {
    return false;
  }
  return true;
}

export const toItemImageUrl = (imageName: string | null): string | null => {
  if (!imageName) {
    return null;
  }
  if (isValidUrl(imageName)) {
    return imageName;
  }
  return `${import.meta.env.VITE_BACKEND_URL}/store/data/${TITLE_PAGES_DATASET_ID}/imgs/${imageName}`;
};
