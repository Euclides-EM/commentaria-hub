package tei

import (
	"fmt"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
)

// featureToRelationName maps profile Type (feature keys) to TEI relation name attribute values.
var featureToRelationName = map[string]string{
	"educated_at":          "educatedAt",
	"taught_at":            "taughtAt",
	"language_mentioned":   "languageMentioned",
	"translated_from":      "translatedFrom",
	"translation":          "translatedFrom",
	"references_to_euclid": "referencesToEuclid",
	"action_verbs":         "actionVerb",
	"base_content":         "baseContent",
	"editor_description":   "editorDescription",
	"enriched_with":        "enrichedWith",
	"origin_language":      "originLanguage",
}

// featureTypeFromAna extracts feature type from ana (e.g. "#feat_references_to_euclid" -> "references_to_euclid").
func featureTypeFromAna(ana string) string {
	s := strings.TrimSpace(ana)
	const prefix = "#feat_"
	if !strings.HasPrefix(s, prefix) {
		return ""
	}
	return strings.TrimPrefix(s, prefix)
}

// anaToRelationName converts ana to camelCase relation name (e.g. "#feat_references_to_euclid" -> "referencesToEuclid").
func anaToRelationName(ana string) string {
	snake := featureTypeFromAna(ana)
	if snake == "" {
		return ""
	}
	if name, ok := featureToRelationName[snake]; ok {
		return name
	}
	return snakeToCamel(snake)
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		r := []rune(p)
		if len(r) > 0 && r[0] >= 'a' && r[0] <= 'z' {
			r[0] = r[0] - ('a' - 'A')
		}
		parts[i] = string(r)
	}
	out := strings.Join(parts, "")
	if out == "" {
		return ""
	}
	// lowercase first character for camelCase
	r := []rune(out)
	if r[0] >= 'A' && r[0] <= 'Z' {
		r[0] = r[0] + ('a' - 'A')
	}
	return string(r)
}

// factRow is an internal row derived from a profile entry with ObjectRef set.
type factRow struct {
	entityID           string // profile key (subject)
	feature            string // profile Type
	objectRef          string
	cert               float64
	evidenceMentionIDs []string
}

// buildStandOff populates doc.Header.StandOff from entities (Items or Profiles). Any entry with
// ObjectRef set is emitted as a relation (or interp fallback). Fact IDs are auto-generated.
func buildStandOff(doc *model.TEI, entities []EntityItem) {
	if doc == nil || entities == nil {
		return
	}

	var rows []factRow
	for i := range entities {
		it := &entities[i]
		if strings.TrimSpace(it.ObjectRef) == "" {
			continue
		}
		ref := ensureHash(it.Ref)
		if ref == "" {
			continue
		}
		rows = append(rows, factRow{
			entityID:           ref,
			feature:            it.Type,
			objectRef:          it.ObjectRef,
			cert:               it.Cert,
			evidenceMentionIDs: it.EvidenceMentionIDs,
		})
	}

	if len(rows) == 0 {
		return
	}

	// Stable sort: Feature, SubjectRef (entityID), ObjectRef
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].feature != rows[j].feature {
			return rows[i].feature < rows[j].feature
		}
		subI := ensureHash(rows[i].entityID)
		subJ := ensureHash(rows[j].entityID)
		if subI != subJ {
			return subI < subJ
		}
		objI := ensureHash(rows[i].objectRef)
		objJ := ensureHash(rows[j].objectRef)
		return objI < objJ
	})

	var relations []model.Relation
	var interpFallbacks []model.Interp

	for seq, r := range rows {
		subjectRef := "#" + normalizedEntityID(r.entityID)
		objectRef := "#" + normalizedEntityID(r.objectRef)
		name, ok := featureToRelationName[r.feature]
		if !ok {
			// Unknown feature: add to InterpGrp fallback
			id := fmt.Sprintf("fact_%d", seq+1)
			interpFallbacks = append(interpFallbacks, model.Interp{
				XmlID:   "fact_" + sanitizeID(id),
				Type:    r.feature,
				Corresp: subjectRef,
				Text:    objectRef,
			})
			continue
		}

		xmlID := fmt.Sprintf("fact_%d", seq+1)
		rel := model.Relation{
			XmlID:   sanitizeID(xmlID),
			Name:    name,
			Active:  subjectRef,
			Passive: objectRef,
			Ana:     "#" + factTypeToCategoryID(r.feature),
		}
		if r.cert > 0 {
			rel.Cert = fmt.Sprintf("%.2f", r.cert)
		}
		if len(r.evidenceMentionIDs) > 0 {
			parts := make([]string, len(r.evidenceMentionIDs))
			for i, mid := range r.evidenceMentionIDs {
				parts[i] = "#" + strings.TrimPrefix(mid, "#")
			}
			rel.Source = strings.Join(parts, " ")
		}
		relations = append(relations, rel)
	}

	if len(relations) > 0 || len(interpFallbacks) > 0 {
		if doc.Header.StandOff == nil {
			doc.Header.StandOff = &model.StandOff{}
		}
		if len(relations) > 0 {
			doc.Header.StandOff.ListRelation = &model.ListRelation{Relations: relations}
		}
		if len(interpFallbacks) > 0 {
			doc.Header.StandOff.InterpGrps = append(doc.Header.StandOff.InterpGrps, model.InterpGrp{
				Type:    "fact_fallback",
				Interps: interpFallbacks,
			})
		}
	}
}

