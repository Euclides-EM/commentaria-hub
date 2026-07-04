# Correspondence Ingest

Standalone Go command for extracting the correspondence index and letters table from page images with a vision-capable LLM.

## Layout

- `cmd/correspondence_ingest`: executable entrypoint
- `internal/app`: extraction, checkpointing, CSV rendering, status, and validation workflows
- `internal/llm`: OpenAI and Ollama provider integration
- `data/raw`: source images grouped by dataset and volume
- `data/outputs`: deterministic CSV exports, resumable manifests, and archived exports
- `data/reviews`: AI validation checkpoints and human-readable reports

Run commands from this directory:

```sh
go run ./cmd/correspondence_ingest status
go run ./cmd/correspondence_ingest validate
go run ./cmd/correspondence_ingest extract --kind index
```

AI commands load `.env` and `.env_private` from this directory. Ollama requires `OLLAMA_BASE_URL` and optionally `OLLAMA_AUTH_TOKEN`; OpenAI requires `OPENAI_API_KEY`.

See `AGENTS.md` for the complete extraction, resume, and validation workflow.
