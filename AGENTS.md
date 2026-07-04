## HumaNum Ollama

When requests to the HumaNum Ollama server return HTTP 502:

- Do not immediately assume the server is unavailable.
- First consider that the model exceeded an upstream timeout.
- Large prompts or tasks that require long reasoning frequently trigger this behavior.
- Before retrying repeatedly, try reducing prompt size, splitting the task into smaller steps, or requesting a shorter output.
- Waiting for the server to "recover" is usually not sufficient if the prompt itself is the cause.
