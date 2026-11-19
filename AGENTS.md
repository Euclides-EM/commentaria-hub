## HumaNum Ollama

When requests to the HumaNum Ollama server return HTTP 502:

- Do not immediately assume the server is unavailable.
- First consider that the model exceeded an upstream timeout.
- Large prompts or tasks that require long reasoning frequently trigger this behavior.
- Before retrying repeatedly, try reducing prompt size, splitting the task into smaller steps, or requesting a shorter output.
- Waiting for the server to "recover" is usually not sufficient if the prompt itself is the cause.

## Transcription Markdown dialect

- `docs/MARKDOWN_DIALECT.md` is the canonical transcription Markdown specification.
- Before creating, editing, validating, or interpreting transcription Markdown—or code and tests that process it—read and follow that specification. Do not load it for unrelated tasks.
- After changing the specification, run `go generate ./pkg/transcriptioncorrector` from `ocrflow/` and verify `go test ./pkg/transcriptioncorrector` so the LLM transcription corrector's embedded copy stays synchronized.

## Docs overview

Read the matching doc before working on its topic; skip the rest.

- `docs/MARKDOWN_DIALECT.md` — transcription Markdown spec (see above).
- `docs/skills/edition-index-curation/SKILL.md` — report-only curation of transcription headers: sequence gaps, header levels, OCR misreads, printer's errors, case and spacing in headers (THEOR./PROB./PROP./CHAP. and the rest), with or without a generated `index.json`. Follow it for any header curation task; it suggests fixes and never edits transcriptions.
- `docs/TRANSCRIPTION_STATUS.md` — per-book OCR/transcription pipeline status and queue.
- `docs/OCR_MODEL_GUIDE.md` — which Kraken OCR models to try per language/era.
- `docs/TEI_SPEC.md` — design of the OCR → NER → TEI export pipeline.
- `docs/EXTERNAL_RESOURCES_METADATA_ENRICHMENT.md` — external catalogues and authority databases for metadata enrichment.
- `docs/HUB_DEPLOYMENT.md` — deploying the hub on the Huma-Num server (SSH deploy keys).
- `docs/ESCRIPTORIUM_DEPLOYMENT.md` — setting up and deploying eScriptorium.
- `docs/GOOGLE_DRIVE_INTEGRATION.md` — Google Drive inbox for facsimile PDFs and diagram crops, and Drive backups.
- `docs/FACSIMILE_MEDIA_BACKUP.md` — backing up facsimile PDFs and diagram crops with rclone.
- `docs/GPU_FARM.md` — running Python jobs on GPU farms (qsub/sbatch/systemd).
- `docs/GITHUB_TOKEN_SETUP.md` — creating a personal GitHub token for a team member.
- `docs/OLD_DEPLOYMENT.md`, `docs/OLD_WEB_APP_DEPLOYMENT.md` — archival only; do not follow.
