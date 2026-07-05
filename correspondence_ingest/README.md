# Correspondence Ingest

Standalone Go command for extracting the correspondence index and letters table from page images with a vision-capable LLM.

## Layout

- `cmd/correspondence_ingest`: executable entrypoint
- `internal/app`: extraction, checkpointing, CSV rendering, status, and validation workflows
- `internal/llm`: OpenAI and Ollama provider integration
- `data/raw`: source images grouped by dataset and volume
- `data/outputs`: resumable source-of-truth manifests and explicitly built CSV exports
- `data/reviews`: AI validation checkpoints and human-readable reports

Run commands from this directory:

```sh
go run ./cmd/correspondence_ingest status
go run ./cmd/correspondence_ingest validate
go run ./cmd/correspondence_ingest extract --kind index
go run ./cmd/correspondence_ingest build-csv --kind index
```

## Manual corrections

Correct one manifest entry by selecting its source image and its 1-based
position in that page's `entries` array. The human identifier is required and
is stored with a timestamped old/new audit record:

```sh
go run ./cmd/correspondence_ingest manual-override --kind letters \
  --image vol_1/page.jpg --entry 1 --by mia \
  --letter-name "Claude Bredeau à Mersenne, 21 octobre 1617" \
  --reason "Checked against the photographed table"

go run ./cmd/correspondence_ingest manual-override --kind index \
  --image vol_1/page.jpg --entry 12 --by mia \
  --name "Abano (Pierre d')" --page-number 38 --is-bold=false
```

Index fields are `--name`, `--page-number`, `--reference`, and `--is-bold`.
Letters fields are `--letter-number`, `--letter-name`, and `--page-number`.
Pass `--reference=` or `--page-number=` to explicitly clear a field. Run
`build-csv` afterward to refresh the disposable CSV export.

AI commands load `.env` and `.env_private` from this directory. Ollama requires `OLLAMA_BASE_URL` and optionally `OLLAMA_AUTH_TOKEN`; OpenAI requires `OPENAI_API_KEY`.

See `AGENTS.md` for the complete extraction, resume, and validation workflow.
