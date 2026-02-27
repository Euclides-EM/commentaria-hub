import csv
import uuid
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List


DATASET_ID = "tps"
ANNOTATION_ID = "ann_1"
SOURCE_RESP = "csv_import"


def gen_random_uuid() -> str:
    return str(uuid.uuid4())


def sql_escape(s: str) -> str:
    return s.replace("'", "''")


def split_values(value: str) -> List[str]:
    if not value:
        return []
    v = value.strip()
    if not v or v.lower() in {"none", "null", "n/a"}:
        return []

    if "::" in v:
        parts = v.split("::")
    elif "," in v:
        parts = v.split(",")
    else:
        parts = [v]

    cleaned = [p.strip() for p in parts if p and p.strip()]
    return [c for c in cleaned if c.lower() not in {"none", "null", "n/a"}]


@dataclass(frozen=True)
class FeatureSeed:
    feature_id: str  # DB id
    name: str
    description: str
    prompt: str
    is_list: bool
    is_default: bool
    color: str
    properties_json: str = "[]"


# Single source of truth
FEATURES: Dict[str, FeatureSeed] = {
    "base_content": FeatureSeed(
        feature_id="base_content",
        name="Base Content",
        description="The minimal designation of the book’s main content, typically appearing at the beginning of the title page, without elaboration.",
        prompt="The minimal designation of the book's main content, typically appearing at the beginning of the title page, without elaboration.",
        is_list=False,
        is_default=True,
        color="#FADADD",
    ),
    "content_description": FeatureSeed(
        feature_id="content_description",
        name="Base Content Description",
        description="Additional elements extending beyond the base content, describing it or highlighting the book’s special features.",
        prompt="Additional elements extending beyond the base content, describing it or highlighting the book’s special features.",
        is_list=False,
        is_default=False,
        color="#AEC6CF",
    ),
    "editor_name": FeatureSeed(
        feature_id="editor_name",
        name="Adapter Attribution",
        description="The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page.",
        prompt="The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page.",
        is_list=False,
        is_default=True,
        color="#909fd7",
    ),
    "editor_description": FeatureSeed(
        feature_id="editor_description",
        name="Adapter Description",
        description="Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations.",
        prompt="Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations.",
        is_list=False,
        is_default=True,
        color="#FFDAB9",
    ),
    "dedicatee_name": FeatureSeed(
        feature_id="dedicatee_name",
        name="Patronage Dedication",
        description="Mentions of patrons.",
        prompt="Mentions of patrons or dedication.",
        is_list=False,
        is_default=True,
        color="#D4C5F9",
    ),
    "edition_details": FeatureSeed(
        feature_id="edition_details",
        name="Edition Statement",
        description="Any information that is highlighted as relevant for this specific edition.",
        prompt="Any information that is highlighted as relevant for this specific edition such as claims regarding the corrections and revisions introduced in it.",
        is_list=False,
        is_default=True,
        color="#FFC1CC",
    ),
    "printing_privilege": FeatureSeed(
        feature_id="printing_privilege",
        name="Publishing Privileges",
        description="Mentions of royal privileges or legal permissions granted for printing.",
        prompt="Mentions of royal privileges or legal permissions granted for printing.",
        is_list=False,
        is_default=True,
        color="#D1E7E0",
    ),
    "references_to_euclid": FeatureSeed(
        feature_id="references_to_euclid",
        name="Euclid References",
        description="Euclid's name as it appears on the title page.",
        prompt="Euclid's name as it appears on the title page.",
        is_list=True,
        is_default=True,
        color="#F0E68C",
    ),
    "educational_authorities_references": FeatureSeed(
        feature_id="educational_authorities_references",
        name="Other Educational Authorities",
        description="Mentions of other scholars, either ancients, such as Theon of Alexandria, or contemporary, like Simon Stevin.",
        prompt="Mentions of other scholars, either ancients, such as Theon of Alexandria, or contemporary, like Simon Stevin.",
        is_list=True,
        is_default=True,
        color="#e567ac",
    ),
    "origin_language": FeatureSeed(
        feature_id="origin_language",
        name="Explicit Language References",
        description="Mentions of the source language (e.g., Latin or Greek) and/or the target language.",
        prompt="Mentions of the source language (e.g., Latin or Greek) and/or the target language.",
        is_list=True,
        is_default=True,
        color="#e59c67",
    ),
    "description_of_euclid": FeatureSeed(
        feature_id="description_of_euclid",
        name="Euclid Description",
        description="Any descriptors found alongside the Euclid's name, such as mentioning him being a mathematician.",
        prompt="Any descriptors found alongside Euclid's name, such as mentioning him being a mathematician.",
        is_list=True,
        is_default=True,
        color="#b0e57c",
    ),
    "action_verbs": FeatureSeed(
        feature_id="action_verbs",
        name="Verbs",
        description="Action verbs such as traduit (translated), commenté (commented), augmenté (expanded) that describe the role the contemporary scholar played in bringing about the work.",
        prompt="Action verbs such as traduit (translated), commenté (commented), augmenté (expanded) that describe the role the contemporary scholar played in bringing about the work.",
        is_list=True,
        is_default=True,
        color="#954caf",
    ),
    "audience": FeatureSeed(
        feature_id="audience",
        name="Intended Audience",
        description="Explicit mentions of the work's intended recipients or audience.",
        prompt="Explicit mentions of the work's intended recipients or audience.",
        is_list=True,
        is_default=True,
        color="#E4A0D8",
    ),
    "elements_designation": FeatureSeed(
        feature_id="elements_designation",
        name="Elements Designation",
        description="The designation of the Elements, such as 'Elements of Geometry' or 'Euclid’s Elements', as it appears on the title page.",
        prompt="The designation of the Elements, such as 'Elements of Geometry' or 'Euclid’s Elements', as it appears on the title page.",
        is_list=False,
        is_default=False,
        color="#A3D5C3",
    ),
    "greek_text": FeatureSeed(
        feature_id="greek_text",
        name="Greek designation",
        description="Greek designation of the book in non-Greek books.",
        prompt="Greek designation of the book in non-Greek books.",
        is_list=False,
        is_default=False,
        color="#F0B2A1",
    ),
    "institutions": FeatureSeed(
        feature_id="institutions",
        name="Institutions",
        description="Mentions of institutions, such as societies or universities, associated with the book.",
        prompt="Mentions of institutions, such as societies or universities, associated with the book.",
        is_list=True,
        is_default=True,
        color="#B0C4DE",
    ),
    "bound_with": FeatureSeed(
        feature_id="bound_with",
        name="Bound With",
        description="Mentions of other works included in the book, such as 'Optics'.",
        prompt=(
            "Mentions of other works that are included in the work, in addition to Euclid's Elements, such as 'Optics', 'Data', theorems by Archimedes. "
            "Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. "
            "Try to mark the minimal unit of the bound work and break down the bound works into their components when possible."
        ),
        is_list=True,
        is_default=True,
        color="#FFB6C1",
    ),
    "enriched_with": FeatureSeed(
        feature_id="enriched_with",
        name="Enriched With",
        description="Mentions of additional content that is not part of the core text that enriches the text and makes it more understandable, accurate, or useful.",
        prompt=(
            "Mentions of additional content that is not part of the core text, such as illustrations, diagrams, explanations, expositions, examples or other supplementary material that enriches the text. "
            "Try to mark the minimal unit of enrichment and break down the enrichment into components when possible."
        ),
        is_list=True,
        is_default=True,
        color="#D3D3D3",
    ),
    "date_in_imprint": FeatureSeed(
        feature_id="date_in_imprint",
        name="Date in Imprint",
        description="The date of publication as it appears on the title page, typically in the form of a year.",
        prompt="Mentions of the date, usually in the form of a year, when the book was printed or published.",
        is_list=False,
        is_default=False,
        color="#FFDEAD",
    ),
    "publisher_in_imprint": FeatureSeed(
        feature_id="publisher_in_imprint",
        name="Publisher in Imprint",
        description="The name of the publisher or printer as it appears on the title page.",
        prompt="Mentions of the publisher or printer. Try to include the minimal unit of the publisher's name.",
        is_list=True,
        is_default=False,
        color="#ADD8E6",
    ),
    "location_in_imprint": FeatureSeed(
        feature_id="location_in_imprint",
        name="Place in Imprint",
        description="The place of publication as it appears on the title page, typically a city.",
        prompt="Mentions of the city or town where the book was printed or published. Do not include full addresses, just the city or town name.",
        is_list=True,
        is_default=False,
        color="#E6E6FA",
    ),
    "printing_privilege_in_imprint": FeatureSeed(
        feature_id="printing_privilege_in_imprint",
        name="Privileges in Imprint",
        description="Mentions of royal privileges or legal permissions granted for printing, such as 'by royal permission' or 'with the approval of the censor'.",
        prompt="Mentions of royal privileges or legal permissions granted for printing.",
        is_list=True,
        is_default=False,
        color="#D1E7E0",
    ),
    "dedication_in_imprint": FeatureSeed(
        feature_id="dedication_in_imprint",
        name="Dedication in Imprint",
        description="Mentions of dedications to patrons or other individuals, typically found on the title page or in the preface.",
        prompt="Mentions of patrons or dedications.",
        is_list=True,
        is_default=False,
        color="#D4C5F9",
    ),
    "editor_in_imprint": FeatureSeed(
        feature_id="editor_in_imprint",
        name="Adapter Attribution in Imprint",
        description="The name of the author as it appears on the title page, typically in the form of 'by [Author Name]'.",
        prompt=(
            "The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. "
            "Do not include descriptors and do not include printer or publisher."
        ),
        is_list=False,
        is_default=False,
        color="#909fd7",
    ),
    "editor_description_in_imprint": FeatureSeed(
        feature_id="editor_description_in_imprint",
        name="Adapter Description in Imprint",
        description="Any descriptors found alongside the author name, such as academic titles, ranks, or affiliations.",
        prompt=(
            "Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations. "
            "Do not include printer or publisher."
        ),
        is_list=True,
        is_default=False,
        color="#FFDAB9",
    ),
    # Kept in import set, but no UI mapping given → give a sensible seed
    "destination_language": FeatureSeed(
        feature_id="destination_language",
        name="Destination Language",
        description="Mentions of the target language of the edition or translation (e.g., “en François”, “in English”).",
        prompt="Mentions of the target language of the edition or translation (e.g., “en François”, “in English”).",
        is_list=True,
        is_default=True,
        color="#e59c67",
    ),
}

