const FEATURE_PROPERTY_DISPLAY_NAMES: Record<string, string> = {
  normalized: 'Normalized',
  language: 'Language',
  institution: 'Institution',
  ancient_persona: 'Ancient Persona',
}

export function getFeaturePropertyDisplayName(property: string): string {
  return FEATURE_PROPERTY_DISPLAY_NAMES[property] ?? property
}

export function normalizeFeatureProperties(properties: string[]): string[] {
  return [...new Set(properties.map((property) => property.trim()))]
    .filter((property) => property.length > 0)
    .sort((left, right) => left.localeCompare(right))
}
