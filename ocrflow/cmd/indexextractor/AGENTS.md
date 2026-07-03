# Index extractor agent guide

This command is the deterministic tool layer for future agent workflows. Keep extraction, validation, checkpointing, and persistence in Go; an agent may orchestrate these operations and investigate failures.

Run commands from the `ocrflow` directory:

```sh
go run ./cmd/indexextractor status
go run ./cmd/indexextractor extract
go run ./cmd/indexextractor validate
```

All commands accept `--kind index`, `--kind letters`, or `--kind all`. Extraction resumes by default from `<output>.manifest.json`. The manifest is the source of truth and the CSV is regenerated from it. Use `extract --rerun` only for an intentional clean rebuild of the selected kind.

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

Before considering a change complete:

1. Run `gofmt` on changed Go files.
2. Run `go test ./cmd/indexextractor` (use a workspace-writable `GOCACHE` if required).
3. Run `validate` against any generated CSV.
4. Do not edit extraction CSVs or manifests by hand during a run.

Each manifest page records its source path, volume, provider, model, extraction time, and structured entries. Future validation and review metadata should be added to the page record while preserving backward-compatible manifest versioning.
