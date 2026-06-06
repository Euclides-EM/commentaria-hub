# Edition V9 Analysis

This folder contains the review-ready analysis artifacts for the 692 printed, non-reprint editions classified with V9.

## Keepers

- `category_summary.csv`: category-level counts for the V9 run.
- `wide_classifications.csv`: one row per edition with the 20 category values as columns.
- `priority_review_queue.csv`: compact review worksheet for unknowns and non-EIP `Theoretical Mathematics` suspects. The two short interactive review batches from 2026-06-06 are recorded in `review_decision`, `corrected_value`, and `review_notes`.
- `accepted_eip_theoretical_overrides.csv`: accepted rule overrides for `origin_eip_csv => Theoretical Mathematics::primary`.

## Counts

- V9 run editions: 692.
- V9 run rows: 13,840.
- Malformed or missing-category rows: 0.
- `origin_eip_csv` overrides accepted: 72.
- Priority queue rows: 323.
- Priority queue rows already reviewed interactively: 20.
- Reprints omitted from V9 and SQL materialization: 69.

## Sanity Check

No broad run-level red flag was found. The large increase in `Theoretical Mathematics` is mostly explained by Euclid, Archimedes, algebra, quadrature, conics, logarithms, Descartes geometry, infinitesimals, and similar clearly theoretical mathematical material.

The recommended review path is to start with `priority_review_queue.csv`, especially `Practical Geometry`, `Theoretical Mathematics`, `Commercial Mathematics`, `Construction`, and the `Instrument Use` / `Instrument Construction` boundary.
