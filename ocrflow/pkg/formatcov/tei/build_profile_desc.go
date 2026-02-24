package tei

import (
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei/model"
)

func buildProfileDesc(entities *EntitiesInput) *model.ProfileDesc {
	var p map[string][]Profile
	if entities != nil {
		p = entities.Profiles
	} else {
		p = make(map[string][]Profile)
	}
	terms := make([]model.Term, 0)

	// stable output: sort by entity id then by type/value
	entityIDs := make([]string, 0, len(p))
	for id := range p {
		entityIDs = append(entityIDs, id)
	}
	sort.Strings(entityIDs)

	for _, entID := range entityIDs {
		profs := append([]Profile(nil), p[entID]...)
		sort.Slice(profs, func(i, j int) bool {
			if profs[i].Type != profs[j].Type {
				return profs[i].Type < profs[j].Type
			}
			return profs[i].Value < profs[j].Value
		})

		corresp := "#" + entID
		for _, pr := range profs {
			if strings.TrimSpace(pr.Value) == "" {
				continue
			}
			terms = append(terms, model.Term{
				Type:    pr.Type,
				Corresp: corresp,
				Text:    pr.Value,
			})
		}
	}

	return &model.ProfileDesc{
		TextClass: &model.TextClass{
			Keywords: &model.Keywords{
				Scheme: "entity-profiles",
				Terms:  terms,
			},
		},
	}
}
