import base64
import mimetypes
import os
import re
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed
from pathlib import Path

from openai import OpenAI
from tqdm import tqdm

MODEL = "gpt-5.2"
CONCURRENCY = 1
ORIGINAL_TEXT_TITLE = "Original text"
TRANSLATION_TITLE = "Translation"
SECTION_SEPARATOR = "---"
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


PROMPT = f"""
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

{ORIGINAL_TEXT_TITLE}:
{SECTION_SEPARATOR}
[transcription]
{SECTION_SEPARATOR}

{TRANSLATION_TITLE}:
{SECTION_SEPARATOR}
[English translation of the transcription, preserving paragraph/line structure where possible]
{SECTION_SEPARATOR}
""".strip()

OUTPUT_PATTERN = re.compile(
    rf"{re.escape(ORIGINAL_TEXT_TITLE)}:\s*{re.escape(SECTION_SEPARATOR)}\s*(?P<original>.*?)\s*{re.escape(SECTION_SEPARATOR)}\s*{re.escape(TRANSLATION_TITLE)}:\s*{re.escape(SECTION_SEPARATOR)}\s*(?P<translation>.*?)\s*{re.escape(SECTION_SEPARATOR)}\s*$",
    re.IGNORECASE | re.DOTALL,
)


def parse_llm_output(text: str):
    match = OUTPUT_PATTERN.search(text.strip())
    if not match:
        return None
    original = match.group("original").strip()
    translation = match.group("translation").strip()
    return original, translation


def output_paths_for_image(image_path: Path):
    raw_path = image_path.with_suffix(".txt")
    original_path = image_path.with_name(f"{image_path.stem}_original_llm.txt")
    translation_path = image_path.with_name(f"{image_path.stem}_modern_en.txt")
    error_path = image_path.with_name(f"{image_path.stem}_llm_error.txt")
    return raw_path, original_path, translation_path, error_path

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

    def process_image(image_path: Path):
        raw_path, original_path, translation_path, error_path = output_paths_for_image(image_path)
        if original_path.exists() and translation_path.exists():
            return None
        result = openai_query(PROMPT, str(image_path))
        if result is None:
            return None
        raw_path.write_text(result, encoding="utf-8")
        parsed = parse_llm_output(result)
        if parsed is None:
            error_path.write_text(
                "Could not match expected LLM output format.\n\nRaw output:\n\n" + result,
                encoding="utf-8",
            )
            return [raw_path, error_path]
        original_text, translation_text = parsed
        original_path.write_text(original_text, encoding="utf-8")
        translation_path.write_text(translation_text, encoding="utf-8")
        if error_path.exists():
            error_path.unlink()
        return [raw_path, original_path, translation_path]

    with ThreadPoolExecutor(max_workers=CONCURRENCY) as executor:
        futures = [executor.submit(process_image, image_path) for image_path in image_paths]
        for future in tqdm(as_completed(futures), total=len(futures), desc="Transcribing"):
            output_paths = future.result()
            if output_paths is not None:
                for output_path in output_paths:
                    print(f"Wrote {output_path}")


if __name__ == "__main__":
    main()
