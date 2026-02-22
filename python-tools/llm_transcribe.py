import base64
import mimetypes
import os
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

from openai import OpenAI
from tqdm import tqdm

MODEL = "gpt-5.2"
CONCURRENCY = 1
IMAGE_EXTENSIONS = {
    ".jpg",
    ".jpeg",
    ".png",
    ".gif",
    ".webp",
    ".bmp",
    ".tif",
    ".tiff",
}

openai_api_key = os.environ.get("OPENAI_API_KEY")
if not openai_api_key:
    raise ValueError("OPENAI_API_KEY environment variable is not set. Please set it to your OpenAI API key.")

openai_client = OpenAI(
    api_key=openai_api_key
)


def openai_query(prompt, image_path, creativity=0):
    mime_type, _ = mimetypes.guess_type(image_path)
    if not mime_type or not mime_type.startswith("image/"):
        mime_type = "image/jpeg"
    with open(image_path, "rb") as f:
        image_data = base64.b64encode(f.read()).decode("utf-8")
    try:
        response = openai_client.chat.completions.create(
            model=MODEL,
            messages=[
                {"role": "user", "content": [
                    {"type": "text", "text": prompt},
                    {"type": "image_url", "image_url": {
                        "url": f"data:{mime_type};base64,{image_data}"
                    }}
                ]}
            ],
            temperature=creativity
        )
        if len(response.choices) != 1:
            print("!!! Unexpected response from OpenAI", response)
        if response.choices[0].finish_reason != 'stop':
            print("!!! OpenAI did not finish processing the request", response)
        return response.choices[0].message.content
    except Exception as e:
        print("!!! Error querying OpenAI", e)


PROMPT = """
You will be given a single image of a page from Euclid’s *Elements* (Book X).

## Task
Produce a **diplomatic transcription** of all visible text, preserving layout as closely as possible. Maintain:
- Line breaks
- Indentation
- Capitalization
- Punctuation
- Hyphenation
- Spelling exactly as printed

**Do not correct, normalize, modernize, or interpret the text in any way.**

## Include
- Page headers and footers
- Page numbers
- Marginal notes
- Proposition titles and labels
- Footnotes
- Catchwords or signatures (if present)
- Any letter labels printed in diagrams (e.g., A, B, C)

## Exclude
- Descriptions or interpretations of diagrams beyond transcribing their printed labels

## Illegible or Uncertain Text
- If text is unreadable, write: `[illegible]`
- If text is unclear but you want to provide a best guess, write: `[uncertain: <best guess>]`

## Translation Rule
Include also an English translation of the full text. It should also follow original formatting. Clearly separate the original text from the translation using the following format:

Original text:
---
[transcription]
---

Translation:
---
[English translation of the transcription, preserving paragraph/line structure where possible]
---
""".strip()

def main():
    if len(sys.argv) != 2:
        print("Missing inputs directory")
        raise SystemExit(1)

    image_dir = Path(sys.argv[1]).expanduser().resolve()
    if not image_dir.is_dir():
        print(f"Not a directory: {image_dir}")
        raise SystemExit(1)

    image_paths = sorted(
        path
        for path in image_dir.rglob("*")
        if path.is_file() and path.suffix.lower() in IMAGE_EXTENSIONS
    )

    print(f"Found {len(image_paths)} images")
    if len(image_paths) == 0:
        raise SystemExit(0)

    image_paths = image_paths[:1]

    def process_image(image_path: Path):
        output_path = image_path.with_suffix(".txt")
        if output_path.exists():
            return None
        result = openai_query(PROMPT, str(image_path))
        if result is None:
            return None
        output_path.write_text(result, encoding="utf-8")
        return output_path

    with ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        futures = [executor.submit(process_image, image_path) for image_path in image_paths]
        for future in tqdm(as_completed(futures), total=len(futures), desc="Transcribing"):
            output_path = future.result()
            if output_path is not None:
                print(f"Wrote {output_path}")


if __name__ == "__main__":
    main()
