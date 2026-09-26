# Run OCR Pipeline

## London_1570

WIP https://euclides.huma-num.fr/hub/?datasetId=ds_9lpdf8

## Lyon_1672

WIP https://euclides.huma-num.fr/hub/?datasetId=ds_71gygq

# Run LLM Corrector

# Alignment

# Manual Curation

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

**Next steps:** manual curation (I'm currently in p. 149 in my manual curation).

## Paris_1566

Facsimile is public domain, I run the full OCR pipeline.

Run LLM corrector: https://euclides.huma-num.fr/hub/index.html?datasetId=ds_sfmbfc with both Fable and Codex Sol, in parts.

Codex usage:

- **Requests:** 246
- **Input tokens:** 23,798,002
- **Cached tokens:** 19,914,368
- **Cache creation tokens:** 0
- **Output tokens:** 1,209,926
- **Reasoning tokens:** 0
- **Total tokens:** 25,007,928

Fable usage:

- **Requests:** 234
- **Input tokens:** 1,020
- **Cached tokens:** 5,293,561
- **Cache creation tokens:** 2,219,009
- **Output tokens:** 1,137,298
- **Reasoning tokens:** 0
- **Total tokens:** 8,650,888
- **Cost:** $107.55

**Next steps:** Manually curate the transcriptions.

## Antwerp_1654

Run the LLM Corrector on:
https://euclides.huma-num.fr/hub/?datasetId=ds_9ni7al&annotationId=ann_fy0hyh&currentPageOrKey=209&annotationTab=text

Codex Sol:

- **Requests:** 37
- **Input tokens:** 1,078,204
- **Cached tokens:** 673,152
- **Cache creation tokens:** 0
- **Output tokens:** 53,313
- **Reasoning tokens:** 0
- **Total tokens:** 1,131,517

Fable:

- **Requests:** 376
- **Input tokens:** 1,508
- **Cached tokens:** 6,677,114
- **Cache creation tokens:** 2,221,300
- **Output tokens:** 434,684
- **Reasoning tokens:** 0
- **Total tokens:** 9,334,606
- **Cost:** $73.72

**Next steps:** Manually curate the transcriptions.

## The_Hague_1758

Data set is public domain. I run the full OCR pipeline.

Run the LLM corrector: https://euclides.huma-num.fr/hub/index.h?datasetId=ds_9m13nh&annotationId=ann_6jfdp6&currentPageOrKey=61&annotationTab=details with both Fable and Codex Sol, in parts.

Fable usage:

- **Requests:** 382
- **Input tokens:** 1,536
- **Cached tokens:** 6,908,995
- **Cache creation tokens:** 2,894,093
- **Output tokens:** 493,701
- **Reasoning tokens:** 0
- **Total tokens:** 10,298,325
- **Cost:** $90.53

### Codex — GPT-5.6 Sol

- **Requests:** 28
- **Input tokens:** 1,675,677
- **Cached tokens:** 1,338,624
- **Cache creation tokens:** 0
- **Output tokens:** 81,425
- **Reasoning tokens:** 0
- **Total tokens:** 1,757,102

**Next steps:** Manually curate the transcriptions.

## Basel_1562

The OCR stage has already been completed on an old facsimile, that has been replaced since with a public domain facsimile.

Re-aligned ALTOs (-1 offset) then run LLM.

Fable:

- **Requests:** 39
- **Input tokens:** 170
- **Cached tokens:** 844,396
- **Cache creation tokens:** 349,874
- **Output tokens:** 126,374
- **Reasoning tokens:** 0
- **Total tokens:** 1,320,814
- **Cost:** $14.30

Codex:

- **Requests:** 174
- **Input tokens:** 16,423,825
- **Cached tokens:** 13,554,432
- **Cache creation tokens:** 0
- **Output tokens:** 838,437
- **Reasoning tokens:** 0
- **Total tokens:** 17,262,262

**Next steps:** Manually curate the transcriptions.

## Glasgow_1756

Data set is public domain. I run the full OCR pipeline.

Run the LLM corrector: https://euclides.huma-num.fr/hub/?datasetId=ds_ii8nbl&annotationId=ann_ab8glk&currentPageOrKey=61&annotationTab=details

Done 323/431 with Codex, the rest with Fable.

Codex run stats:

- **Provider:** Codex
- **Model:** GPT-5.6 Sol
- **Total requests:** 401
- **Input tokens:** 13,162,647
- **Cached input tokens:** 9,313,280
- **Uncached input tokens:** 3,849,367
- **Output tokens:** 816,648
- **Reasoning tokens:** 0
- **Total tokens:** 13,979,295
- **Cache hit rate (tokens):** ~70.8%
- **Cache read requests:** 393 / 396 opportunities (~99.2%)
- **Misses after warmup:** 3

Fable stats:

- **Requests:** 108
- **Input tokens:** 432
- **Cached input tokens:** 1,939,980
- **Cache creation tokens:** 774,316
- **Output tokens:** 106,875
- **Reasoning tokens:** 0
- **Total tokens:** 2,821,603
- **Cost:** $23.054672
- **Cache read opportunities:** 107
- **Cache read requests:** 107
- **Misses after warmup:** 0

**Next steps:** Manually curate the transcriptions.

## Paris_1794

Facsimile is public domain, I run the full OCR pipeline.

Run the corrector on the Paris_1794 facsimile: https://euclides.huma-num.fr/hub/index.html?datasetId=ds_fcnxho&datasetTab=annotations&annotationId=ann_1eeww1&currentPageOrKey=210&annotationTab=text
using Codex. Partial run stats:

- Requests: 90
- Input tokens: 2,204,507
- Cached tokens: 1,490,816
- Cache creation tokens: 0
- Output tokens: 96,714
- Reasoning tokens: 0
- Total tokens: 2,301,221
- Cache hit requests: 89/89 (100%)

**Next steps:** Manually curate the transcriptions.

## Rome_1574

Two facsimiles are available:

* **`Rome_1574_transkribus`**: Public domain. The existing Transkribus transcriptions have been scraped and committed. I have also run OCR on the facsimile.
* **`Rome_1574`**: Copyright status allows research, but probably not display, of the facsimile. OCR has already been completed.

Run the corrector on the Rome_1574_transkribus facsimile with "include_edition_transcription" set to "true": https://euclides.huma-num.fr/hub/index.html?datasetId=ds_0m73by&annotationId=ann_lyz5sb&currentPageOrKey=377

Done 734/753 pages with Fable and the rest with Codex Sol. Got very partial token counts.

Fable:
- Input tokens: 2
- Cache creation tokens: 4,951
- Cache read tokens: 7,346
- Output tokens: 92
- Haiku 4.5 tokens: 4,210
- Total tokens: 16,601
- Total cost: $0.115292

Codex:
- Requests: 19
- Input tokens: 1,536,835
- Cached tokens: 1,285,504
- Cache creation tokens: 0
- Output tokens: 71,078
- Reasoning tokens: 0
- Total tokens: 1,607,913
- Cache hit requests: 18/18 (100%)
- Cost: unavailable

**Next steps:** Manually curate the transcriptions.

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

The transcription was not that good, so we tried the page-by-page option with Codex Sol for 30 pages, and added it to the transcriptions. Relevant pages: `17,35,68,82,101,124,156,193,208,227,259,274,296,312,341,363,389,421,437,455,476,494,509,528,551,578,603,615,632,645`

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

## Paris_1634, Kiel_and_Leipzig_1699 and Basel_1537

Full Fable transcriptions exist and have already been manually curated. Transcriptions manually aligned to public the new public-domain facsimile.

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
