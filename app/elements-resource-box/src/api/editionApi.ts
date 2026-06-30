import type { model_Edition, model_USTC, search_Query } from "@hub-api";
import { EditionsService, OpenAPI, ThirdPartyCatalogsService } from "@hub-api";
import { uploadImage } from "./imageApi.ts";

export const upsertEdition = async (
  data: model_Edition,
  images: Record<string, File>,
  options?: { isNew?: boolean },
): Promise<void> => {
  console.log("Upserting edition:", data);

  const uploads: Promise<void>[] = [];

  for (let i = 0; i < (data.shelfmarks?.length ?? 0); i++) {
    const shelfmark = data.shelfmarks![i];
    if (shelfmark.title_page_img) {
      const file = images[shelfmark.title_page_img];
      if (!file) {
        continue;
      }
      uploads.push(
        (async () => {
          shelfmark.title_page_img = await uploadImage(data.key!, file, "tp");
        })(),
      );
    }
    if (shelfmark.frontispiece_img) {
      const file = images[shelfmark.frontispiece_img];
      if (!file) {
        continue;
      }
      uploads.push(
        (async () => {
          shelfmark.frontispiece_img = await uploadImage(
            data.key!,
            file,
            "frontispiece",
          );
        })(),
      );
    }
  }

  for (let i = 0; i < (data.visualElements?.length ?? 0); i++) {
    const visualElement = data.visualElements![i];
    for (let j = 0; j < (visualElement.examples?.length ?? 0); j++) {
      const example = visualElement.examples![j];
      if (example.img) {
        const file = images[example.img];
        if (!file) {
          continue;
        }
        uploads.push(
          (async () => {
            example.img = await uploadImage(data.key!, file, "facsimile");
          })(),
        );
      }
    }
  }

  await Promise.all(uploads);

  if (options?.isNew) {
    await EditionsService.postEditions({ edition: data });
  } else {
    await EditionsService.putEditions({ editionId: data.key!, edition: data });
  }
};

export const deleteEdition = async (editionId: string): Promise<void> => {
  await EditionsService.deleteEditions({ editionId });
};

export const ustcLookup = async (
  ustcId: string,
): Promise<Partial<model_USTC>> => {
  return ThirdPartyCatalogsService.postCatalogsUstcLookup({
    ustc: { ustc_id: parseInt(ustcId, 10) },
  });
};

export const getEdition = async (editionId: string): Promise<model_Edition> => {
  return EditionsService.getEditions({ editionId });
};

export const searchEditionsPage = async (query?: search_Query) =>
  EditionsService.postEditionsSearch({ edition: query });

export type ReprintRelationship = {
  editionKey: string;
  reprintOf: string;
};

const postReprintRequest = async <T>(
  path: string,
  body: unknown,
  token: string,
): Promise<T> => {
  const response = await fetch(`${OpenAPI.BASE}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    const detail = await response.text();
    throw new Error(detail || `Request failed (${response.status})`);
  }
  return response.json() as Promise<T>;
};

export const detectReprints = async (token: string) =>
  postReprintRequest<{ candidates: ReprintRelationship[] }>(
    "/editions/reprints/detect",
    {},
    token,
  );

export const applyReprints = async (
  token: string,
  relationships: ReprintRelationship[],
) =>
  postReprintRequest<{ updated: string[]; skipped: string[] }>(
    "/editions/reprints/apply",
    { relationships },
    token,
  );

const pageSignature = (items: model_Edition[] | undefined) =>
  (items || []).map((item) => item.key || "").join("|");

export const listAllEditions = async (
  query?: Omit<search_Query, "offset" | "limit">,
): Promise<model_Edition[]> => {
  const limit = 500;
  let offset = 0;
  const results: model_Edition[] = [];
  let previousPage = "";

  while (true) {
    const page = await searchEditionsPage({
      ...query,
      offset,
      limit,
    });
    const items = page.items || [];
    const currentPage = pageSignature(items);
    if (currentPage && currentPage === previousPage) {
      console.warn(
        "Pagination seems to be stuck, stopping to avoid infinite loop.",
        { currentPage },
      );
      break;
    }
    results.push(...items);
    if (
      items.length === 0 ||
      items.length < limit ||
      (page.total !== undefined && results.length >= page.total)
    ) {
      break;
    }
    previousPage = currentPage;
    offset += limit;
  }

  return results;
};
