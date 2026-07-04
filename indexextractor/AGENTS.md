# Index extractor agent guide

This command extracts two datasets from page images with a vision-capable LLM:

- `index`: names, individual page references or cross-references, bold state, and volume
- `letters`: letter number, correspondent name, page reference, and volume

Keep discovery, extraction rules, validation, checkpointing, and persistence in Go. Agents may orchestrate commands and investigate flagged pages, but must preserve the manifest-first workflow.

Run commands from the standalone `indexextractor` directory:

```sh
go run ./cmd/indexextractor status
go run ./cmd/indexextractor extract
go run ./cmd/indexextractor validate
go run ./cmd/indexextractor validate-transcriptions
go run ./cmd/indexextractor validate-parsing
```

Extraction defaults to the original single vision call (`--extraction-mode one-pass`). Use `--extraction-mode two-pass` to first create an auditable text transcription from the image and then make a text-only LLM call that converts that transcription to the required JSON:

```sh
go run ./cmd/indexextractor extract --kind index --extraction-mode two-pass
```

In two-pass mode, only the transcription call receives the image attachment. The structured extraction call receives the transcription in its prompt and no attachment. The transcription is checkpointed to the manifest immediately after pass one. If pass two fails, that transcription-only page remains pending; the next two-pass run reuses it and starts at pass two. Rerun a page (or use `--rerun`) when changing modes.

The default command is `extract`; prefer spelling the command out in agent workflows. All commands accept `--kind index`, `--kind letters`, or `--kind all` (the default). Default paths are relative to the `indexextractor` project root, which is why commands must run there.

## AI validation and review files

The two AI validation commands operate on completed two-pass manifest pages because both checks require the checkpointed transcription:

```sh
go run ./cmd/indexextractor validate-transcriptions --kind index \
  --ai-provider ollama --ai-model qwen3.6:35b
go run ./cmd/indexextractor validate-parsing --kind index \
  --ai-provider openai --ai-model gpt-5-mini
```

`validate-transcriptions` sends the transcription and its underlying image to the selected vision-capable provider/model. It checks exact text, numbers, punctuation, ordering, completeness, and meaningful bold markup. Its differences normally require a human to inspect the image.

`validate-parsing` is text-only: it sends the checkpointed transcription and that page's structured manifest entries, with no image attachment. Its prompt treats the transcription as the source of truth and asks a meticulous machine reviewer to find omissions, duplicates, grouping mistakes, cross-reference errors, page-number errors, name errors, and bold-flag errors. This review is deliberately suitable for Codex to investigate and act on without visually reading every image; do not automatically edit the CSV or manifest from a validator claim. Verify the reported source entry and rerun extraction for the affected page.

For each dataset and check, the command writes two files under `data/reviews/` when using the default outputs:

- `<output>.validate-<check>.json` is machine state and the resume checkpoint.
- `<output>.validate-<check>.md` is the ordered human review, with accurate pages, failures, and actionable expected/actual differences. For example, `data/reviews/index.csv.validate-parsing.md`.

Both files are atomically rewritten after every attempted page. Normal reruns skip successful validations whose source fingerprint and validator provider/model are unchanged, retry failures, and automatically revalidate pages whose transcription, structured entries, or validator changed. Stop the process at any time and run the same command to resume. `--skip-failures` also skips recorded validation failures. `--rerun` clears the selected validation checkpoint; `--rerun-images` revalidates only named pages and requires one explicit kind. Provider errors are recorded per page and do not erase earlier results.

When the provider returns a response that cannot be parsed or fails validation schema checks, the failed result stores that response as `raw_response` and includes it in the Markdown review. Successful responses and provider errors do not store this field.

Prefer one kind per long agent run. First run `status` and ordinary `validate`, then run transcription validation, then parsing validation. Review the generated Markdown after each resumable batch. A practical agent instruction is:

> From the `indexextractor` directory, run `status --kind <kind>` and `validate --kind <kind>`, then resume `validate-transcriptions --kind <kind>` and `validate-parsing --kind <kind>` with the requested providers/models. Do not delete checkpoints. Inspect the parsing-validation Markdown closely: verify every `REVIEW REQUIRED` item against the manifest transcription and parsed entries, group findings by page, identify false positives, and recommend exact `extract --rerun-images` commands for genuine errors. Do not hand-edit generated CSV or manifest files. Report failures separately. If HumaNum Ollama returns HTTP 502, assume prompt/model timeout first; reduce scope to one kind or targeted pages and avoid blind repeated retries.

Raw images live under `data/raw/{index,letters_table}/<volume>/`. Supported extensions are `.jpg`, `.jpeg`, `.png`, and `.webp`, case-insensitively. Every image must be beneath a volume directory; that first relative directory component becomes the CSV `volume` value.

Extraction resumes by default from `<output>.manifest.json`. The manifest is the source of truth, is checkpointed after pass one and every successfully processed image, and deterministically regenerates the CSV. A page is complete only when it has structured entries (an empty but non-null entries array is a valid completed result); a transcription-only page is pending. Use `extract --rerun` only for an intentional clean rebuild of the selected kind; it discards the selected in-memory manifest state and rewrites its manifest and CSV before extraction begins.

