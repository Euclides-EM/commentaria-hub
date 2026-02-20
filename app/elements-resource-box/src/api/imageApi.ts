import { EditionsService } from "@hub-api";
import { TITLE_PAGES_DATASET_ID } from "../constants";

export const uploadImage = async (key: string, file: File, type: string) => {
  console.log("Uploading image...", file.name);
  const result = await EditionsService.postDatasetsImagesUpload({
    dataSetId: TITLE_PAGES_DATASET_ID,
    key,
    type,
    file,
  });
  return result.path;
};
