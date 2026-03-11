import { DatasetImagesService } from "@hub-api";
import { TITLE_PAGES_DATASET_ID } from "../constants";

export const uploadImage = async (
  key: string,
  file: File,
  type: "facsimile" | "tp" | "frontispiece",
) => {
  console.log("Uploading image...", file.name);
  const result = await DatasetImagesService.postDatasetsImagesUpload({
    dataSetId: TITLE_PAGES_DATASET_ID,
    key,
    type,
    file,
  });
  return result.path;
};
