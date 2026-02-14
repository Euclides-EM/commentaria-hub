package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/samber/lo"
)

func TestAnnotation_ApplyRules(t *testing.T) {
	expectedBookHeads := [][]string{
		{"D. HENRION.", "AV LECTEVR."},
	}

	OrdinalNumerals := []string{
		"PREMIER",
		"SECOND",
		"TROISIESME",
		"QVATRIESME",
		"CINQVIESME",
		"SIXIESME",
		"SEPTIESME",
		"HVITIESME",
		"NEVFIESME",
		"DIXIESME",
		"VNZIESME",
		"DOVZIESME",
		"TREIZIESME",
		"QVATORZIESME",
		"QVINZIESME",
	}

	for _, OrdinalNumeral := range OrdinalNumerals {
		expectedBookHeads = append(expectedBookHeads, []string{"ELEMENT", OrdinalNumeral + "."})
	}

	fmt.Println(strings.Join(lo.Map(expectedBookHeads, func(expectedBookHead []string, _ int) string {
		return strings.Join(lo.Map(expectedBookHead, func(item string, _ int) string {
			return fmt.Sprintf("%q", item)
		}), ", ")
	}), "},\n{"))

}
