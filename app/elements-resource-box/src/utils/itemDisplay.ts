import { NO_EDITOR, NO_YEAR } from "../constants";
import type { Item } from "../types";
import { joinArr } from "./util";

export const isManuscriptItem = (item: Item) =>
  item.materialType === "Manuscript";

export const formatDisplayYear = (item: Item) => {
  if (isManuscriptItem(item) && item.yearFrom) {
    const prefix = item.yearIsApproximate ? "~" : "";
    if (item.yearTo && item.yearTo !== item.yearFrom) {
      return `${prefix}${item.yearFrom}-${item.yearTo}`;
    }
    return `${prefix}${item.yearFrom}`;
  }
  return item.year || NO_YEAR;
};

export const formatDisplayEditors = (item: Item) =>
  joinArr(item.editors) || NO_EDITOR;

export const formatDisplayBooks = (item: Item) => {
  if (isManuscriptItem(item)) {
    return item.elementsBooksRaw || "";
  }
  return item.elementsBooks
    .map((range) =>
      range.start === range.end
        ? `${range.start}`
        : `${range.start}-${range.end}`,
    )
    .join(", ");
};
