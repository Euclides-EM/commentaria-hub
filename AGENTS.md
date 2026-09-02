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
