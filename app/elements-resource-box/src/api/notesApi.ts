import { EditionsService } from '@hub-api';

export const saveNote = async (
  editionId: string,
  data: { note: string },
): Promise<void> => {
  console.log("Saving note", { editionId, ...data });
  await EditionsService.postEditionsNotes({
    editionId,
    note: { note: data.note },
  });
};
