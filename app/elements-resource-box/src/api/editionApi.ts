import type { model_Edition, model_USTC, search_Query } from "@hub-api";
import { EditionsService, ThirdPartyCatalogsService } from "@hub-api";
import { uploadImage } from "./imageApi.ts";

export const upsertEdition = async (
  data: model_Edition,
  images: Record<string, File>,
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
          shelfmark.title_page_img = await uploadImage(
            data.key!,
            file,
            `tp_${i + 1}`,
          );
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
            `frontispiece_${i + 1}`,
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
            example.img = await uploadImage(
              data.key!,
              file,
              `visEl_${i + 1}_ex_${j + 1}`,
            );
          })(),
        );
      }
    }
  }

  await Promise.all(uploads);

  try {
    await EditionsService.getEditions({ editionId: data.key! });
    await EditionsService.putEditions({ editionId: data.key!, edition: data });
  } catch {
    await EditionsService.postEditions({ edition: data });
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

export const listAllEditions = async (
  query?: Omit<search_Query, "offset" | "limit">,
): Promise<model_Edition[]> => {
  const limit = 500;
  let offset = 0;
  const results: model_Edition[] = [];

  while (true) {
    const page = await searchEditionsPage({
      ...query,
      offset,
      limit,
    });
    const items = page.items || [];
    results.push(...items);
    if (
      items.length === 0 ||
      items.length < limit ||
      (page.total !== undefined && results.length >= page.total)
    ) {
      break;
    }
    offset += limit;
  }

  return results;
};
