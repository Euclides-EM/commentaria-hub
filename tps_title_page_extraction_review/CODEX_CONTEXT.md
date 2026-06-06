# Codex Context: TPS Title-Page Extraction Final State

Use this file when returning to the TPS title-page extraction work.

## Workspace

- Repo root: `/Users/mia/dev/personal/elements-dh`
- Go project: `/Users/mia/dev/personal/elements-dh/ocrflow`
- Review folder: `/Users/mia/dev/personal/elements-dh/tps_title_page_extraction_review`

## Important Rule

Do not run Ollama-backed extraction directly unless Mia asks for it or has coordinated server availability. The Ollama server is shared/unstable, so future LLM runs should be commands for Mia to run.

## Final Data State

- Existing production/paper snapshot is preserved in `ann_1` by `ocrflow/internal/migrations/ocrflow/1774207517_tps_feature_results_opt.sql`.
- Final reviewed V8 results are inserted into `ann_tps_v8_reviewed` by `ocrflow/internal/migrations/ocrflow/1774300011_tps_title_page_v8_reviewed_results.sql`.
- Final latest V8 feature definitions/prompts are in `ocrflow/internal/migrations/ocrflow/1774300009_tps_title_page_latest_revisions.sql`.
- Source CSV for the V8 result migration: `tps_title_page_extraction_review/v8_results_preview_reviewed.csv`.

## Final V8 Run Summary

- Completed keys: 217/217.
- Reviewed CSV rows after manual cleanup: 5,573.
- Non-empty values after manual cleanup: 3,333.
- Features: 22.
- Two `Enriched With` rows were manually dropped after review:
  - `Amsterdam_1626`: purpose/audience phrase, not enrichment.
  - `Rotterdam_1661`: purpose phrase, not enrichment.

## Final Feature Set

The offline runner embeds the current final feature registry in `ocrflow/cmd/title-page-extraction-offline/features.go`.

Default features are:

- `action_verbs`
- `audience`
- `base_content`
- `bound_with`
- `bound_with_minimal`
- `content_description`
- `date_in_imprint`
- `dedicatee_name`
- `description_of_euclid`
- `destination_language`
- `edition_details`
- `editor_description`
- `editor_name`
- `educational_authorities_references`
- `elements_designation`
- `enriched_with`
- `institutions`
- `location_in_imprint`
- `origin_language`
- `printing_privilege`
- `publisher_in_imprint`
- `references_to_euclid`

## Key Policy Decisions

- `Base Content`: minimal title nucleus of the core Euclidean work.
- `Elements Designation`: printed Elements/book designation; can overlap with Base Content.
- `Base Content Description`: descriptions of the core Elements/content only, not enrichments or Euclid/person descriptors.
- `Enriched With`: additions/enrichments integrated with the main text; do not include purpose/audience phrases unless grammatically part of the enrichment object.
- `Bound With`: fuller descriptions of physically bound additional works.
- `Bound With - Minimal`: minimal title units for physically bound additional works.
- `Verbs`: bibliographic/edition action verbs; exclude ordinary mathematical-operation verbs when they only describe examples inside an enrichment.
- Imprint features are default: date, place/address, publisher/printer.

## Offline Runner

The production runner is now only:

```sh
cd /Users/mia/dev/personal/elements-dh/ocrflow

go run ./cmd/title-page-extraction-offline \
  -output-csv ../tps_title_page_extraction_review/v8_offline_results_preview.csv
```

Resume is on by default and uses `<output-csv>.done`. The command accepts `-keys`, `-keys-file`, `-features`, `-checkpoint-file`, and `-resume=false`.

## Cleanup Notes

Removed intermediate TPS migration files:

- `1774300006_tps_features_improvements.sql`
- `1774300007_tps_v6_feature_revisions.sql`
- `1774300009_tps_v7_targeted_feature_revisions.sql`
- `1774300011_tps_restore_core_content_features.sql`
- `1774300012_tps_v8_policy_revisions.sql`
- `1774300013_tps_v8_tighten_action_verbs.sql`
- `1774300014_tps_v8_tighten_content_enrichment_boundaries.sql`

Removed the DB-backed `ocrflow/cmd/title-page-extraction` command. Future long extraction work should use `ocrflow/cmd/title-page-extraction-offline`.
