import { Item } from "../types";
import { NO_AUTHOR, NO_CITY, NO_YEAR } from "../constants";

interface Author {
  name: string;
  lastName: string;
  firstName: string;
}

function parseAuthorName(authorString: string): Author {
  const trimmed = authorString.trim();

  if (trimmed.includes(",")) {
    const [lastName, firstName] = trimmed.split(",").map((s) => s.trim());
    return {
      name: trimmed,
      lastName: lastName || "",
      firstName: firstName || "",
    };
  }

  const parts = trimmed.split(" ").filter((p) => p);
  if (parts.length === 0) {
    return { name: "", lastName: "", firstName: "" };
  }

  const lastName = parts[parts.length - 1];
  const firstName = parts.slice(0, -1).join(" ");

  return {
    name: trimmed,
    lastName,
    firstName,
  };
}

function formatAuthorForCitation(
  authors: string[],
  isMultiEntry: boolean = false,
): string {
  if (!authors || authors.length === 0) {
    return NO_AUTHOR;
  }

  if (isMultiEntry) {
    return "———";
  }

  if (authors.length === 1) {
    const author = parseAuthorName(authors[0]);
    if (author.name.endsWith(" (?)")) {
      const cleanName = author.name.slice(0, -4).trim();
      const cleanAuthor = parseAuthorName(cleanName);
      return cleanAuthor.lastName && cleanAuthor.firstName
        ? `[${cleanAuthor.lastName}, ${cleanAuthor.firstName}?]`
        : `[${cleanName}?]`;
    }
    return author.lastName && author.firstName
      ? `${author.lastName}, ${author.firstName}`
      : author.name || NO_AUTHOR;
  }

  if (authors.length === 2) {
    const firstAuthor = parseAuthorName(authors[0]);
    const secondAuthor = parseAuthorName(authors[1]);

    const firstFormatted =
      firstAuthor.lastName && firstAuthor.firstName
        ? `${firstAuthor.lastName}, ${firstAuthor.firstName}`
        : firstAuthor.name;

    const secondFormatted = secondAuthor.name;

    return `${firstFormatted}, and ${secondFormatted}`;
  }

  if (authors.length <= 6) {
    const firstAuthor = parseAuthorName(authors[0]);
    const firstFormatted =
      firstAuthor.lastName && firstAuthor.firstName
        ? `${firstAuthor.lastName}, ${firstAuthor.firstName}`
        : firstAuthor.name;

    const otherAuthors = authors
      .slice(1, -1)
      .map((a) => parseAuthorName(a).name);
    const lastAuthor = parseAuthorName(authors[authors.length - 1]).name;

    return `${firstFormatted}, ${otherAuthors.join(", ")}${otherAuthors.length > 0 ? "," : ""} and ${lastAuthor}`;
  }

  const firstAuthor = parseAuthorName(authors[0]);
  const firstFormatted =
    firstAuthor.lastName && firstAuthor.firstName
      ? `${firstAuthor.lastName}, ${firstAuthor.firstName}`
      : firstAuthor.name;

  const secondAuthor = parseAuthorName(authors[1]).name;
  const thirdAuthor = parseAuthorName(authors[2]).name;

  return `${firstFormatted}, ${secondAuthor}, ${thirdAuthor}, et al.`;
}

function formatTitle(title: string | null): string {
  if (!title) return "[No Title]";

  const formatted = title
    .replace(/[\r\n]+/g, " ")
    .replaceAll(/-\s+/gi, "")
    .replaceAll(/\[vol\. 1]:?\s*/gi, "")
    .replaceAll(/\[general title page]:?\s*/gi, "")
    .replace(/\.\s*$/g, "")
    .replace(/\s+/g, " ")
    .trim();

  if (formatted === "?" || formatted === "") {
    return "[No Title]";
  }

  const words = formatted.split(" ");
  const titleCased = words.map((word, index) => {
    const lower = word.toLowerCase();
    if (index === 0 || index === words.length - 1) {
      return word.charAt(0).toUpperCase() + word.slice(1).toLowerCase();
    }

    const articlesPrepositions = [
      "a",
      "an",
      "the",
      "and",
      "but",
      "or",
      "for",
      "nor",
      "on",
      "at",
      "to",
      "from",
      "by",
      "of",
      "in",
      "with",
      "as",
    ];

    if (articlesPrepositions.includes(lower)) {
      return lower;
    }

    return word.charAt(0).toUpperCase() + word.slice(1).toLowerCase();
  });

  return titleCased.join(" ");
}

function formatPublishers(cities: string[], publishers: string[]): string {
  if (
    (!cities || cities.length === 0) &&
    (!publishers || publishers.length === 0)
  ) {
    return `${NO_CITY}: [Publisher Unknown]`;
  }

  const cityList = cities && cities.length > 0 ? cities : [NO_CITY];
  const publisherList =
    publishers && publishers.length > 0 ? publishers : ["[Publisher Unknown]"];

  if (cityList.length === 1 && publisherList.length === 1) {
    return `${cityList[0]}: ${publisherList[0]}`;
  }

  const maxLength = Math.max(cityList.length, publisherList.length);
  const publisherPairs: string[] = [];

  for (let i = 0; i < maxLength; i++) {
    const city = cityList[i] || cityList[cityList.length - 1];
    const publisher =
      publisherList[i] || publisherList[publisherList.length - 1];
    publisherPairs.push(`${city}: ${publisher}`);
  }

  return publisherPairs.join("; ");
}

