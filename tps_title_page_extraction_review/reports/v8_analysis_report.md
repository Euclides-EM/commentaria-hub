# V8 Title-Page Extraction Analysis

## Run Status

- Completed checkpoint keys: 217/217.
- CSV rows: 5,575.
- Distinct features: 22.
- Non-empty rows: 3,335 (V7: 2,291; delta: +1,044).
- Hard grounding omissions in log: 22.
- Fuzzy-grounded source-spelling rescues: 33.

## V7 to V8 Shape

- V7 had 4,360 rows, 17 features, and 2,291 non-empty rows.
- V8 has 5,575 rows, 22 features, and 3,335 non-empty rows.
- Restored/default-added feature families present in V8: `content_description`, `elements_designation`, `date_in_imprint`, `location_in_imprint`, `publisher_in_imprint`.

## Policy Comparison Rows

Compared 72 targeted review rows from `v7_policy_review.csv` against V8 output.

- exact_set: 20
- target_contains: 18
- span_contains: 13
- different: 9
- partial_overlap: 8
- missing: 4

Policy-aligned or usable scores (`exact_set`, `span_contains`, `target_contains`, `both_empty`): 51/72 (70.8%).

## Remaining Watch Items

- Possible audience/purpose leakage into `Enriched With`: 2 rows.
- Math-operation verbs in `Verbs`: 0 rows.
- Hard omissions are concentrated in: {'action_verbs': 10, 'location_in_imprint': 9, 'content_description': 1, 'bound_with_minimal': 1, 'enriched_with': 1}.
- Fuzzy rescues are concentrated in: {'content_description': 5, 'editor_description': 4, 'bound_with': 3, 'action_verbs': 3, 'bound_with_minimal': 2, 'enriched_with': 2, 'dedicatee_name': 2, 'institutions': 2, 'edition_details': 2, 'base_content': 2, 'elements_designation': 2, 'description_of_euclid': 1, 'editor_name': 1, 'educational_authorities_references': 1, 'audience': 1}.

## Output Files

- `v8_feature_summary.csv`
- `v8_policy_comparison.csv`
- `v8_suspicious_rows.csv`
