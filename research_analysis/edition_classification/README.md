# Edition Classification

## Conclusion

The retained table is the final reviewed/hybrid subject classification for 155 editions across 20 subject categories (3,100 rows). It preserves the selected value and compact provenance needed to audit it.

## Evidence

- [`data/edition_subject_classifications_reviewed.csv`](data/edition_subject_classifications_reviewed.csv): reviewed values, selected source, reason, concern flags, review notes, and earlier candidate values.

Row grain: one edition (`Page/Key`) × one subject (`Classification`). The composite key is unique. `final_value` is the selected decision; `final_source` and `final_reason` explain its provenance. The `Orig_Value` and `llm_Value_6` through `llm_Value_8` columns are retained audit evidence, not competing current answers.

`data/edition_classification_keys.txt` freezes the 155 editions included in this analysis.

The database installer is `ocrflow/internal/migrations/ocrflow/1774300010_edition_classification_final_hybrid_opt.sql`. New editions should use the V9 edition-classification prompt through `go run ./cmd/edition-classification` from `ocrflow/`.

No intermediate review runs are retained here.
