## Venice_1482

Two facsimiles are available:

* **`Venice_1482_transkribus`**: The Transkribus facsimile is openly licensed, but the image quality is relatively low. I scraped the existing Transkribus transcription and stored it locally, but the transcription quality is poor. On the server, the processing pipeline has reached the **LineDetect** stage for this facsimile.
* **`Venice_1482`**: The copyright status of this facsimile is currently unknown. OCR has already been completed.

**Next steps:** Wait for Liri to update the copyright status of the facsimiles. Based on the result, decide which facsimile to use and continue the appropriate pipeline.

---

## Venice_1505 vol 1

The copyright status of the facsimile is currently unknown. Only a **Base annotation** has been completed so far.

**Next steps:** Wait for Liri to update the copyright status. If the facsimile can be used, continue the pipeline. Otherwise, delete it and locate a suitable replacement.

Note: To fully complete the transcription of this edition, I need to handle the second volume as well, which is a separate facsimile.

---

## Lyon_1557

The facsimile is public domain and can be used. OCR is currently running.

**Next steps:** Once OCR is complete, run the LLM corrector.

---

## Basel_1562

The copyright status of the facsimile is currently unknown. The OCR stage has already been completed.

**Next steps:** Wait for Liri to update the copyright status. If the facsimile can be used, proceed to the LLM corrector. Otherwise, restart the pipeline with a suitable facsimile.

---

## Rome_1574

Two facsimiles are available:

* **`Rome_1574_transkribus`**: Public domain. The existing Transkribus transcriptions have been scraped and committed. I have also run OCR on the facsimile.
* **`Rome_1574`**: Copyright status is currently unknown. OCR has already been completed.

**Next steps:** Wait for Liri to update the copyright status of the second facsimile, then decide which version to use for the LLM correction stage.

---

## Paris_1598a, Paris_1615, Paris_1667 and Nuremberg_1821

OCR has been completed for all four facsimiles, but their copyright status is currently unknown.

**Next steps:** Wait for Liri to update the copyright status. For each facsimile that can be used, proceed to the LLM correction stage. For any that cannot be used, identify a suitable replacement and restart the necessary parts of the pipeline.

---

## Paris_1634, Paris_1639 and Kiel_and_Leipzig_1699

Full Fable transcriptions exist and have already been manually curated. However, the copyright status of the corresponding facsimiles is currently unknown.

**Next steps:** Wait for Liri to update the copyright status. If the existing facsimiles can be used, no further transcription work should be necessary. Otherwise, determine whether the curated transcriptions can be aligned with suitable public-domain facsimiles before considering reprocessing.

---

## Paris_1536 and Basel_1537

The copyright status of the facsimiles is currently unknown. Full Fable transcriptions exist, but they still require manual curation.

**Next steps:** Wait for Liri to update the copyright status. If the existing facsimiles can be used, manually curate the Fable transcriptions. Otherwise, attempt to align the existing Fable Markdown files with the new public-domain facsimiles. If the alignment works sufficiently well, manually curate the resulting transcriptions.

---
