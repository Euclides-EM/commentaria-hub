import {
  EditionsService,
  FacsimilesService,
  model_DiagramCropVolume,
} from "@hub-api";
import type { model_DiagramCrops } from "@hub-api";

export interface VolumeData {
  volume?: number;
  key: string;
  images: string[];
  hasNoDiagrams: boolean;
}

export interface DiagramsResult {
  images?: string[];
  hasNoDiagrams?: boolean;
  volumes?: VolumeData[];
  error?: string;
}

const mapImages = (imagesByName?: Record<string, string>): string[] => {
  return Object.values(imagesByName || {});
};

const mapVolume = (
  key: string,
  volume: model_DiagramCropVolume,
): VolumeData => ({
  volume: volume.volume,
  key: volume.key || key,
  images: mapImages(volume.imageUrlsByName),
  hasNoDiagrams: !volume.hasDiagrams,
});

const mapDiagramResponse = (
  key: string,
  response: model_DiagramCrops,
): DiagramsResult => {
  if (response.volumes && response.volumes.length > 0) {
    return {
      hasNoDiagrams: !response.hasDiagrams,
      volumes: response.volumes.map((volume) => mapVolume(key, volume)),
    };
  }

  return {
    hasNoDiagrams: !response.hasDiagrams,
    images: mapImages(response.imageURLsByName),
  };
};

export const fetchDiagrams = async (
  editionId: string,
  facsimileId?: string | null,
): Promise<DiagramsResult> => {
  if (!editionId && !facsimileId) {
    return {
      images: [],
      hasNoDiagrams: false,
      error: "No key provided",
    };
  }

  try {
    const response = facsimileId
      ? await FacsimilesService.getFacsimiliesDiagrams({ id: facsimileId })
      : await EditionsService.getEditionsDiagrams({ editionId });
    return mapDiagramResponse(editionId, response);
  } catch {
    return {
      images: [],
      hasNoDiagrams: false,
      error: "Failed to load diagrams",
    };
  }
};