For targeted work, prefer a single kind:

```sh
go run ./cmd/indexextractor status --kind index
go run ./cmd/indexextractor extract --kind index
go run ./cmd/indexextractor validate --kind index
```

Rerun one or more pages by using an unambiguous filename or path. Existing page results are replaced in the manifest, then the CSV is regenerated without duplicates:

```sh
go run ./cmd/indexextractor extract --kind index --rerun-images vol_1/page01.jpg,vol_1/page02.jpg
```

`--rerun-images` requires one explicit kind and cannot be combined with `--rerun`. Selectors may be a full path, a unique basename, or a unique path suffix; ambiguous selectors fail. Targeted results replace the matching page records, and the CSV is regenerated without duplicate page records.

## Providers and configuration

The defaults are Ollama with `qwen3.6:35b`, except that `--second-pass-ai-model` defaults to `gpt-oss:120b`. `--ai-provider` and `--ai-model` override the general extraction defaults. In two-pass mode, `--first-pass-ai-provider`/`--first-pass-ai-model` and `--second-pass-ai-provider`/`--second-pass-ai-model` take precedence for their respective passes. Only the effective first-pass model must support image input. For example, use free vision transcription followed by more reliable structured extraction:

```sh
go run ./cmd/indexextractor extract --kind index --extraction-mode two-pass \
  --first-pass-ai-provider ollama --first-pass-ai-model qwen3.6:35b \
  --second-pass-ai-provider openai --second-pass-ai-model gpt-5-mini
```

Provider selection is cost-sensitive. HumaNum Ollama is free through the workplace, but it is slow, its advertised model list includes models that may not work, and long requests commonly surface upstream timeouts as HTTP 502. Treat a 502 as a likely prompt/model timeout first: reduce the prompt or output, split the work, or reuse a checkpointed first pass instead of repeatedly waiting and retrying the same request. The personal OpenAI API is paid, but is faster and more reliable and returns clearer errors. Reserve it for complex or accuracy-sensitive work, or when time pressure justifies the cost; a useful default is Ollama for transcription and OpenAI only for the harder structured pass.

The process loads `.env` and `.env_private` from the working directory:

- OpenAI extraction requires `OPENAI_API_KEY`.
- Ollama extraction requires `OLLAMA_BASE_URL`; authenticated endpoints also use `OLLAMA_AUTH_TOKEN`.
- `status` and `validate` do not create an LLM client and do not require provider credentials.

Use `--index-dir`, `--letters-dir`, `--index-output`, and `--letters-output` only when intentionally working with alternate fixtures or outputs. When `--kind all` is used, the two output paths must differ.

## Failure and review semantics

LLM output must be strict JSON with no unknown fields. For index pages, a structurally invalid response is attempted twice before the page is marked failed. A structurally valid response may contain invalid rows: those rows are skipped, recorded in the page's `issues`, and printed as warnings. Letters-table extraction also skips and records invalid rows, but currently does not retry a structurally invalid response.

An LLM or response-parsing failure does not stop the batch. It is stored on the manifest page with the failed phase (`single-pass`, `first-pass`, or `second-pass`), effective provider/model, UTC failure time, and error, then extraction continues with the next image. Second-pass failures retain the checkpointed transcription, so the default retry behavior starts directly at pass two. First-pass failures have no transcription and retry pass one. Failed pages are retried by default on the next extraction; pass `--skip-failures` to skip pages whose latest attempt failed. `status` and `validate` both list persisted failures.

Single-letter index section headings with no page number or reference are ignored. Every retained index row must have a name and exactly one of `page_number` or `reference`; cross-references must have `is_bold=false`. Every retained letters row must have all three fields populated.

Treat tolerated issues as review work, not as a clean extraction. Both `status` and `validate` report persisted issues by image. `status` compares discovered images with manifest pages and reports completed/pending counts. `validate` checks CSV shape and content against the manifest, but it does not prove that every raw image has been processed; run `status` as well before declaring a dataset complete.

Before considering a change complete:

1. Run `gofmt` on changed Go files.
2. Run `go test ./...` (use a workspace-writable `GOCACHE` if required).
3. Run `status` for each affected kind and investigate pending images or tolerated issues.
4. Run `validate` against each generated CSV/manifest pair.
5. Do not edit extraction CSVs or manifests by hand. Rerun the affected page so provenance, issues, and rendered CSV remain synchronized.

Each manifest page records its source path, volume, extraction mode, structured-pass provider/model/time, structured entries, optional parsing issues, and (for two-pass extraction) the transcription plus its provider/model/time. Failed pages additionally record failure provenance. Loading rejects unknown manifest fields and metadata whose version or kind does not match. Manifest v1 is migrated as legacy `one-pass` data, v2 as completed two-pass data, and v3 as resumable transcription data; newly saved manifests use v4, which adds failure state. If the manifest schema changes, update validation and rendering together and deliberately version or migrate the format; do not silently reinterpret existing manifests.
