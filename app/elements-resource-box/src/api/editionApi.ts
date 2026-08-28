import type {
  model_Edition,
  model_EditionShelfmark,
  model_USTC,
  search_Query,
} from "@hub-api";
import {
  EditionsService,
  FeatureResultsService,
  OpenAPI,
  ShelfmarksService,
  ThirdPartyCatalogsService,
} from "@hub-api";
import { uploadImage } from "./imageApi.ts";

export type EditionWithShelfmarks = model_Edition & {
  shelfmarks?: model_EditionShelfmark[];
};

export const upsertEdition = async (
  data: EditionWithShelfmarks,
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

  const { shelfmarks = [], ...edition } = data;

  if (options?.isNew) {
    await EditionsService.postEditions({ edition });
  } else {
    await EditionsService.putEditions({ editionId: data.key!, edition });
    const existingShelfmarks = await ShelfmarksService.getEditionsShelfmarks({
      editionId: data.key!,
    });
    const nextIDs = new Set(
      shelfmarks.map((shelfmark) => shelfmark.id).filter(Boolean),
    );
    await Promise.all(
      existingShelfmarks
        .filter((shelfmark) => shelfmark.id && !nextIDs.has(shelfmark.id))
        .map((shelfmark) =>
          ShelfmarksService.deleteEditionsShelfmarks({
            editionId: data.key!,
            shelfmarkId: shelfmark.id!,
          }),
        ),
    );
  }

  await Promise.all(
    shelfmarks.map((shelfmark) =>
      shelfmark.id
        ? ShelfmarksService.putEditionsShelfmarks({
            editionId: data.key!,
            shelfmarkId: shelfmark.id,
            shelfmark,
          })
        : ShelfmarksService.postEditionsShelfmarks({
            editionId: data.key!,
            shelfmark,
          }),
    ),
  );
};

export const deleteEdition = async (editionId: string): Promise<void> => {
  await EditionsService.deleteEditions({ editionId });
};

export const updateEditionSubjectCategories = async (
  editionId: string,
  categories: { category: string; classification: string }[],
): Promise<void> => {
  await FeatureResultsService.postFeaturesResults({
    result: [
      {
        scope: { type: "editions" },
        key: editionId,
        feature_id: "m_classifier",
        values: categories.map(({ category, classification }) => ({
          surface: `${category}::${classification}`,
        })),
      },
    ],
  });
};

export const ustcLookup = async (
  ustcId: string,
): Promise<Partial<model_USTC>> => {
  return ThirdPartyCatalogsService.postCatalogsUstcLookup({
    ustc: { ustc_id: parseInt(ustcId, 10) },
  });
};

export const getEdition = async (
  editionId: string,
): Promise<EditionWithShelfmarks> => {
  const [edition, shelfmarks] = await Promise.all([
    EditionsService.getEditions({ editionId }),
    ShelfmarksService.getEditionsShelfmarks({ editionId }),
  ]);
  return { ...edition, shelfmarks };
};

export const searchEditionsPage = async (query?: search_Query) =>
  EditionsService.postEditionsSearch({ edition: query });

export const listEditionShelfmarks = (editionId: string) =>
  ShelfmarksService.getEditionsShelfmarks({ editionId });

const SHELFMARKS_BATCH_SIZE = 50;

const chunksOf = <T>(items: T[], size: number): T[][] => {
  const chunks: T[][] = [];
  for (let i = 0; i < items.length; i += size) {
    chunks.push(items.slice(i, i + size));
  }
  return chunks;
};

export const listShelfmarks = async (
  editionIds?: string[],
): Promise<model_EditionShelfmark[]> => {
  if (!editionIds || editionIds.length === 0) {
    return ShelfmarksService.getShelfmarks({});
  }

  const uniqueEditionIds = Array.from(
    new Set(editionIds.filter((editionId) => Boolean(editionId))),
  );
  if (uniqueEditionIds.length === 0) {
    return [];
  }

  const batches = await Promise.all(
    chunksOf(uniqueEditionIds, SHELFMARKS_BATCH_SIZE).map((batch) =>
      ShelfmarksService.getShelfmarks({ editionId: batch }),
    ),
  );
  return batches.flat();
};

export const attachShelfmarks = async <T extends model_Edition>(
  editions: T[],
): Promise<Array<T & { shelfmarks: model_EditionShelfmark[] }>> => {
  const editionIds = editions
    .map((edition) => edition.key)
    .filter((key): key is string => Boolean(key));
  if (editionIds.length === 0) {
    return editions.map((edition) => ({ ...edition, shelfmarks: [] }));
  }
  const shelfmarks = await listShelfmarks(editionIds);
  const byEdition = new Map<string, model_EditionShelfmark[]>();
  for (const shelfmark of shelfmarks) {
    if (!shelfmark.edition_id) continue;
    byEdition.set(shelfmark.edition_id, [
      ...(byEdition.get(shelfmark.edition_id) ?? []),
      shelfmark,
    ]);
  }
  return editions.map((edition) => ({
    ...edition,
    shelfmarks: edition.key ? (byEdition.get(edition.key) ?? []) : [],
  }));
};

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
