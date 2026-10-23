# OCR Pipeline Queue

> Rome_1574_vol2 - TODO
> Strasbourg_1564 - https://euclides.huma-num.fr/hub/?datasetId=ds_ypncw4 - OCR running
> Pisa_1658 - https://euclides.huma-num.fr/hub/?datasetId=ds_469aie - LineDetect running
> Würzburg_1661 - https://euclides.huma-num.fr/hub/?datasetId=ds_n0qsmi - Seg running [LARGE FILE]
> Paris_1814 vol1 - https://euclides.huma-num.fr/hub/?datasetId=ds_mz3oph - OCR running
> Paris_1814 vol3 - https://euclides.huma-num.fr/hub/?datasetId=ds_9s7hg3 - OCR running

# Train to improve segmentation

> Venice_1505 - https://euclides.huma-num.fr/hub/?datasetId=ds_j4i726 - Seg job done. Need to train to improve. - very similar issues to Venice_1543 [LARGE FILE].

> Venice_1543 - https://euclides.huma-num.fr/hub/?datasetId=ds_nbbyj6 - Seg job done. Need to train to improve. [LARGE FILE]

> Basel_1550 - https://euclides.huma-num.fr/hub/?datasetId=ds_wu29cj - Seg done. Might need train to improve, i.e page 83.

> Cambridge_1655 - https://euclides.huma-num.fr/hub/?datasetId=ds_x86ksn - Seg job done. Need to train to improve. i.e page 45.

> Rome_1589_vol1 - https://euclides.huma-num.fr/hub/?datasetId=ds_737a7n - Seg done. Need to train to improve.

> Rome_1589_vol2 - https://euclides.huma-num.fr/hub/?datasetId=ds_8e08i6 - Seg done. Need to train to improve. - very similar issues to Rome_1589_vol1

> Arnhem_1605a - https://euclides.huma-num.fr/hub/?datasetId=ds_d10ys7 - Seg done. Need to train to improve, i.e page 114.

> Arnhem_1605 - https://euclides.huma-num.fr/hub/?datasetId=ds_5koq2a -  Seg done. Need to train to improve. - very similar issues to Arnhem_1605a

> Edinburgh_1795 - https://euclides.huma-num.fr/hub/?datasetId=ds_4sklu9 - Seg done. Need to train to improve, i.e page 205.

> Paris_1693 - https://euclides.huma-num.fr/hub/?datasetId=ds_mi34kk - Seg done. Need to train to improve, probably can rely on above trained and no need to slice from it.

> Oxford_1701 - https://euclides.huma-num.fr/hub/?datasetId=ds_73fy51 - Seg done. Need to train to improve.

> Vienna_1694 - https://euclides.huma-num.fr/hub/?datasetId=ds_ak4z2r  Seg done. Need to train to improve.

# Run LLM Corrector

## Paris_1814 vol2

https://euclides.huma-num.fr/hub/?datasetId=ds_a2q3yy&annotationId=ann_pr4ar8

285/576

## Strasbourg_1566

https://euclides.huma-num.fr/hub/?datasetId=ds_ilx0l7&annotationId=ann_fsoleq

Running OCR, 190/221

# Manual Curation

## Paris_1685

https://euclides.huma-num.fr/hub/?datasetId=ds_fa95xr&annotationId=ann_i31qou

### Dataset Usage Summary

Claude Code — Fable

- **Requests:** 388
- **Input tokens:** 1,578
- **Cached tokens:** 9,173,575
- **Cache creation tokens:** 4,744,324
- **Output tokens:** 467,560
- **Reasoning tokens:** 0
- **Total tokens:** 14,387,037
- **Cost:** $128.22

## Oxford_1703

OCR for Latin:
https://euclides.huma-num.fr/hub/?datasetId=ds_ej5frn&annotationId=ann_skdyji

OCR for Greek:
https://euclides.huma-num.fr/hub/?datasetId=ds_ej5frn&annotationId=ann_bzfzp2

### Dataset Usage Summary

Codex — GPT-5.6 Sol

- **Requests:** 627
- **Input tokens:** 90,964,375
- **Cached tokens:** 77,983,104
- **Cache creation tokens:** 0
- **Output tokens:** 4,894,010
- **Reasoning tokens:** 0
- **Total tokens:** 95,858,385

Claude Code — Fable

- **Requests:** 57
- **Input tokens:** 270
- **Cached tokens:** 1,828,701
- **Cache creation tokens:** 1,077,683
- **Output tokens:** 496,574
- **Reasoning tokens:** 0
- **Total tokens:** 3,403,228
- **Cost:** $48.45

## Lyon_1672

https://euclides.huma-num.fr/hub/?datasetId=ds_71gygq&annotationId=ann_57z69k

Claude Code — Fable

- **Requests:** 790
- **Input tokens:** 3,172
- **Cached tokens:** 16,059,312
- **Cache creation tokens:** 7,433,879
- **Output tokens:** 618,653
- **Reasoning tokens:** 0
- **Total tokens:** 24,114,?16
- **Cost:** $197.06

Codex — GPT-5.6 Sol

- **Requests:** 108
- **Input tokens:** 2,562,295
- **Cached tokens:** 1,486,848
- **Cache creation tokens:** 0
- **Output tokens:** 135,186
- **Reasoning tokens:** 0
- **Total tokens:** 2,697,481

**Next steps:** Ready for curation.

## London_1570