// MentionForStandOff is one mention in document order for building standOff spanGrp and relations.
type MentionForStandOff struct {
	MentionID string // e.g. "m_1"
	Ref       string // entity ref e.g. "#ent_d_evclide"
	Ana       string // e.g. "#feat_references_to_euclid"
}

// buildStandOffMentions populates doc.Header.StandOff with a spanGrp (ner-mentions) and listRelation
// (feature-assignment facts). Each mention gets a span (from/to anchors) and a relation (passive=entity, corresp=mention).
func buildStandOffMentions(doc *model.TEI, mentions []MentionForStandOff) {
	if doc == nil || len(mentions) == 0 {
		return
	}
	if doc.Header.StandOff == nil {
		doc.Header.StandOff = &model.StandOff{}
	}

	var spans []model.Span
	var relations []model.Relation
	for i, m := range mentions {
		mid := strings.TrimPrefix(m.MentionID, "#")
		anchorPart := strings.ReplaceAll(mid, "m_", "m") // m_1 -> m1 for anchor ids a_m1_s, a_m1_e
		anchorStart := "#a_" + anchorPart + "_s"
		anchorEnd := "#a_" + anchorPart + "_e"
		ref := ensureHash(m.Ref)
		if ref == "" {
			ref = "#" + normalizedEntityID(m.Ref)
		}
		// Every span must carry @ana (feature category) so mentions are self-describing and queryable by feature.
		spanAna := ensureHash(strings.TrimSpace(m.Ana))
		if spanAna == "#" {
			spanAna = ""
		}
		spans = append(spans, model.Span{
			XmlID: mid,
			From:  anchorStart,
			To:    anchorEnd,
			Ana:   spanAna,
			Ref:   ref,
		})
		// Predicate-specific @name (e.g. originLanguage, referencesToEuclid) so facts are distinguishable.
		name := anaToRelationName(m.Ana)
		if name == "" && strings.TrimSpace(m.Ana) != "" {
			name = snakeToCamel(featureTypeFromAna(m.Ana))
		}
		if name == "" {
			name = "featureAssignment"
		}
		rel := model.Relation{
			XmlID:   fmt.Sprintf("f_%d", i+1),
			Name:    name,
			Passive: ref,
			Corresp: "#" + mid,
		}
		// Encode the predicate: use feature category on the relation so the fact is clearly typed.
		if spanAna != "" {
			rel.Ana = spanAna
		} else {
			rel.Ana = "#fact_feature_assignment"
		}
		relations = append(relations, rel)
	}

	doc.Header.StandOff.SpanGrp = &model.SpanGrp{
		XmlID: "mentions",
		Type:  "ner-mentions",
		Spans: spans,
	}
	doc.Header.StandOff.ListRelation = &model.ListRelation{Relations: relations}
}