function formatYear(year: string | null): string {
  return year || NO_YEAR;
}

function formatVolumes(volumesCount: number | null): string {
  if (!volumesCount || volumesCount <= 1) {
    return "";
  }
  return `${volumesCount} vols. `;
}

function formatCitationEntry(
  item: Item,
  isMultiEntry: boolean = false,
): string {
  const authors = formatAuthorForCitation(item.authors, isMultiEntry);
  const title = formatTitle(item.title || item.shortTitle);
  const volumes = formatVolumes(item.volumesCount);
  const publishers = formatPublishers(item.cities, item.publishers);
  const year = formatYear(item.year);

  return `${authors}. ${title}. ${volumes}${publishers}, ${year}.`;
}

interface CitationGroup {
  key: string;
  items: Item[];
  primaryAuthor: string;
  sortKey: string;
}

function getAuthorSortKey(authors: string[]): string {
  if (!authors || authors.length === 0) {
    return "zzz" + NO_AUTHOR;
  }

  const firstAuthorName = authors[0].endsWith(" (?)")
    ? authors[0].slice(0, -4).trim()
    : authors[0];
  const firstAuthor = parseAuthorName(firstAuthorName);
  const lastName = firstAuthor.lastName.toLowerCase();
  const firstName = firstAuthor.firstName.toLowerCase();

  const authorCount = authors.length.toString().padStart(3, "0");

  const coauthors = authors
    .slice(1)
    .map((a) => {
      const cleanName = a.endsWith(" (?)") ? a.slice(0, -4).trim() : a;
      return parseAuthorName(cleanName).lastName.toLowerCase();
    })
    .join("_");

  return `${lastName}_${firstName}_${authorCount}_${coauthors}`;
}

function groupItemsByAuthor(items: Item[]): CitationGroup[] {
  const groups: Map<string, CitationGroup> = new Map();

  items.forEach((item) => {
    const authorsKey = JSON.stringify(item.authors || []);

    if (!groups.has(authorsKey)) {
      groups.set(authorsKey, {
        key: authorsKey,
        items: [],
        primaryAuthor: item.authors?.[0] || NO_AUTHOR,
        sortKey: getAuthorSortKey(item.authors),
      });
    }

    groups.get(authorsKey)!.items.push(item);
  });

  groups.forEach((group) => {
    group.items.sort((a, b) => {
      const yearA = a.year || "";
      const yearB = b.year || "";

      if (yearA !== yearB) {
        return yearA.localeCompare(yearB);
      }

      const titleA = formatTitle(a.title || a.shortTitle);
      const titleB = formatTitle(b.title || b.shortTitle);
      return titleA.localeCompare(titleB);
    });
  });

  return Array.from(groups.values()).sort((a, b) =>
    a.sortKey.localeCompare(b.sortKey),
  );
}

function assignYearSuffixes(items: Item[]): Item[] {
  const yearGroups: Map<string, Item[]> = new Map();

  items.forEach((item) => {
    const year = item.year || "n.d.";
    if (!yearGroups.has(year)) {
      yearGroups.set(year, []);
    }
    yearGroups.get(year)!.push(item);
  });

  const modifiedItems: Item[] = [];

  yearGroups.forEach((group) => {
    if (group.length === 1) {
      modifiedItems.push(group[0]);
    } else {
      group.forEach((item, index) => {
        const suffix = String.fromCharCode(97 + index);
        modifiedItems.push({
          ...item,
          year: item.year ? `${item.year}${suffix}` : `n.d.${suffix}`,
        });
      });
    }
  });

  return modifiedItems;
}

export function generateChicagoCitations(items: Item[]): string {
  const citations: string[] = [];
  const groups = groupItemsByAuthor(items);

  groups.forEach((group) => {
    const itemsWithSuffixes = assignYearSuffixes(group.items);

    itemsWithSuffixes.forEach((item, index) => {
      const isMultiEntry = index > 0;
      const citation = formatCitationEntry(item, isMultiEntry);
      citations.push(citation);
    });
  });

  return citations.join("\n\n");
}

export function exportCitationsAsRTF(
  items: Item[],
  filename: string = "citations.rtf",
): void {
  const citations = generateChicagoCitations(items);

  const rtfHeader =
    "{\\rtf1\\ansi\\deff0 {\\fonttbl{\\f0 Times New Roman;}}\\f0\\fs24";
  const rtfFooter = "}";

  const citationLines = citations.split("\n\n");
  const rtfContent = citationLines
    .map((line) => {
      const escapedLine = Array.from(line)
        .map((char) => {
          const code = char.charCodeAt(0);
          if (code > 127) {
            return `\\u${code}?`;
          }
          switch (char) {
            case "\\":
              return "\\\\";
            case "{":
              return "\\{";
            case "}":
              return "\\}";
            default:
              return char;
          }
        })
        .join("");

      return `\\li720\\fi-720 ${escapedLine}\\par\\par`;
    })
    .join("");

  const rtfDocument = `${rtfHeader}\n${rtfContent}${rtfFooter}`;

  const blob = new Blob([rtfDocument], { type: "application/rtf" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
