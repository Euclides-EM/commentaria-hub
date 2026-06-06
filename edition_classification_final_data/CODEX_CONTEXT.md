# Codex Context: Edition Classification Final Data

This folder is intentionally minimal for production.

## Keep

- `final_hybrid_classifications.csv`: final reviewed/hybrid classifications for the 155-edition comparison batch.
- `README.md`: human-facing instructions.
- `CODEX_CONTEXT.md`: this future-session context.

Everything else from the review cycle was removed from the folder to avoid committing exploratory analysis and intermediate run outputs.

## Final CSV

`final_hybrid_classifications.csv` has 3,100 rows: 155 editions * 20 subject categories.

It preserves compact provenance:

- final selected value;
- chosen source;
- reason/concern flags;
- review notes where Mia manually filled a value;
- original, V6, V7, and V8 values.

## Migration

Use:

`ocrflow/internal/migrations/ocrflow/1774300010_edition_classification_final_hybrid_opt.sql`

The optional migration installs:

- `m_classifier`;
- V6, V7, V8, and V9 prompt revisions;
- final hybrid result rows for the 155 existing editions.

The final result source revision is:

`final_hybrid_v1_2026_06_03`

## Future/New Editions

For new editions, run the latest prepared single prompt, V9:

`6a0d47e3-f472-4b63-a6f5-67c693a0adf9`

The runner is:

```sh
cd ocrflow
go run ./cmd/edition-classification \
  -keys-file path/to/new_keys.txt \
  -output-csv path/to/new_results_preview.csv
```

Add `-dry-run` for previews. Omit it when writing to DB.

Spot-check `Commercial Mathematics`, `Construction`, `Practical Geometry`, `Instrument Use`, `Instrument Construction`, and any `unknown` value first.
