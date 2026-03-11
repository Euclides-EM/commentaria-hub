import { TranscriptionsService } from "@hub-api";

const getCommentariaAppUrl = () => {
  const baseUrl = import.meta.env.VITE_COMMENTARIA_APP_URL;
  return typeof baseUrl === "string" && baseUrl.length > 0 ? baseUrl : null;
};

export const buildCommentariaHubTranscriptionUrl = ({
  annotationId,
  datasetId,
  editionKey,
}: {
  annotationId: string;
  datasetId: string;
  editionKey: string;
}) => {
  let baseUrl = getCommentariaAppUrl();
  if (!baseUrl) {
    return null;
  }
  if (!baseUrl.endsWith("/")) {
    baseUrl += "/";
  }

  const url = new URL(baseUrl);
  url.searchParams.set("datasetId", datasetId);
  url.searchParams.set("annotationId", annotationId);
  url.searchParams.set("currentPageOrKey", editionKey);

  return url.toString();
};

export const getCommentariaHubPreferredTranscriptionUrl = async (
  editionKey: string,
) => {
  if (!editionKey) {
    return null;
  }

  const transcriptions = await TranscriptionsService.getEditionsTranscriptions({
    editionId: [editionKey],
  });
  const preferred =
    transcriptions.find(
      (transcription) =>
        transcription.edition_id === editionKey &&
        transcription.preferred_annotation?.dataset_id &&
        transcription.preferred_annotation?.id,
    )?.preferred_annotation ??
    transcriptions.find(
      (transcription) =>
        transcription.preferred_annotation?.dataset_id &&
        transcription.preferred_annotation?.id,
    )?.preferred_annotation;

  if (!preferred?.dataset_id || !preferred.id) {
    return null;
  }

  return buildCommentariaHubTranscriptionUrl({
    annotationId: preferred.id,
    datasetId: preferred.dataset_id,
    editionKey,
  });
};
