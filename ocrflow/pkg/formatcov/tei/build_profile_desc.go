package tei

import (
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei/model"
)

// entityProfile holds aggregated profile data for one entity (surface, normalized value, feature name).
type entityProfile struct {
	EntID       string
	Surface     string
	Value       string
	FeatureName string
}

func buildProfileDesc(entities []EntityItem) *model.ProfileDesc {
	// Group by normalized entity ID; collect surface, value, feature_name (one feature per entity for display)
	byEnt := make(map[string]*entityProfile)
	for _, it := range entities {
		if strings.TrimSpace(it.Ref) == "" {
			continue
		}
		entID := normalizedEntityID(it.Ref)
		if entID == "" {
			continue
		}
		if byEnt[entID] == nil {
			byEnt[entID] = &entityProfile{EntID: entID}
		}
		switch strings.TrimSpace(it.Type) {
		case "feature_name":
			byEnt[entID].FeatureName = strings.TrimSpace(it.Value)
		case "surface":
			byEnt[entID].Surface = strings.TrimSpace(it.Value)
		case "value":
			byEnt[entID].Value = strings.TrimSpace(it.Value)
		}
	}
	if len(byEnt) == 0 {
		return nil
	}

	var entIDs []string
	for id := range byEnt {
		entIDs = append(entIDs, id)
	}
	sort.Strings(entIDs)

	var items []model.Item
	for _, id := range entIDs {
		p := byEnt[id]
		if p.FeatureName == "" && p.Surface == "" && p.Value == "" {
			continue
		}
		label := strings.ReplaceAll(strings.ReplaceAll(p.Surface, "\n", " "), "\r", " ")
		if label == "" {
			label = strings.ReplaceAll(strings.ReplaceAll(p.Value, "\n", " "), "\r", " ")
		}
		label = strings.TrimSpace(label)
		var notes []model.Note
		if p.FeatureName != "" {
			featID := featureNameToCategoryID(p.FeatureName)
			notes = append(notes, model.Note{Type: "feature", Ana: "#" + featID, Text: p.FeatureName})
		}
		notes = append(notes, model.Note{Type: "normalized", Text: p.Value})
		items = append(items, model.Item{
			XmlID: id,
			Label: label,
			Notes: notes,
		})
	}
	pd := &model.ProfileDesc{}
	if len(items) > 0 {
		if pd.ParticDesc == nil {
			pd.ParticDesc = &model.ParticDesc{}
		}
		pd.ParticDesc.List = &model.List{Items: items}
	}
	if pd.ParticDesc == nil {
		return nil
	}
	return pd
}
