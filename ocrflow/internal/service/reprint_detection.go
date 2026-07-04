package service

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/editdistance"
	"github.com/samber/lo"
)

type Reprint struct {
	editionSvc *Edition
}

func NewReprintService(editionSvc *Edition) *Reprint {
	return &Reprint{editionSvc: editionSvc}
}

func normalizeText(value string) string {
	return strings.Join(strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}), " ")
}

func normalizeComparableTitle(value string) []rune {
	return []rune(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, value))
}

func titlesAlmostEqual(a, b string) bool {
	left, right := normalizeComparableTitle(a), normalizeComparableTitle(b)
	if len(left) < 20 || len(right) < 20 {
		return false
	}
	maxLength := max(len(left), len(right))
	allowedEdits := max(2, maxLength/50) // At most roughly 2% variation.
	return editdistance.Runes(left, right) <= allowedEdits
}

func normalizeList(values []string) string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		if value := normalizeText(value); value != "" {
			normalized = append(normalized, value)
		}
	}
	sort.Strings(normalized)
	return strings.Join(normalized, ";")
}

func haveSharedEditor(a, b []string) bool {
	left := make(map[string]struct{}, len(a))
	for _, editor := range a {
		if editor := normalizeText(editor); editor != "" {
			left[editor] = struct{}{}
		}
	}
	for _, editor := range b {
		if _, ok := left[normalizeText(editor)]; ok {
			return true
		}
	}
	return false
}

func reprintMetadataMatches(original, suspected *model.Edition) bool {
	if original == nil || suspected == nil || original.Title == nil || suspected.Title == nil {
		return false
	}
	originalLanguages := normalizeList(original.Languages)
	return originalLanguages != "" &&
		originalLanguages == normalizeList(suspected.Languages) &&
		haveSharedEditor(original.Editor, suspected.Editor) &&
		titlesAlmostEqual(*original.Title, *suspected.Title)
}

func editionYear(ed *model.Edition) string {
	if ed == nil || ed.Year == nil {
		return ""
	}
	return strings.TrimSpace(*ed.Year)
}

// DetectReprints is read-only. Existing reprint relationships are never
// returned as candidates and are followed to their original root when found.
func (r *Reprint) DetectReprints() (*model.ReprintDetection, error) {
	editions, err := r.editionSvc.ListAllEditions()
	if err != nil {
		return nil, err
	}
	byKey := lo.KeyBy(editions, func(ed *model.Edition) string { return ed.Key })

	sort.SliceStable(editions, func(i, j int) bool {
		yi, yj := editionYear(editions[i]), editionYear(editions[j])
		if yi != yj {
			if yi == "" {
				return false
			}
			if yj == "" {
				return true
			}
			return yi < yj
		}
		return editions[i].Key < editions[j].Key
	})

	candidates := make([]model.ReprintRelationship, 0)
	for candidateIndex, candidate := range editions[1:] {
		if candidate.IsManuscript || candidate.ReprintOf != nil || editionYear(candidate) == "" {
			continue
		}
		var parent *model.Edition
		for _, possibleParent := range editions[:candidateIndex+1] {
			if !possibleParent.IsManuscript &&
				editionYear(possibleParent) != "" &&
				editionYear(possibleParent) < editionYear(candidate) &&
				reprintMetadataMatches(possibleParent, candidate) {
				parent = possibleParent
				break
			}
		}
		if parent == nil {
			continue
		}
		visitedParents := map[string]bool{}
		for parent.ReprintOf != nil {
			if visitedParents[parent.Key] {
				break
			}
			visitedParents[parent.Key] = true
			next := byKey[strings.TrimSpace(*parent.ReprintOf)]
			if next == nil || next == parent {
				break
			}
			parent = next
		}
		candidates = append(candidates, model.ReprintRelationship{
			EditionKey: candidate.Key,
			ReprintOf:  parent.Key,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		originalYearI := editionYear(byKey[candidates[i].ReprintOf])
		originalYearJ := editionYear(byKey[candidates[j].ReprintOf])
		if originalYearI != originalYearJ {
			if originalYearI == "" {
				return false
			}
			if originalYearJ == "" {
				return true
			}
			return originalYearI < originalYearJ
		}
		if candidates[i].ReprintOf != candidates[j].ReprintOf {
			return candidates[i].ReprintOf < candidates[j].ReprintOf
		}
		return candidates[i].EditionKey < candidates[j].EditionKey
	})
	return &model.ReprintDetection{Candidates: candidates}, nil
}

// ApplyReprints rechecks each edition immediately before writing, protecting
// existing relationships if the catalog changed after detection.
func (r *Reprint) ApplyReprints(ar *model.ApplyReprints, login string) (*model.ApplyReprints, error) {
	// These fields are response-only in practice. Reset them so successful
	// responses consistently encode [] rather than null (and ignore any values
	// supplied by a caller).
	ar.Updated = make([]string, 0)
	ar.Skipped = make([]string, 0)
	seen := map[string]bool{}
	for _, relationship := range ar.Relationships {
		key := strings.TrimSpace(relationship.EditionKey)
		parentKey := strings.TrimSpace(relationship.ReprintOf)
		if key == "" || parentKey == "" || key == parentKey || seen[key] {
			return nil, fmt.Errorf("invalid reprint relationship %q -> %q", key, parentKey)
		}
		seen[key] = true
		if _, err := r.editionSvc.GetEditionByID(parentKey); err != nil {
			return nil, fmt.Errorf("invalid original edition %s: %w", parentKey, err)
		}
		ed, err := r.editionSvc.GetEditionByID(key)
		if err != nil {
			return nil, err
		}
		if ed.ReprintOf != nil {
			ar.Skipped = append(ar.Skipped, key)
			continue
		}
		ed.ReprintOf = &parentKey
		if _, err := r.editionSvc.UpdateEdition(ed, login); err != nil {
			return nil, err
		}
		ar.Updated = append(ar.Updated, key)
	}
	return ar, nil
}
