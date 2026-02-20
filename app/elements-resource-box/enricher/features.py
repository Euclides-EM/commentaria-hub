from tpparts import TitlePagePart


class Feature:
    def __init__(self, name: str, description: str, is_list: bool = False, title_page_part : TitlePagePart = TitlePagePart.TITLE_PAGE):
        self.name = name
        self.description = description
        self.is_list = is_list
        self.title_page_part = title_page_part

    def __repr__(self):
        return f"Feature(name={self.name}, description={self.description}, is_list={self.is_list}), title_page_part={self.title_page_part})"

BASE_CONTENT = Feature(
    "base_content",
    "The minimal designation of the book's main content, typically appearing at the beginning of the title page, without elaboration."
)
ADAPTER_ATTRIBUTION = Feature(
    "editor_name",
    "The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page."
)
ADAPTER_DESCRIPTION = Feature(
    "editor_description",
    "Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations."
)
ADAPTER_DESCRIPTION2 = Feature(
    "AUTHOR DESCRIPTION 2",
    "Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations. (an additional quote if applicable)."
)
PATRONAGE_DEDICATION = Feature(
    "dedicatee_name",
    "Mentions of patrons or dedication."
)
EDITION_STATEMENT = Feature(
    "edition_details",
    "Any information that is highlighted as relevant for this specific edition such as claims regarding the corrections and revisions introduced in it."
)
PUBLISHING_PRIVILEGES = Feature(
    "printing_privilege",
    "Mentions of royal privileges or legal permissions granted for printing."
)
VERBS = Feature(
    "action_verbs",
    "Action verbs such as traduit (translated), commenté (commented), augmenté (expanded) that describe the role the contemporary scholar played in bringing about the work.",
    is_list=True
)
EXPLICIT_LANGUAGE_REFERENCES = Feature(
    "origin_language",
    "Mentions of the source language (e.g., Latin or Greek) and/or the target language.",
    is_list=True
)
REFERENCES_TO_OTHER_EDUCATIONAL_AUTHORITIES = Feature(
    "educational_authorities_references",
    "Mentions of other scholars, either ancients, such as Theon of Alexandria, or contemporary, like Simon Stevin.",
    is_list=True
)
EUCLID_MENTIONS = Feature(
    "references_to_Euclid",
    "Euclid's name as it appears on the title page.",
    is_list=True
)
EUCLID_DESCRIPTION = Feature(
    "description_of_Euclid",
    "Any descriptors found alongside Euclid's name, such as mentioning him being a mathematician.",
    is_list=True
)
EUCLID_DESCRIPTION2 = Feature(
    "EUCLID DESCRIPTION 2",
    "Any descriptors found alongside Euclid’s name, such as mentioning him being a mathematician. (an additional quote if applicable).",
    is_list=True
)
INTENDED_AUDIENCE = Feature(
    "audience",
    "Explicit mentions of the work's intended recipients or audience.",
    is_list=True
)
INTENDED_AUDIENCE2 = Feature(
    "RECIPIENT 2",
    "Explicit mentions of the work's intended recipients or audience. (an additional quote if applicable).",
    is_list=True
)
ELEMENTS_DESIGNATION = Feature(
    "Elements_designation",
    "The designation of the Elements, such as 'Elements of Geometry' or 'Euclid's Elements', as it appears on the title page."
)
GREEK_DESIGNATION = Feature(
    "Greek_text",
    "Greek designation of the book in non-Greek books."
)
INSTITUTIONS = Feature(
    "institutions",
    "Mentions of institutions, such as societies or universities, associated with the book.",
    is_list=True
)
BOUND_WITH = Feature(
    "bound_with",
    "Mentions of other works that are included in the work, in addition to Euclid's Elements, such as 'Optics', 'Data', theorems by Archimedes. "
    "Mentions of additions ingrained in the core text and written by the adapter/translator of the text, such as examples or explanations, should not be included here. "
    "Try to mark the minimal unit of the bound work. For example, instead of marking 'cum expositione Theonis.', mark 'expositione Theonis' as the bound work. "
    "Try to also break down the bound works into their components. For example, instead of marking 'Phaenomena, Optics, and Catoptrics', mark 'Phaenomena', 'Optics', and 'Catoptrics' as the bound works.",
    is_list=True
)
ENRICHED_WITH = Feature(
    "enriched_with",
    "Mentions of additional content that is not part of the core text, such as illustrations, diagrams, explanations, expositions, examples or other supplementary material that enriches the text and makes it more understandable, accurate, or useful. "
    "Mentions of other distinct works that are included in the book, in addition to Euclid's Elements, such as 'Optics', 'Data', theorems by Archimedes, should not be included here."
    "Try to mark the minimal unit of enrichment. For example, instead of marking 'To which are added some utilities', mark 'utilities' as the enriched content. "
    "Try to also break down the enrichment into its components, such as 'explanations', 'examples', 'diagrams', etc. ",
    is_list=True
)
IMPRINT_DATE = Feature(
    "date_in_imprint",
    "Mentions of the date, usually in the form of a year, when the book was printed or published.",
    title_page_part=TitlePagePart.IMPRINT
)
IMPRINT_PUBLISHER = Feature(
    "publisher_in_imprint",
    "Mentions of the publisher or printer, such as the name of the printing house or the person responsible for the publication."
    "Try to include the minimal unit of the publisher's name. For example, instead of marking 'Apud Vincentium Accoltum', mark 'Vincentium Accoltum', "
    "and instead of marking 'Chez GVILLAVME AVVRAY, au haut de la ruë sainct Iean de Beauuais, au Bellerophon couronné', mark 'GVILLAVME AVVRAY'.",
    title_page_part=TitlePagePart.IMPRINT,
    is_list=True
)
IMPRINT_CITY = Feature(
    "location_in_imprint",
    "Mentions of the city or town where the book was printed or published. "
    "Do not include full addresses, just the city or town name. "
    "For example, instead of marking 'A Paris, chez GVILLAVME AVVRAY, au haut de la ruë sainct Iean de Beauuais, au Bellerophon couronné', mark 'Paris'.",
    title_page_part=TitlePagePart.IMPRINT,
    is_list=True
)
IMPRINT_PRIVILEGES = Feature(
    "printing_privilege_in_imprint",
    "Mentions of royal privileges or legal permissions granted for printing.",
    title_page_part=TitlePagePart.IMPRINT,
    is_list=True
)
IMPRINT_DEDICATION = Feature(
    "dedication_in_imprint",
    "Mentions of patrons or dedications.",
    title_page_part=TitlePagePart.IMPRINT,
    is_list=True
)
IMPRINT_ADAPTER_ATTRIBUTION = Feature(
    "editor_in_imprint",
    "The name of the contemporary adapter (author, editor, translator, commentator, etc.) as it appears on the title page. "
    "Do not include any descriptors, just the name itself. "
    "Do not include the printer or publisher's name, just the adapter's name, if exists.",
    title_page_part=TitlePagePart.IMPRINT,
)
IMPRINT_ADAPTER_DESCRIPTION = Feature(
    "editor_description_in_imprint",
    "Any descriptors found alongside the adapter name, such as academic titles, ranks, or affiliations. "
    "Do not include the printer or publisher's name, just the adapter's description, if exists.",
    is_list=True,
    title_page_part=TitlePagePart.IMPRINT,
)

def prompt(features : list[Feature], title_page_part : TitlePagePart = TitlePagePart.TITLE_PAGE) -> str:
    output_formats = []
    definitions = []
    for feature in features:
        definitions.append(f'- {feature.name}: {feature.description}')
        if feature.is_list:
            output_formats.append(f'  "{feature.name}": [...], // zero or more quotes')
        else:
            output_formats.append(f'  "{feature.name}": "...", // a single quote or empty if not applicable')
    output_format = ",\n".join(output_formats)
    definitions_str = "\n".join(definitions)
    return """
You are an AI agent designed to extract structured metadata from historical title pages of translations of Euclid’s Elements.

You will be given:
- The transcribed text of a """ + "title page's imprint" if title_page_part == TitlePagePart.IMPRINT  else "title page" + """".
- The language of the transcription.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
""" + output_format + """
}

Definitions: 
""" + definitions_str + """
"""