FEATURE_IDS_TO_IMPORT = set(FEATURES.keys())


def generate_sql(csv_path: Path, output_path: Path) -> None:
    # Deterministic: one revision per feature, reused by all results of that feature
    revision_by_feature = {fid: gen_random_uuid() for fid in FEATURES.keys()}

    feature_rows: List[str] = []
    revision_rows: List[str] = []
    result_rows: List[str] = []
    value_rows: List[str] = []

    # Seed features
    for fid in sorted(FEATURES.keys()):
        meta = FEATURES[fid]
        feature_rows.append(
            "("
            f"'{sql_escape(meta.feature_id)}', "
            f"'{sql_escape(meta.name)}', "
            f"'{sql_escape(meta.description)}', "
            f"'{sql_escape(DATASET_ID)}', "
            f"{1 if meta.is_default else 0}, "
            f"{1 if meta.is_list else 0}, "
            f"'{sql_escape(meta.color)}', "
            f"'{sql_escape(meta.properties_json)}'"
            ")"
        )

        rev_id = revision_by_feature[fid]
        revision_rows.append(
            "("
            f"'{sql_escape(rev_id)}', "
            f"'v1', "
            f"'Initial seeded revision', "
            f"'{sql_escape(DATASET_ID)}', "
            f"'{sql_escape(meta.feature_id)}', "
            f"'{sql_escape(meta.prompt)}', "
            f"'{sql_escape(SOURCE_RESP)}'"
            ")"
        )

    # Seed results + values from CSV
    with open(csv_path, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)

        for row in reader:
            page_key = (row.get("key") or "").strip()
            if not page_key:
                continue

            for feature_id, raw in row.items():
                if feature_id not in FEATURE_IDS_TO_IMPORT:
                    continue

                vals = split_values(raw or "")
                if not vals:
                    continue

                meta = FEATURES[feature_id]
                res_id = gen_random_uuid()
                rev_id = revision_by_feature[feature_id]

                result_rows.append(
                    "("
                    f"'{sql_escape(res_id)}', "
                    f"'{sql_escape(meta.name)}', "
                    f"'' , "
                    f"'{sql_escape(DATASET_ID)}', "
                    f"'{sql_escape(meta.feature_id)}', "
                    f"'{sql_escape(ANNOTATION_ID)}', "
                    f"'{sql_escape(page_key)}', "
                    f"'{sql_escape(SOURCE_RESP)}', "
                    f"NULL, "
                    f"'{sql_escape(rev_id)}', "
                    f"'v1'"
                    ")"
                )

                for v in vals:
                    value_rows.append(
                        "("
                        f"'{sql_escape(res_id)}', "
                        f"'{sql_escape(v)}', "
                        f"'{{}}'"
                        ")"
                    )

    with open(output_path, "w", encoding="utf-8") as out:
        out.write("-- Seeded by csv_import script\n")
        out.write("PRAGMA foreign_keys = ON;\n\n")

        if feature_rows:
            out.write(
                "INSERT INTO features (id, name, description, dataset_id, is_default, is_list, color, properties)\n"
                "VALUES\n"
                + ",\n".join(feature_rows)
                + ";\n\n"
            )

        if revision_rows:
            out.write(
                "INSERT INTO feature_revisions (id, name, description, dataset_id, feature_id, prompt, categorizer)\n"
                "VALUES\n"
                + ",\n".join(revision_rows)
                + ";\n\n"
            )

        if result_rows:
            out.write(
                "INSERT INTO feature_results (id, name, description, dataset_id, feature_id, annotation_id, page_key, source_resp, source_id, source_revision, source_name)\n"
                "VALUES\n"
                + ",\n".join(result_rows)
                + ";\n\n"
            )

        if value_rows:
            out.write(
                "INSERT INTO result_values (result_id, surface, properties)\n"
                "VALUES\n"
                + ",\n".join(value_rows)
                + ";\n"
            )


if __name__ == "__main__":
    csv_path = Path(
        "/Users/mia/dev/personal/elements-dh/ocrflow/store/items_metadata/title_page.csv"
    )
    output_path = Path(
        "/ocrflow/internal/migrations/ocrflow/1772144022_feature_result_tps_seed.sql"
    )
    generate_sql(csv_path, output_path)