https://euclides.huma-num.fr/hub/?datasetId=ds_9lpdf8&annotationId=ann_ea4io3

Claude Code — Fable

- **Runs:** 8
- **Requests:** 656
- **Input tokens:** 2,786
- **Cached tokens:** 14,816,611
- **Cache creation tokens:** 7,815,233
- **Output tokens:** 1,715,991
- **Reasoning tokens:** 0
- **Total tokens:** 24,350,621
- **Cost:** $259.51

Codex — GPT-5.6 Sol

- **Runs:** 4
- **Requests:** 267
- **Input tokens:** 12,776,948
- **Cached tokens:** 9,661,184
- **Cache creation tokens:** 0
- **Output tokens:** 1,023,810
- **Reasoning tokens:** 0
- **Total tokens:** 13,800,758

**Next steps:** Ready for curation.

## Pesaro_1572

https://euclides.huma-num.fr/hub/?datasetId=ds_g17zx1&annotationId=ann_uomk97

Claude Code — Fable

- **Requests:** 248
- **Input tokens:** 1,106
- **Cached tokens:** 6,172,118
- **Cache creation tokens:** 3,552,784
- **Output tokens:** 1,181,982
- **Reasoning tokens:** 0
- **Total tokens:** 10,907,990
- **Cost:** $137.20

Codex — GPT-5.6 Sol

- **Requests:** 237
- **Input tokens:** 34,277,269
- **Cached tokens:** 30,052,992
- **Cache creation tokens:** 0
- **Output tokens:** 1,260,640
- **Reasoning tokens:** 0
- **Total tokens:** 35,537,909

**Next steps:** Ready for curation.

# Completed

## Venice_1482

Two facsimiles are available:

* **`Venice_1482_transkribus`**: The Transkribus facsimile is openly licensed, but the image quality is relatively low. I scraped the existing Transkribus transcription and stored it locally, but the transcription quality is poor. On the server, the processing pipeline has reached the **LineDetect** stage for this facsimile.
* **`Venice_1482`**: The facsimile is public domain. OCR has already been completed.

https://euclides.huma-num.fr/hub/?datasetId=ds_6vegrr&annotationId=ann_3pg8ot

Claude Code — Fable

- **Requests:** 284
- **Input tokens:** 1,422
- **Cached tokens:** 8,850,210
- **Cache creation tokens:** 5,399,806
- **Output tokens:** 3,882,069
- **Reasoning tokens:** 0
- **Total tokens:** 18,133,507
- **Cost:** $312.29

## Lyon_1603

https://euclides.huma-num.fr/hub/?datasetId=ds_69jfii&annotationId=ann_d0ffde

Run with Codex:

- **Requests:** 89
- **Input tokens:** 9,011,177
- **Cached tokens:** 7,577,856
- **Cache creation tokens:** 0
- **Output tokens:** 411,740
- **Reasoning tokens:** 0
- **Total tokens:** 9,422,917

## Seville_1576

https://euclides.huma-num.fr/hub/?datasetId=ds_uwclpr&annotationId=ann_mz29ot

Done with Fable.

- **Requests:** 252
- **Input tokens:** 1,042
- **Cached tokens:** 5,261,731
- **Cache creation tokens:** 2,661,667
- **Output tokens:** 495,351
- **Reasoning tokens:** 0
- **Total tokens:** 8,419,791
- **Cost:** $83.87

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

Codex — GPT-5.6 Sol

- **Requests:** 28
- **Input tokens:** 1,675,677
- **Cached tokens:** 1,338,624
- **Cache creation tokens:** 0
- **Output tokens:** 81,425
- **Reasoning tokens:** 0
- **Total tokens:** 1,757,102

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


## Paris_1536

Facsimile is public domain.

https://euclides.huma-num.fr/hub/?datasetId=ds_ub81ja

## Paris_1615

OCR has been completed, but the facsimile is not copyrighted.

There is a new facsimile that has the appropriate copyright. I run the full OCR pipeline on it.

https://euclides.huma-num.fr/hub/?datasetId=ds_57vh5j&annotationId=ann_joneuv

Full correction with Codex Sol was run using dir mode.

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

# Curation Reviewed

## Nuremberg_1821

OCR was done on Google books facs:

https://euclides.huma-num.fr/hub/?datasetId=ds_tjokpg

Actual final version is on public domain facs:

https://euclides.huma-num.fr/hub/?datasetId=ds_yp36ia

Fable usage:

- Input tokens: 340
- Cache creation tokens: 583,576
- Cache read tokens: 1,697,451
- Output tokens: 165,781
- Total tokens: 2,447,148
- Total cost: $21.876591

Curation reviewed, no action item identified. Only an open Q: in the edition, the vocabulary is Lehrſatz and Aufgabe, not theorem and problem. In the curation index category name, do we want to preserve the editor's vocabulary or introduce ours?

## Rome_1574_vol1

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

Curation reviewed, no action item identified. 

To remember: Clavius mentions that the marginal Roman numerals refer to Campanus’s numbering immediately before the propositions of Book I:
> “Primum cuilibet propositioni duos numeros affiximus, quorum alter in margine depictus significat ordinem, quem Campanus ex traditione Arabum est secutus in Euclidis propositionibus…”
> (English: “First, we have attached two numbers to each proposition, one of which, printed in the margin, indicates the order that Campanus followed in Euclid’s propositions, according to the Arabic tradition…”)

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
