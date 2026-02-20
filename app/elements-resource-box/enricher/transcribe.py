import base64
import json

from tqdm import tqdm

from tools import openai_client, write_csv, read_csv, strip_surrounding_quotes

#file_path = "../public/docs/EiP.csv"
file_path = "../public/docs/EiP-secondary.csv"

_EXTRACT_FEATURES = False
_MERGE_FEATURES = True

entries, fieldnames = read_csv(file_path)
for field in [
    "num_of_types",
]:
    if field not in fieldnames:
        fieldnames.append(field)

for i in tqdm(range(len(entries)), desc="Processing entries"):
    entry = entries[i]
    if not entry["tp_url"]:
        continue
    out_path = f'out/{entry["key"].replace("/", "_")}.json'
    if _EXTRACT_FEATURES:
        try:
            response = openai_client.chat.completions.create(
                model="gpt-4o",
                max_tokens=None,
                temperature=0,
                messages=[
                    {
                        "role": "user",
                        "content": [
                            {
                                "type": "text",
                                "text": """
You are an AI agent designed to analyze historical title pages of books, such as translations of Euclid’s Elements.

You will be given a scanned image of a printed title page from the 16th–18th century.

Your task is to count how many distinct **type sets** appear on the page. A new type set should be counted when:
- The typeface (e.g. Roman, Gothic, Italic, Blackletter) changes.
- The **size** changes within the same typeface.

Ignore ink quality, wear, or minor variations due to print artifacts. Do not include seals, borders, or handwriting.

Return only a valid JSON in the format:
{
    "num_of_types": <integer>
}
Do not include any additional commentary or explanation.
                    """.strip()
                            },
                            {
                                "type": "image_url",
                                "image_url": {
                                    "url": entry["tp_url"],
                                },
                            },
                        ],
                    }
                ],
            )
            if len(response.choices) != 1:
                print("!!! Unexpected response from OpenAI", response)
            if response.choices[0].finish_reason != 'stop':
                print("!!! OpenAI did not finish processing the request", response)
            with open(out_path, "w") as f:
                f.write(response.choices[0].message.content.strip())
        except Exception as e:
            print("!!! Error querying OpenAI", e)

    if _MERGE_FEATURES:
        with open(out_path, "r") as f:
            features = f.read().replace("```json", "").replace("```", "").strip()
        try:
            features_dict = json.loads(features)
            for key, value in features_dict.items():
                entries[i][key] = (
                    ", ".join([strip_surrounding_quotes(v) for v in value])
                    if isinstance(value, list)
                    else strip_surrounding_quotes(value)
                    if isinstance(value, str)
                    else value)
        except Exception as e:
            print(f"Error processing features for {entry['key']}: {e}")
            continue

if _MERGE_FEATURES:
    write_csv(entries, file_path, fieldnames)
