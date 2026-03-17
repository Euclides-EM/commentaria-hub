import type { search_Query } from "@hub-api";
import { inEuclidesMode } from "./mode";

export const injectEuclidesEditionConstraints = (
  query: Omit<search_Query, "offset" | "limit"> = {},
): Omit<search_Query, "offset" | "limit"> => {
  if (!inEuclidesMode()) {
    return query;
  }

  return {
    ...query,
    fields_filter: {
      ...query.fields_filter,
      isElements: ["true"],
    },
    filter_includes: {
      ...query.filter_includes,
      isElements: true,
    },
  };
};
