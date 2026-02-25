package tei

import "strings"

// ensureHash normalizes an entity ref to # form (e.g. "ent_john" -> "#ent_john").
func ensureHash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "#") {
		return "#" + s
	}
	return s
}

// normalizedEntityID returns an xml:id-safe entity ID (e.g. "#ent_de latin" -> "ent_de_latin").
// Use for ref attributes and profile particDesc so IDs are consistent and valid.
func normalizedEntityID(ref string) string {
	s := ensureHash(ref)
	if s == "" {
		return ""
	}
	return sanitizeID(strings.TrimPrefix(s, "#"))
}

// FeatureCategoryID converts a feature display name to a taxonomy category xml:id (e.g. "Origin Language" -> "feat_origin_language").
// Use when setting @ana on mention spans and relations so they point at the encodingDesc taxonomy.
func FeatureCategoryID(name string) string {
	return featureNameToCategoryID(name)
}

// featureNameToCategoryID converts a feature display name to a taxonomy category xml:id (e.g. "Origin Language" -> "feat_origin_language").
func featureNameToCategoryID(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return ""
	}
	return "feat_" + sanitizeID(strings.ToLower(strings.ReplaceAll(s, " ", "_")))
}

// factTypeToCategoryID converts a relation/fact type to a fact taxonomy category xml:id (e.g. "translated_from" -> "fact_translation").
func factTypeToCategoryID(relationType string) string {
	s := strings.TrimSpace(relationType)
	if s == "" {
		return ""
	}
	return "fact_" + sanitizeID(strings.ToLower(strings.ReplaceAll(s, " ", "_")))
}

func occKey(pageID, blockID, lineID string) string {
	return pageID + "||" + blockID + "||" + lineID
}

func surfaceID(pageID string) string {
	// Ensure valid xml:id
	return "page_" + sanitizeID(pageID)
}

func blockID(blockID string) string {
	return "b_" + sanitizeID(blockID)
}

func zoneID(kind, pageID, raw string) string {
	return "z_" + kind + "_" + sanitizeID(pageID) + "_" + sanitizeID(raw)
}

func sanitizeID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "x"
	}
	// conservative: keep letters, digits, underscore, dash, replace others with underscore
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}
