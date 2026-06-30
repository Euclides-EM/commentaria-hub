# Latest Larger-Corpus TPS Data

Main file:

`printed_missing_tps_v8_preview.csv`

This is the latest larger-corpus title-page extraction CSV. It has 7,169 non-empty feature rows for 650 editions. It came from a 690-key target run; 40 target keys produced no non-empty feature rows, and 2 additional printed keys lacked title-page transcriptions.

## CSV Columns

- `edition_id`: edition key, matching `key` in `ocrflow/store/items_metadata/items_print.csv`.
- `feature_id`: machine-readable feature identifier.
- `feature_name`: human-readable feature label.
- `source_id`: extraction run/source identifier.
- `source_revision`: feature prompt/revision identifier.
- `source_name`: source type, usually `llm`.
- `value`: extracted title-page text span.
- `properties`: optional structured metadata; usually empty in this CSV.

## Companion Files

- `printed_missing_tps_v8_keys.txt`: keys targeted by the run.
- `printed_missing_tps_without_transcription_keys.txt`: target keys skipped because no title-page transcription was available.
- `printed_tps_target_counts.txt`: quick accounting of target and completion counts.

## Feature Inventory

| feature_id | feature_name | rows | editions |
|---|---:|---:|---:|
| `action_verbs` | Verbs | 984 | 423 |
| `audience` | Intended Audience | 266 | 172 |
| `base_content` | Base Content | 388 | 384 |
| `bound_with` | Bound With | 314 | 111 |
| `bound_with_minimal` | Bound With - Minimal | 322 | 110 |
| `content_description` | Base Content Description | 299 | 264 |
| `date_in_imprint` | Date in Imprint | 559 | 559 |
| `dedicatee_name` | Patronage Dedication | 122 | 122 |
| `description_of_euclid` | Euclid Description | 20 | 20 |
| `destination_language` | Destination Language | 69 | 63 |
| `edition_details` | Edition Statement | 244 | 195 |
| `editor_description` | Adapter Description | 419 | 358 |
| `editor_name` | Adapter Attribution | 542 | 508 |
| `educational_authorities_references` | Other Educational Authorities | 152 | 91 |
| `elements_designation` | Elements Designation | 148 | 146 |
| `enriched_with` | Enriched With | 450 | 267 |
| `institutions` | Institutions | 220 | 173 |
| `location_in_imprint` | Place in Imprint | 768 | 555 |
| `origin_language` | Explicit Language References | 54 | 46 |
| `printing_privilege` | Publishing Privileges | 34 | 34 |
| `publisher_in_imprint` | Publisher in Imprint | 646 | 545 |
| `references_to_euclid` | Euclid References | 149 | 129 |

## Useful Joins

For bibliographic metadata, join:

- `edition_id` in this CSV
- to `key` in `ocrflow/store/items_metadata/items_print.csv`

For the currently reviewed genre-classification subset, join:

- `edition_id` in this CSV
- to `Page/Key` in `edition_classification_final_data/final_hybrid_classifications.csv`

Only part of the larger TPS extraction currently overlaps with the genre-classification table.
