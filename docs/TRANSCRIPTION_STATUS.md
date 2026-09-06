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

The OCR stage has already been completed on an old facsimile, that has been replaced since with a public domain facsimile. 

**Next steps:** Recreate the dataset with the new facsimile, run OCR and LLM corrector, and manually curate the Fable transcriptions.

---

## Rome_1574

Two facsimiles are available:

* **`Rome_1574_transkribus`**: Public domain. The existing Transkribus transcriptions have been scraped and committed. I have also run OCR on the facsimile.
* **`Rome_1574`**: Copyright status allows research, but probably not display, of the facsimile. OCR has already been completed.

**Next steps:** Liri to run the corrector on the Rome_1574_transkribus facsimile with "include_edition_transcription" set to "true": https://euclides.huma-num.fr/hub/index.html?datasetId=ds_0m73by&annotationId=ann_lyz5sb&currentPageOrKey=377

---

## Paris_1598a

OCR has been completed, but the copyright status is unclear. There is no public domain facsimile available for this edition (Liri checked).

**Next steps:** Discuss with PIs what to do with editions that do not have a public domain facsimile available.

---

## Paris_1615

OCR has been completed, but the facsimile is not copyrighted. There is a new facsimile that has the appropriate copyright.

I created a new dataset, segmentation stage is currently running. I will run the OCR and LLM corrector on it.

**Next steps:** Run all the steps - OCR, LLM corrector, and manual curation - on the new facsimile: https://euclides.huma-num.fr/hub/index.html?datasetId=ds_y46hr6&datasetTab=annotations

---

## Paris_1667

OCR has been completed, the facsimile is public domain.

**Next steps:** Liri to run the corrector on the two parts of the Paris_1667 facsimile:
https://euclides.huma-num.fr/hub/index.html?datasetId=ds_jinmif&datasetTab=details&annotationId=ann_thiq1g&currentPageOrKey=171
https://euclides.huma-num.fr/hub/index.html?datasetId=ds_jinmif&datasetTab=details&annotationId=ann_aotfx5&currentPageOrKey=86

---

## Nuremberg_1821

OCR has been completed, the facsimile is public domain.

Liri run the corrector:
```shell
2026/09/05 11:40:17 complete pages=72 rounds=1 requests=72 tokens_input=340 tokens_cached=1697451 tokens_cache_creation=583576 tokens_output=165781 tokens_reasoning=0 tokens_total=2447148 cost_usd=21.876591 cost_reports=72/72 final_outputs=store/data/ds_0n6l0d/annotations/ann_i74rcq/transcriptions/page-NNNN/original.md
```

**Next steps:** Manually curate the Fable transcriptions.

---

## Paris_1634, Kiel_and_Leipzig_1699 and Basel_1537

Full Fable transcriptions exist and have already been manually curated. However, no copyright for the facsimile that the transcriptions were based on. Another public domain facsimile is available, but the transcriptions have not yet been aligned with it.

**Next steps:** Align the existing Fable Markdown files with the new public-domain facsimile.

---

## Paris_1639  

Full Fable transcriptions exist and have already been manually curated. The facsimile is public domain.

DONE

---

## Paris_1536

Full Fable transcriptions exist, but they still require manual curation. Facsimile is public domain.

**Next steps:** Manually curate the Fable transcriptions. 

---

## Paris_1794 

Creating the dataset now, facsimile is public domain. 

**Next steps:** Run all the steps - OCR, LLM corrector, and manual curation - on the new facsimile: https://euclides.huma-num.fr/hub/index.html?datasetId=ds_fcnxho
