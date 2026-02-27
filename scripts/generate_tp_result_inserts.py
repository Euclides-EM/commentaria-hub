import csv
from pathlib import Path
import uuid


DATASET_ID = "tps"
ANNOTATION_ID = "ann_1"
SOURCE_RESP = "csv_import"

features_to_import = {
    "base_content",
    "content_description",
    "destination_language",
    "editor_name",
    "editor_description",
    "dedicatee_name",
    "edition_details",
    "printing_privilege",
    "action_verbs",
    "origin_language",
    "educational_authorities_references",
    "references_to_euclid",
    "description_of_euclid",
    "audience",
    "elements_designation",
    "greek_text",
    "institutions",
    "bound_with",
    "enriched_with",
    "date_in_imprint",
    "publisher_in_imprint",
    "location_in_imprint",
    "printing_privilege_in_imprint",
    "dedication_in_imprint",
    "editor_in_imprint",
    "editor_description_in_imprint",
}

def gen_random_uuid():
    return str(uuid.uuid4())


def split_values(value: str):
    if not value:
        return []

    if "::" in value:
        parts = value.split("::")
    elif "," in value:
        parts = value.split(",")
    else:
        return [value.strip()]

    return [p.strip() for p in parts if p.strip()]


def sql_escape(s: str) -> str:
    return s.replace("'", "''")


def generate_sql(csv_path: Path, output_path: Path):
    feature_results_rows = []
    feature_values_rows = []

    with open(csv_path, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)

        for row in reader:
            page_key = row["key"].strip()

            for feature, value in row.items():
                if feature not in features_to_import:
                    continue

                clean_value = value.strip()
                if not clean_value or clean_value.lower() in {"none", "null", "n/a"}:
                    continue

                values = split_values(value)

                feature_results_rows.append(
                    f"""('{gen_random_uuid()}', '{DATASET_ID}', '{ANNOTATION_ID}', '{sql_escape(feature)}', '{sql_escape(page_key)}', '{SOURCE_RESP}')"""
                )

                for idx, val in enumerate(values):
                    feature_values_rows.append(
                        f"""('{DATASET_ID}', '{ANNOTATION_ID}', '{sql_escape(feature)}', '{sql_escape(page_key)}', {idx}, '{sql_escape(val)}', '{{}}')"""
                    )

    with open(output_path, "w", encoding="utf-8") as out:
        if feature_results_rows:
            out.write(
                f"""INSERT INTO feature_results (id, dataset_id, annotation_id, feature, page_key, source_resp) VALUES
{",\n".join(feature_results_rows)};
\n\n"""
            )

        if feature_values_rows:
            out.write(
                f"""INSERT INTO feature_result_values (dataset_id, annotation_id, feature, page_key, value_index, surface, properties_json) VALUES
{",\n".join(feature_values_rows)};
"""
            )


if __name__ == "__main__":
    csv_path = Path("/Users/mia/dev/personal/elements-dh/ocrflow/store/items_metadata/title_page.csv")
    output_path = Path("/ocrflow/internal/migrations/ocrflow/1772144022_feature_result_tps_seed.sql")
    generate_sql(csv_path, output_path)