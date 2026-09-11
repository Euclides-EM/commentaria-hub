# Run OCR Pipeline


## Glasgow_1756

Data set is public domain. I run the full OCR pipeline.

**Next steps:** Run the OCR pipeline: https://euclides.huma-num.fr/hub/?datasetId=ds_ii8nbl

## The_Hague_1758

Data set is public domain. I run the full OCR pipeline.

**Next steps:** Run the OCR pipeline: https://euclides.huma-num.fr/hub/index.h?datasetId=ds_9m13nh

# Run LLM Corrector

## Rome_1574

Two facsimiles are available:

* **`Rome_1574_transkribus`**: Public domain. The existing Transkribus transcriptions have been scraped and committed. I have also run OCR on the facsimile.
* **`Rome_1574`**: Copyright status allows research, but probably not display, of the facsimile. OCR has already been completed.

**Status**: currently running with Fable page-by-page. Progress so far, around 10%.

**Next steps:** Liri to run the corrector on the Rome_1574_transkribus facsimile with "include_edition_transcription" set to "true": https://euclides.huma-num.fr/hub/index.html?datasetId=ds_0m73by&annotationId=ann_lyz5sb&currentPageOrKey=377

## Paris_1794

Facsimile is public domain, I run the full OCR pipeline.

**Next steps:** Liri to run the corrector on the Paris_1794 facsimile: https://euclides.huma-num.fr/hub/index.html?datasetId=ds_fcnxho&datasetTab=annotations&annotationId=ann_1eeww1&currentPageOrKey=210&annotationTab=text

## Paris_1566

Facsimile is public domain, I run the full OCR pipeline.

**Next steps:** Run LLM corrector: https://euclides.huma-num.fr/hub/index.html?datasetId=ds_sfmbfc

## Basel_1562

The OCR stage has already been completed on an old facsimile, that has been replaced since with a public domain facsimile.

This means we need to re-align ALTOs when running LLM.

**Next steps:** Realign ALTOs and run LLM.

# Alignment

## Paris_1667

OCR has been completed, the facsimile is public domain.

Four cycles of LLM corrector run, due to token limit. Summarized metrics:

```shell
pages=388
tokens_input=796
tokens_cached=3,701,775
tokens_cache_creation=1,386,033
tokens_output=296,854
tokens_reasoning=0
tokens_total=5,385,458
cost_usd=$46.742450
```

**Next steps:** Align the existing Fable Markdown files with the new public-domain facsimile + manual curation (I'm currently in p. 114 in my manual curation).

## Paris_1634, Kiel_and_Leipzig_1699 and Basel_1537

Full Fable transcriptions exist and have already been manually curated. However, no copyright for the facsimile that the transcriptions were based on. Another public domain facsimile is available, but the transcriptions have not yet been aligned with it.

**Next steps:** Align the existing Fable Markdown files with the new public-domain facsimile.

# Manual Curation

## Nuremberg_1821

OCR has been completed, the facsimile is public domain.

Liri run the corrector:
```shell
2026/09/05 11:40:17 complete pages=72 rounds=1 requests=72 tokens_input=340 tokens_cached=1697451 tokens_cache_creation=583576 tokens_output=165781 tokens_reasoning=0 tokens_total=2447148 cost_usd=21.876591 cost_reports=72/72 final_outputs=store/data/ds_0n6l0d/annotations/ann_i74rcq/transcriptions/page-NNNN/original.md
```

**Next steps:** Manually curate the Fable transcriptions.

## Paris_1536

Full Fable transcriptions exist, but they still require manual curation. Facsimile is public domain.

**Next steps:** Manually curate the Fable transcriptions.

## Paris_1615

OCR has been completed, but the facsimile is not copyrighted.

There is a new facsimile that has the appropriate copyright. I run the full OCR pipeline on it.

Full correction with Codex Sol was run using dir mode. Token logs were lost due to job technical failure.

**Next steps:** Manually curate the Fable transcriptions.


## Venice_1482

Two facsimiles are available:

* **`Venice_1482_transkribus`**: The Transkribus facsimile is openly licensed, but the image quality is relatively low. I scraped the existing Transkribus transcription and stored it locally, but the transcription quality is poor. On the server, the processing pipeline has reached the **LineDetect** stage for this facsimile.
* **`Venice_1482`**: The facsimile is public domain. OCR has already been completed.

Applied rule:

```json
{
  "provider": "codex",
  "model": "gpt-5.6-sol",
  "rounds": 1,
  "execution_mode": "directory",
  "skip_existing": true,
  "additional_annotations": null,
  "include_edition_transcription": false
}
```

Usage summary:

- **Pages:** 284
- **Requests:** 1
- **Input tokens:** 1,691,046
    - Cached: 1,574,528
    - Non-cached: 116,518
- **Output tokens:** 19,273
- **Reasoning tokens:** 0
- **Total tokens:** 1,710,319
- **Cached input:** ~93.1%
- **Cost:** Unavailable

**Next steps:** Manually curate the Codex transcriptions.

# Completed 

## Paris_1639  

Full Fable transcriptions exist and have already been manually curated. The facsimile is public domain.

DONE

## Lyon_1557

The facsimile is public domain and can be used. OCR and llm correction with fable has been completed.

LLM corrector run in two cycles, due to token limit. Log available only from the second run:

```shell
2026/09/04 15:55:49 complete pages=79 rounds=1 requests=79 tokens_input=326 tokens_cached=1509871 tokens_cache_creation=659121 tokens_output=222941 tokens_reasoning=0 tokens_total=2392259 cost_usd=26.574715 cost_reports=79/79 final_outputs=store/data/ds_5da9w5/annotations/ann_8k4g9h/transcriptions/page-NNNN/original.md
```

Curated Fable transcriptions have been committed.

# Open Questions

## Paris_1598a

OCR has been completed, but the copyright status is unclear. There is no public domain facsimile available for this edition (Liri checked).

**Next steps:** Discuss with PIs what to do with editions that do not have a public domain facsimile available.

I could not find any facsimiles with explicit rights for the following works. 

Discuss with PIs what to do with these editions. Think of alternatives before the discussion.

- Paris_1564
- Paris_1598a
- London_1747
- HZ5UVJ

The following editions have facsimiles that have restrictive rights:
- London_1678a
- Basel_1533 

Status is unclear for the following editions:
- Leiden_1606
