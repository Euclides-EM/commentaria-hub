## Venice_1482

Two facsimiles are available:

* **`Venice_1482_transkribus`**: The Transkribus facsimile is openly licensed, but the image quality is relatively low. I scraped the existing Transkribus transcription and stored it locally, but the transcription quality is poor. On the server, the processing pipeline has reached the **LineDetect** stage for this facsimile.
* **`Venice_1482`**: The facsimile is public domain. OCR has already been completed.

**Next steps:** Liri to run the corrector on the Venice_1482 facsimile: https://euclides.huma-num.fr/hub/index.html?datasetId=ds_6vegrr&annotationId=ann_3pg8ot&currentPageOrKey=149&annotationTab=details

---

## Lyon_1557

The facsimile is public domain and can be used. OCR and llm correction with fable has been completed.

LLM corrector run in two cycles, due to token limit. Log available only from the second run:

```shell
2026/09/04 15:55:49 complete pages=79 rounds=1 requests=79 tokens_input=326 tokens_cached=1509871 tokens_cache_creation=659121 tokens_output=222941 tokens_reasoning=0 tokens_total=2392259 cost_usd=26.574715 cost_reports=79/79 final_outputs=store/data/ds_5da9w5/annotations/ann_8k4g9h/transcriptions/page-NNNN/original.md
```

**Next steps:** Manually curate the Fable transcriptions.

---

## Basel_1562

The copyright status of the facsimile is currently unknown. The OCR stage has already been completed.

**Next steps:** Wait for Liri to update the copyright status. If the facsimile can be used, proceed to the LLM corrector. Otherwise, restart the pipeline with a suitable facsimile.

Q to Liri: there is no public domain scan for this edition, can you check: https://euclides.huma-num.fr/resourcebox/item/edit?key=Basel_1562

---

## Rome_1574

Two facsimiles are available:

* **`Rome_1574_transkribus`**: Public domain. The existing Transkribus transcriptions have been scraped and committed. I have also run OCR on the facsimile.
* **`Rome_1574`**: Copyright status allows research, but probably not display, of the facsimile. OCR has already been completed.

**Next steps:** Liri to run the corrector on the Rome_1574_transkribus facsimile with "include_edition_transcription" set to "true": https://euclides.huma-num.fr/hub/index.html?datasetId=ds_0m73by&annotationId=ann_lyz5sb&currentPageOrKey=377

---

## Paris_1598a

OCR has been completed, but the copyright status is unclear.

Q to Liri: The copyright status of Paris_1598a is unclear, can you check: https://euclides.huma-num.fr/resourcebox/item/edit?key=Paris_1598a

---

## Paris_1615

OCR has been completed, but the facsimile is not copyrighted. There is a new facsimile that has the appropriate copyright.

**Next steps:** Create a new dataset with the new facsimile and run all the steps.

---

## Paris_1667

OCR has been completed, the facsimile is public domain.

**Next steps:** Liri to run the corrector on the two parts of the Paris_1667 facsimile:
https://euclides.huma-num.fr/hub/index.html?datasetId=ds_jinmif&datasetTab=details&annotationId=ann_thiq1g&currentPageOrKey=171
https://euclides.huma-num.fr/hub/index.html?datasetId=ds_jinmif&datasetTab=details&annotationId=ann_aotfx5&currentPageOrKey=86

---

## Nuremberg_1821

OCR has been completed, the facsimile is public domain.

**Next steps:** Liri to run the corrector on the following annotation, with `additional_annotations=["ann_m49sua"]`:  https://euclides.huma-num.fr/hub/index.html?datasetId=ds_tjokpg&datasetTab=details&annotationId=ann_ov9taz&currentPageOrKey=52&annotationTab=details

---

## Paris_1634

Full Fable transcriptions exist and have already been manually curated. However, no copyright for the facsimile that the transcriptions were based on. Another public domain facsimile is available, but the transcriptions have not yet been aligned with it.

**Next steps:** Align the existing Fable Markdown files with the new public-domain facsimile.

---

## Paris_1639 and Kiel_and_Leipzig_1699

Full Fable transcriptions exist and have already been manually curated. However, the copyright status of the corresponding facsimiles is currently unknown.

**Next steps:** Wait for Liri to update the copyright status. If the existing facsimiles can be used, no further transcription work should be necessary. Otherwise, determine whether the curated transcriptions can be aligned with suitable public-domain facsimiles before considering reprocessing.

Q to Liri: The copyright status of Paris_1639 and Kiel_and_Leipzig_1699 is unclear, can you check: https://euclides.huma-num.fr/resourcebox/item/edit?key=Paris_1639

---

## Paris_1536

Full Fable transcriptions exist, but they still require manual curation. Facsimile is public domain.

**Next steps:** Manually curate the Fable transcriptions. 

---

## Basel_1537

The copyright status of the facsimiles is currently unknown. Full Fable transcriptions exist, but they still require manual curation.

**Next steps:** Wait for Liri to update the copyright status. If the existing facsimiles can be used, manually curate the Fable transcriptions. Otherwise, attempt to align the existing Fable Markdown files with the new public-domain facsimiles. If the alignment works sufficiently well, manually curate the resulting transcriptions.

Q to Liri: The copyright status of Basel_1537 is unclear, can you check: https://euclides.huma-num.fr/resourcebox/item/edit?key=Basel_1537

---
