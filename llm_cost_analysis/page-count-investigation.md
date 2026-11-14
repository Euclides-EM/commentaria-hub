# Provider page-count investigation (2026-11-14)

Read the main CSV together with page-count-evidence.csv: its Count status distinguishes reported counts from estimates. The main CSV preserves the requested columns. Costs and cost-per-facsimile-page estimates have not changed.

Code findings:

- ocrflow/pkg/transcriptioncorrector/corrector.go: DefaultRounds=1. The page_by_page loop makes one successful provider call per non-skipped page per round, then increments requestCount. A failed call returns before that increment. Therefore these counters are not simply retry counts. Across rounds/reruns a page can still be counted more than once.
- The logged pages value is the selected image count, not necessarily completed-page count on a failed/interrupted run.
- Directory mode invokes one workspace call for all selected pages and validates every output. Request counts cannot be used as page counts in that mode.
- ocrflow/internal/service/annotation_rule_applier.go: API invocation uses annotation page ranges. ocrflow/internal/service/annotation.go: applied_rules is appended only after rule execution succeeds. CLI invocation does not update this history.

API findings:

All 28 datasets' /annotations responses were checked. None contains llm_transcription_corrector in applied_rules. The returned page ranges describe segmentation/OCR annotation scope; they do not identify which provider corrected each page. This does not prove the corrector was never run: CLI runs do not write these rules and unsuccessful API runs also do not append them. Public /jobs returned null. /logs requires authentication (HTTP 401), so historical completion logs were unavailable through that endpoint. No local corrector log files were found in the repository.

Inference policy:

- Preserve explicit page counts for Glasgow, Rome vol1, Paris_1667, and the logged second run of Lyon_1557.
- For other page-by-page runs, use recorded requests as an estimate when they do not exceed facsimile pages, assuming one round and no repeated pages across runs. Providers may overlap; sums need not equal facsimile pages. These are estimates, not recovered unique-page evidence.
- For documented complete single-provider corrections/transcriptions without counts, estimate from full facsimile pages. This includes Paris_1615 (directory mode). Completion can predate a replacement facsimile, so historical page counts may differ.
- Lyon_1672 Fable remains blank: 790 requests over 449 pages cannot establish unique-page coverage.
- Lyon_1557's 79 is only its logged second-run coverage, matching its partial cost. Rome's Fable cost and Paris_1794's OpenAI cost are also partial.

The previous requirement that request totals exactly equal facsimile pages was unnecessarily restrictive and has been removed.
