# Index extractor agent guide

This command extracts two datasets from page images with a vision-capable LLM:

- `index`: names, individual page references or cross-references, bold state, and volume
- `letters`: letter number, correspondent name, page reference, and volume

Keep discovery, extraction rules, validation, checkpointing, and persistence in Go. Agents may orchestrate commands and investigate flagged pages, but must preserve the manifest-first workflow.

Run commands from the `ocrflow` directory:

```sh
go run ./cmd/indexextractor status
go run ./cmd/indexextractor extract
go run ./cmd/indexextractor validate
```

The default command is `extract`; prefer spelling the command out in agent workflows. All commands accept `--kind index`, `--kind letters`, or `--kind all` (the default). Default paths are relative to `ocrflow`, which is why commands must run there.

Raw images live under `cmd/indexextractor/data/raw/{index,letters_table}/<volume>/`. Supported extensions are `.jpg`, `.jpeg`, `.png`, and `.webp`, case-insensitively. Every image must be beneath a volume directory; that first relative directory component becomes the CSV `volume` value.

Extraction resumes by default from `<output>.manifest.json`. The manifest is the source of truth, is checkpointed after every successfully processed image, and deterministically regenerates the CSV. A page already present in the manifest is considered complete. Use `extract --rerun` only for an intentional clean rebuild of the selected kind; it discards the selected in-memory manifest state and rewrites its manifest and CSV before extraction begins.

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

The defaults are Ollama with `qwen3.6:35b`. Override them with `--ai-provider openai|ollama` and `--ai-model <model>`. The model must support image input. The process loads `.env` and `.env_private` from the working directory:

- OpenAI extraction requires `OPENAI_API_KEY`.
- Ollama extraction requires `OLLAMA_BASE_URL`; authenticated endpoints also use `OLLAMA_AUTH_TOKEN`.
- `status` and `validate` do not create an LLM client and do not require provider credentials.

Use `--index-dir`, `--letters-dir`, `--index-output`, and `--letters-output` only when intentionally working with alternate fixtures or outputs. When `--kind all` is used, the two output paths must differ.

## Failure and review semantics

LLM output must be strict JSON with no unknown fields. For index pages, a structurally invalid response is attempted twice before the command stops. A structurally valid response may contain invalid rows: those rows are skipped, recorded in the page's `issues`, and printed as warnings. Letters-table extraction also skips and records invalid rows, but currently does not retry a structurally invalid response.

Single-letter index section headings with no page number or reference are ignored. Every retained index row must have a name and exactly one of `page_number` or `reference`; cross-references must have `is_bold=false`. Every retained letters row must have all three fields populated.

Treat tolerated issues as review work, not as a clean extraction. Both `status` and `validate` report persisted issues by image. `status` compares discovered images with manifest pages and reports completed/pending counts. `validate` checks CSV shape and content against the manifest, but it does not prove that every raw image has been processed; run `status` as well before declaring a dataset complete.

Before considering a change complete:

1. Run `gofmt` on changed Go files.
2. Run `go test ./cmd/indexextractor` (use a workspace-writable `GOCACHE` if required).
3. Run `status` for each affected kind and investigate pending images or tolerated issues.
4. Run `validate` against each generated CSV/manifest pair.
5. Do not edit extraction CSVs or manifests by hand. Rerun the affected page so provenance, issues, and rendered CSV remain synchronized.

Each manifest page records its source path, volume, provider, model, UTC extraction time, structured entries, and optional parsing issues. Loading rejects unknown manifest fields and metadata whose version or kind does not match. If the manifest schema changes, update validation and rendering together and deliberately version or migrate the format; do not silently reinterpret existing manifests.
